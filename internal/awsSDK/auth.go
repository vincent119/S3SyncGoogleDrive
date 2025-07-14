package awsSDK

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"log"
	"os"
)

// AwsConnect loads the AWS configuration using the environment variable AWS_PROFILE
func AwsConnect() aws.Config {
	awsProfile := os.Getenv("AWS_PROFILE")
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithSharedConfigProfile(awsProfile),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config with profile: %v", err)
	}
	return cfg
}

// AwsConnectWithRegion loads the AWS configuration with a specified region
func AwsConnectWithRegion(region string) aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config with region: %v", err)
	}
	return cfg
}
