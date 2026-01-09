package s3

import (
	"context"
	"log"
	"time"

	"S3SyncGoogleDrive/internal/awsSDK"
	"S3SyncGoogleDrive/internal/configs"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3API defines the interface for S3 client operations
type S3API interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// PresignAPI defines the interface for S3 presigner operations
type PresignAPI interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// S3Manager handles S3 operations
type S3Manager struct {
	Client        S3API
	PresignClient PresignAPI
}

// NewS3Manager creates a new S3Manager
func NewS3Manager(client S3API, presignClient PresignAPI) *S3Manager {
	return &S3Manager{
		Client:        client,
		PresignClient: presignClient,
	}
}

// NewDefaultManager creates an S3Manager with default AWS config
func NewDefaultManager() *S3Manager {
	cfg := awsSDK.AwsConnectWithRegion(configs.Config.S3.Region)
	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)
	log.Println("S3 client initialized successfully")
	return NewS3Manager(client, presignClient)
}

// ListS3Objects lists all objects under the specified prefix
func (m *S3Manager) ListS3Objects(bucket string, prefix string) ([]types.Object, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	resp, err := m.Client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to list objects in bucket %s with prefix %s: %v", bucket, prefix, err)
		return nil, err
	}

	log.Printf("Successfully listed %d objects from S3 bucket %s", len(resp.Contents), bucket)
	return resp.Contents, nil
}

// ListS3Folders lists all top-level folders in the specified S3 bucket
func (m *S3Manager) ListS3Folders(bucket string) ([]string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
	}

	resp, err := m.Client.ListObjectsV2(context.TODO(), input)
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
func (m *S3Manager) GetPresignedURL(bucket, key string) (string, error) {
	req, err := m.PresignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		log.Printf("Failed to generate presigned URL for %s: %v", key, err)
		return "", err
	}

	return req.URL, nil
}

