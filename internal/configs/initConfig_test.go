package configs

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	// Create a temporary config file
	configContent := `
S3:
  bucketName: "test-bucket"
  region: "test-region"
  accessKeyId: "test-key"
  secretAccess: "test-secret"
Drive:
  client_id: "test-client-id"
  client_secret: "test-client-secret"
  refresh_token: "test-refresh-token"
  folder_id: "test-folder-id"
  maxConcurrent: 5
`
	tmpDir := t.TempDir()
	configPath := tmpDir
	configFile := tmpDir + "/base.yaml"
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	// Test successful init
	err = Init(configPath)
	if err != nil {
		t.Errorf("Init() failed with valid config: %v", err)
	}

	if Config.S3.BucketName != "test-bucket" {
		t.Errorf("Expected bucket name 'test-bucket', got '%s'", Config.S3.BucketName)
	}
	if Config.Drive.MaxConcurrent != 5 {
		t.Errorf("Expected maxConcurrent 5, got %d", Config.Drive.MaxConcurrent)
	}

	// Test invalid path
	err = Init("/invalid/path")
	if err == nil {
		t.Error("Init() should fail with invalid path")
	}
}
