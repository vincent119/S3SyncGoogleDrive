package drive

import (
	"testing"
)

func TestDetectMimeType(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     string
	}{
		{"MP4", "video.mp4", "video/mp4"},
		{"JPEG", "image.jpeg", "image/jpeg"},
		{"JPG", "image.jpg", "image/jpeg"},
		{"PNG", "image.png", "image/png"},
		{"Unknown", "file.txt", "application/octet-stream"},
		{"NoExt", "file", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectMimeType(tt.fileName); got != tt.want {
				t.Errorf("detectMimeType(%s) = %v, want %v", tt.fileName, got, tt.want)
			}
		})
	}
}
