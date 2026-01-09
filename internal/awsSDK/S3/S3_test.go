package s3

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// MockS3Client is a mock implementation of S3API
type MockS3Client struct {
	ListObjectsV2Func func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if m.ListObjectsV2Func != nil {
		return m.ListObjectsV2Func(ctx, params, optFns...)
	}
	return &s3.ListObjectsV2Output{}, nil
}

// MockPresignClient is a mock implementation of PresignAPI
type MockPresignClient struct {
	PresignGetObjectFunc func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

func (m *MockPresignClient) PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	if m.PresignGetObjectFunc != nil {
		return m.PresignGetObjectFunc(ctx, params, optFns...)
	}
	return &v4.PresignedHTTPRequest{URL: "https://mock-url"}, nil
}

func TestListS3Objects(t *testing.T) {
	mockClient := &MockS3Client{
		ListObjectsV2Func: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			if *params.Bucket != "test-bucket" {
				t.Errorf("Expected bucket 'test-bucket', got %s", *params.Bucket)
			}
			return &s3.ListObjectsV2Output{
				Contents: []types.Object{
					{Key: aws.String("file1.txt"), Size: aws.Int64(100)},
					{Key: aws.String("file2.txt"), Size: aws.Int64(200)},
				},
			}, nil
		},
	}

	manager := NewS3Manager(mockClient, &MockPresignClient{})
	objects, err := manager.ListS3Objects("test-bucket", "prefix")

	if err != nil {
		t.Fatalf("ListS3Objects failed: %v", err)
	}

	if len(objects) != 2 {
		t.Errorf("Expected 2 objects, got %d", len(objects))
	}
}

func TestListS3ObjectsError(t *testing.T) {
	mockClient := &MockS3Client{
		ListObjectsV2Func: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			return nil, errors.New("AWS Error")
		},
	}

	manager := NewS3Manager(mockClient, &MockPresignClient{})
	_, err := manager.ListS3Objects("test-bucket", "prefix")

	if err == nil {
		t.Error("Expected error from ListS3Objects, got nil")
	}
}

func TestGetPresignedURL(t *testing.T) {
	mockPresignClient := &MockPresignClient{
		PresignGetObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
			if *params.Bucket != "test-bucket" {
				t.Errorf("Expected bucket 'test-bucket', got %s", *params.Bucket)
			}
			if *params.Key != "test-key" {
				t.Errorf("Expected key 'test-key', got %s", *params.Key)
			}
			return &v4.PresignedHTTPRequest{URL: "https://presigned-url.com"}, nil
		},
	}

	manager := NewS3Manager(&MockS3Client{}, mockPresignClient)
	url, err := manager.GetPresignedURL("test-bucket", "test-key")

	if err != nil {
		t.Fatalf("GetPresignedURL failed: %v", err)
	}

	if url != "https://presigned-url.com" {
		t.Errorf("Expected URL 'https://presigned-url.com', got %s", url)
	}
}

func TestGetPresignedURLError(t *testing.T) {
	mockPresignClient := &MockPresignClient{
		PresignGetObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
			return nil, errors.New("Presign Error")
		},
	}

	manager := NewS3Manager(&MockS3Client{}, mockPresignClient)
	_, err := manager.GetPresignedURL("test-bucket", "test-key")

	if err == nil {
		t.Error("Expected error from GetPresignedURL, got nil")
	}
}

func TestListS3Folders(t *testing.T) {
	mockClient := &MockS3Client{
		ListObjectsV2Func: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			if *params.Delimiter != "/" {
				t.Errorf("Expected delimiter '/', got %s", *params.Delimiter)
			}
			return &s3.ListObjectsV2Output{
				CommonPrefixes: []types.CommonPrefix{
					{Prefix: aws.String("folder1/")},
					{Prefix: aws.String("folder2/")},
				},
			}, nil
		},
	}

	manager := NewS3Manager(mockClient, &MockPresignClient{})
	folders, err := manager.ListS3Folders("test-bucket")

	if err != nil {
		t.Fatalf("ListS3Folders failed: %v", err)
	}

	if len(folders) != 2 {
		t.Errorf("Expected 2 folders, got %d", len(folders))
	}
	if folders[0] != "folder1/" {
		t.Errorf("Expected folder1/, got %s", folders[0])
	}
}

