package s3

import (
	"context"
	"log"
	"sync"
	"time"

	"S3SyncGoogleDrive/internal/awsSDK"
	"S3SyncGoogleDrive/internal/configs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	s3Client *s3.Client
	once     sync.Once
)

// InitS3Client initializes the global S3 client (singleton)
func InitS3Client() {
	once.Do(func() {
		cfg := awsSDK.AwsConnectWithRegion(configs.Config.S3.Region)
		s3Client = s3.NewFromConfig(cfg)
		log.Println("S3 client initialized successfully")
	})
}

// GetS3Client returns the global S3 client
func GetS3Client() *s3.Client {
	if s3Client == nil {
		log.Println("S3 client is not initialized, initializing now...")
		InitS3Client()
	}
	return s3Client
}

// ListS3Objects lists all objects under the specified prefix
func ListS3Objects(bucket string, prefix string) ([]types.Object, error) {
	client := GetS3Client()

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	resp, err := client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to list objects in bucket %s with prefix %s: %v", bucket, prefix, err)
		return nil, err
	}

	log.Printf("Successfully listed %d objects from S3 bucket %s", len(resp.Contents), bucket)
	return resp.Contents, nil
}

// ListS3Folders lists all top-level folders in the specified S3 bucket
func ListS3Folders(bucket string) ([]string, error) {
	client := GetS3Client()

	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
	}

	resp, err := client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to list folders in bucket %s: %v", bucket, err)
		return nil, err
	}

	var folders []string
	for _, prefix := range resp.CommonPrefixes {
		folders = append(folders, *prefix.Prefix)
	}
	log.Printf("Successfully listed %d folders from S3 bucket %s", len(folders), bucket)
	return folders, nil
}

// GetPresignedURL generates a presigned URL for an S3 object (valid for 15 mins by default)
func GetPresignedURL(bucket, key string) (string, error) {
	client := GetS3Client()

	// 直接用 s3.NewPresignClient
	presignClient := s3.NewPresignClient(client)

	req, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		log.Printf("Failed to generate presigned URL for %s: %v", key, err)
		return "", err
	}

	// log.Printf("Presigned URL generated for %s (15 min valid)", key)
	return req.URL, nil
}
