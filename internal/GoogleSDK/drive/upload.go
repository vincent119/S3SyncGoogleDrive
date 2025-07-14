package drive

import (
	"context"
	"fmt"
	"github.com/vbauerster/mpb/v8"
	drive "google.golang.org/api/drive/v3"
	"log"
	//"mime"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	folderCreateMutex sync.Mutex
	uploading         sync.Map
	Debug             bool
)

func debugLog(format string, v ...any) {
	if Debug {
		log.Printf(format, v...)
	}
}

func CreateFolder(srv *drive.Service, folderName, parentID string) string {
	folderMetadata := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
	}
	if parentID != "" {
		folderMetadata.Parents = []string{parentID}
	}
	folder, err := srv.Files.Create(folderMetadata).Do()
	if err != nil {
		log.Fatalf("Failed to create folder: %v", err)
	}
	debugLog("Folder created: %s (ID: %s)", folderName, folder.Id)
	return folder.Id
}

func FindOrCreateFolder(srv *drive.Service, folderName, parentID string) string {
	query := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false and '%s' in parents",
		strings.ReplaceAll(folderName, "'", "\\'"), parentID)

	for retries := 0; retries < 3; retries++ {
		resp, err := srv.Files.List().Q(query).Fields("files(id)").Do()
		if err == nil && len(resp.Files) > 0 {
			return resp.Files[0].Id
		}
		time.Sleep(time.Second * time.Duration(retries+1))
	}

	folderCreateMutex.Lock()
	defer folderCreateMutex.Unlock()

	resp, err := srv.Files.List().Q(query).Fields("files(id)").Do()
	if err == nil && len(resp.Files) > 0 {
		return resp.Files[0].Id
	}
	return CreateFolder(srv, folderName, parentID)
}

func SyncS3PathToDrive(srv *drive.Service, s3Key, rootDriveID string) string {
	parentID := rootDriveID
	for _, folder := range strings.Split(filepath.Dir(s3Key), "/") {
		if folder != "" {
			parentID = FindOrCreateFolder(srv, folder, parentID)
		}
	}
	return parentID
}

func FileETagExistsInDrive(srv *drive.Service, s3ETag, parentID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	query := fmt.Sprintf(`'%s' in parents and trashed=false and appProperties has { key='s3etag' and value='%s' }`, parentID, s3ETag)
	debugLog("ETag query: %s", query)

	resp, err := srv.Files.List().Context(ctx).Q(query).Fields("files(id, name)").Do()
	if err != nil {
		debugLog("ETag check failed, skipping file: %v", err)
		return true // Fail-safe: treat as exists to avoid duplicate uploads
	}

	if len(resp.Files) > 0 {
		debugLog("File with matching s3etag found: %s (ID: %s)", resp.Files[0].Name, resp.Files[0].Id)
		return true
	}
	debugLog("ETag check completed, no match found")
	return false
}

func StreamUploadWithProgress(srv *drive.Service, fileURL, s3Key, rootDriveID, s3ETag string, bar *mpb.Bar) error {
	if _, exists := uploading.LoadOrStore(s3Key, true); exists {
		bar.Abort(true)
		return nil
	}
	defer uploading.Delete(s3Key)

	parentFolderID := SyncS3PathToDrive(srv, s3Key, rootDriveID)

	resp, err := http.Get(fileURL)
	if err != nil {
		bar.Abort(true)
		return fmt.Errorf("Failed to download: %v", err)
	}
	defer resp.Body.Close()

	fileName := filepath.Base(s3Key)
	mimeType := detectMimeType(fileName)

	fileMetadata := &drive.File{
		Name:     fileName,
		MimeType: mimeType,
		Parents:  []string{parentFolderID},
		AppProperties: map[string]string{
			"s3etag": s3ETag,
		},
	}

	progressReader := bar.ProxyReader(resp.Body)
	defer progressReader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()

	uploadedFile, err := srv.Files.Create(fileMetadata).Context(ctx).Media(progressReader).Do()
	if err != nil {
		bar.Abort(true)
		return fmt.Errorf("Google Drive upload failed: %v", err)
	}

	log.Printf("Upload completed: %s (ID: %s)", fileName, uploadedFile.Id)
	return nil
}

func detectMimeType(fileName string) string {
	switch filepath.Ext(fileName) {
	case ".mp4":
		return "video/mp4"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}
