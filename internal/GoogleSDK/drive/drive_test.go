package drive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vbauerster/mpb/v8"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// helper to create a mock drive service
func newMockDriveService(t *testing.T, handler http.HandlerFunc) (*drive.Service, *httptest.Server) {
	server := httptest.NewServer(handler)

	srv, err := drive.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
	if err != nil {
		server.Close()
		t.Fatalf("Failed to create drive service: %v", err)
	}
	return srv, server
}

func TestFindOrCreateFolder_Create(t *testing.T) {
	// Scenario: Folder not found, so it should create it.
	handler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/files") && r.Method == "GET" {
			// List files: return empty list (not found)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"files": []interface{}{},
			})
			return
		}
		if strings.Contains(r.URL.Path, "/files") && r.Method == "POST" {
			// Create file
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   "new-folder-id",
				"name": "test-folder",
			})
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	}

	srv, server := newMockDriveService(t, handler)
	defer server.Close()

	d := NewDriveManager(srv)
	// We need to unlock global mutex after test if it locks?
	// The FindOrCreateFolder uses globalFolderMutex.
	// Since tests run sequentially usually, it might be fine, but parallel tests could block.

	id := d.FindOrCreateFolder("test-folder", "root")
	if id != "new-folder-id" {
		t.Errorf("FindOrCreateFolder = %s, want new-folder-id", id)
	}
}

func TestFindOrCreateFolder_Found(t *testing.T) {
	// Scenario: Folder found.
	handler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/files") && r.Method == "GET" {
			// List files: return one file
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"files": []map[string]interface{}{
					{"id": "existing-folder-id", "name": "test-folder"},
				},
			})
			return
		}
		http.Error(w, "Unexpected request", http.StatusBadRequest)
	}

	srv, server := newMockDriveService(t, handler)
	defer server.Close()

	d := NewDriveManager(srv)

	id := d.FindOrCreateFolder("test-folder", "root")
	if id != "existing-folder-id" {
		t.Errorf("FindOrCreateFolder = %s, want existing-folder-id", id)
	}
}

func TestFileETagExistsInDrive_True(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Expect query to contain appProperties
		if !strings.Contains(r.URL.Query().Get("q"), "appProperties has { key='s3etag' and value='test-etag' }") {
			t.Errorf("Query missing ETag check: %s", r.URL.Query().Get("q"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"files": []map[string]interface{}{
				{"id": "file-id", "name": "file.txt"},
			},
		})
	}

	srv, server := newMockDriveService(t, handler)
	defer server.Close()

	d := NewDriveManager(srv)
	exists := d.FileETagExistsInDrive("test-etag", "parent-id")
	if !exists {
		t.Errorf("FileETagExistsInDrive = false, want true")
	}
}

func TestFileETagExistsInDrive_False(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"files": []interface{}{},
		})
	}

	srv, server := newMockDriveService(t, handler)
	defer server.Close()

	d := NewDriveManager(srv)
	exists := d.FileETagExistsInDrive("test-etag", "parent-id")
	if exists {
		t.Errorf("FileETagExistsInDrive = true, want false")
	}
}

func TestSyncS3PathToDrive(t *testing.T) {
	// Scenario: Path "folderA/folderB"
	// 1. List folderA under root -> Not Found -> Create folderA (id: id_A)
	// 2. List folderB under id_A -> Found (id: id_B)
	// Result: id_B

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Create request
		if r.Method == "POST" {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			name := body["name"].(string)

			if name == "folderA" {
				json.NewEncoder(w).Encode(map[string]interface{}{"id": "id_A", "name": "folderA"})
				return
			}
			t.Errorf("Unexpected create request: %s", name)
			return
		}

		// List request
		q := r.URL.Query().Get("q")
		if strings.Contains(q, "name = 'folderA'") && strings.Contains(q, "'root' in parents") {
			// folderA not found
			json.NewEncoder(w).Encode(map[string]interface{}{"files": []interface{}{}})
			return
		}
		if strings.Contains(q, "name = 'folderB'") && strings.Contains(q, "'id_A' in parents") {
			// folderB found
			json.NewEncoder(w).Encode(map[string]interface{}{
				"files": []map[string]interface{}{
					{"id": "id_B", "name": "folderB"},
				},
			})
			return
		}

		http.Error(w, fmt.Sprintf("Unexpected request: %s", q), http.StatusBadRequest)
	}

	srv, server := newMockDriveService(t, handler)
	defer server.Close()

	d := NewDriveManager(srv)
	// Mock global mutex? It is fine, tests in parallel might fail but we run sequential here mostly.
	// Actually we should mock it or ensure it doesn't block. It is a real mutex/map.

	id := d.SyncS3PathToDrive("folderA/folderB/file.txt", "root")
	if id != "id_B" {
		t.Errorf("SyncS3PathToDrive = %s, want id_B", id)
	}
}

func TestStreamUploadWithProgress(t *testing.T) {
	// 1. Mock File Server
	fileContent := "Hello World"
	fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fileContent))
	}))
	defer fileServer.Close()

	// 2. Mock Drive API
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Handle SyncS3PathToDrive logic (Mocking root/folder found immediately to simplify)
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"files": []map[string]interface{}{
					{"id": "folder_id", "name": "folder"},
				},
			})
			return
		}

		// Handle Upload
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/upload/") {
			// Read body to verify content
			// Since it is multipart, it's complex to parse fully without knowing boundary,
			// but we can check calls.
			// Actually Google API client uses Resumable upload often or Multipart.
			// For small files it might be Multipart.

			// Just return success
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "new-file-id",
				"name": "file.txt",
			})
			return
		}

		// Fallback for Create request (if not upload endpoint, checking metadata create?)
		// The API client calls POST /upload/drive/v3/files?uploadType=multipart
		// or POST /upload/drive/v3/files?uploadType=resumable
		// Our mock checks paths containing /upload/

		http.Error(w, "Unexpected request", http.StatusBadRequest)
	}

	srv, server := newMockDriveService(t, handler)
	defer server.Close()

	d := NewDriveManager(srv)

	// Create a dummy bar
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.AddBar(int64(len(fileContent)))

	err := d.StreamUploadWithProgress(fileServer.URL, "folder/file.txt", "root", "etag123", bar)
	if err != nil {
		t.Fatalf("StreamUploadWithProgress failed: %v", err)
	}
}

