package awsSDK

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// ConfigLoader defines a function type for loading AWS config
type ConfigLoader func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error)

// AWSManager handles AWS authentication
type AWSManager struct {
	Loader ConfigLoader
}

// NewAWSManager creates a new AWSManager
func NewAWSManager(loader ConfigLoader) *AWSManager {
	return &AWSManager{Loader: loader}
}

// NewDefaultAWSManager creates an AWSManager using the real config loader
func NewDefaultAWSManager() *AWSManager {
	return NewAWSManager(config.LoadDefaultConfig)
}

// Connect loads configuration using AWS_PROFILE
func (m *AWSManager) Connect() aws.Config {
	awsProfile := os.Getenv("AWS_PROFILE")
	cfg, err := m.Loader(
		context.Background(),
		config.WithSharedConfigProfile(awsProfile),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config with profile: %v", err)
	}
	return cfg
}

// ConnectWithRegion loads configuration with a specific region
func (m *AWSManager) ConnectWithRegion(region string) aws.Config {
	cfg, err := m.Loader(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config with region: %v", err)
	}
	return cfg
}

// Wrappers for backward compatibility
func AwsConnect() aws.Config {
	return NewDefaultAWSManager().Connect()
}

func AwsConnectWithRegion(region string) aws.Config {
	return NewDefaultAWSManager().ConnectWithRegion(region)
}

