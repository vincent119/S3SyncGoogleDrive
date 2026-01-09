package googlesdk

import (
	"os"
	"testing"
)

// MockFileIO for testing
type MockFileIO struct {
	Files map[string][]byte
}

func NewMockFileIO() *MockFileIO {
	return &MockFileIO{
		Files: make(map[string][]byte),
	}
}

func (m *MockFileIO) ReadFile(name string) ([]byte, error) {
	if content, ok := m.Files[name]; ok {
		return content, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileIO) WriteFile(name string, data []byte, perm os.FileMode) error {
	m.Files[name] = data
	return nil
}

func TestSaveRefreshToken(t *testing.T) {
	mockFS := NewMockFileIO()
	manager := NewAuthManager(mockFS)

	token := "test-refresh-token"
	manager.saveRefreshToken(token)

	content, err := mockFS.ReadFile("config/refresh_token.txt")
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(content) != token {
		t.Errorf("Saved token = %s, want %s", string(content), token)
	}
}

func TestGetDriveService(t *testing.T) {
	mockFS := NewMockFileIO()
	mockFS.Files["config/refresh_token.txt"] = []byte("dummy-token")

	manager := NewAuthManager(mockFS)

	// This will try to create a service.
	// Since we don't mock oauth2 Config/TokenSource logic interaction with web in this layer easily
	// (it uses oauth2.Config which calls google endpoint inside TokenSource),
	// creating a service might succeed as long as it doesn't dry-run the token.
	// drive.NewService creates the client.

	srv := manager.GetDriveService()
	if srv == nil {
		t.Error("GetDriveService returned nil")
	}
	// We succeeded if no panic (log.Fatalf) happened.
	// The real token validation happens when requests are made.
}
