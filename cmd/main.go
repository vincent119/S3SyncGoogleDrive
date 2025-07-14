package main

import (
	"S3SyncGoogleDrive/internal/GoogleSDK/drive"
	"S3SyncGoogleDrive/internal/awsSDK/s3"
	"S3SyncGoogleDrive/internal/configs"
	"S3SyncGoogleDrive/internal/googlesdk"
	progressReader "S3SyncGoogleDrive/internal/pkg/progressReader"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
)

var debug bool

func debugLog(format string, v ...any) {
	if debug {
		log.Printf(format, v...)
	}
}

func main() {
	configs.Init()
	s3.InitS3Client()
	driveService := googlesdk.GetDriveService()

	s3Prefix := flag.String("p", "", "Enter S3 prefix path (e.g.: test999)")
	driveRootID := flag.String("droot", "root", "Google Drive root folder ID")
	flag.BoolVar(&debug, "d", false, "Enable debug log")
	flag.Parse()

	if *s3Prefix == "" {
		log.Fatal("❌ Please provide S3 prefix path, e.g.: -p=test999")
	}

	prefix := *s3Prefix
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	s3Files, err := s3.ListS3Objects(configs.Config.S3.BucketName, prefix)
	if err != nil {
		log.Fatalf("❌ Failed to fetch S3 file list: %v", err)
	}
	fmt.Printf("Total S3 files fetched: %d\n", len(s3Files))

	pm := progressReader.NewProgressManager()
	semaphore := make(chan struct{}, configs.Config.Drive.MaxConcurrent)
	var wg sync.WaitGroup

	for _, obj := range s3Files {
		s3ETag := strings.Trim(*obj.ETag, "\"")
		s3Key := *obj.Key
		fileName := filepath.Base(s3Key)

		wg.Add(1)
		semaphore <- struct{}{}

		go func(s3Key, s3ETag, fileName string, size int64) {
			defer wg.Done()
			defer func() { <-semaphore }()

			presignedURL, err := s3.GetPresignedURL(configs.Config.S3.BucketName, s3Key)
			if err != nil {
				debugLog("Failed to generate presigned URL %s: %v", s3Key, err)
				return
			}

			parentID := drive.SyncS3PathToDrive(driveService, s3Key, *driveRootID)
			debugLog("Drive folder ID: %s (S3Key: %s)", parentID, s3Key)

			debugLog("Checking if ETag exists: %s", s3ETag)
			if drive.FileETagExistsInDrive(driveService, s3ETag, parentID) {
				debugLog("File already exists with the same ETag, skipping upload: %s", s3Key)
				return
			}

			bar := pm.NewBar(size, fileName)
			err = drive.StreamUploadWithProgress(driveService, presignedURL, s3Key, *driveRootID, s3ETag, bar)
			if err != nil {
				log.Printf("Failed to upload %s: %v", s3Key, err)
			}
		}(s3Key, s3ETag, fileName, *obj.Size)
	}

	wg.Wait()
	pm.Wait()
	fmt.Println("All uploads completed.")
}
