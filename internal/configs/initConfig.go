package configs

import (
	//"fmt"
	"log"

	"github.com/spf13/viper"
)

type S3Config struct {
	BucketName   string `mapstructure:"bucketName"`
	Region       string `mapstructure:"region"`
	AccessKeyId  string `mapstructure:"accessKeyId"`
	SecretAccess string `mapstructure:"secretAccess"`
}

type DriveConfig struct {
	ClientID      string `mapstructure:"client_id"`
	ClientSecret  string `mapstructure:"client_secret"`
	RefreshToken  string `mapstructure:"refresh_token"`
	FolderID      string `mapstructure:"folder_id"`
	MaxConcurrent int    `mapstructure:"maxConcurrent"`
}

type BaseConfig struct {
	S3    S3Config    `mapstructure:"S3"`
	Drive DriveConfig `mapstructure:"Drive"`
}

var Config BaseConfig

func Init() {
	viper.SetConfigName("base")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read base.yaml: %v", err)
	}

	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Print test
	// fmt.Printf("AWS S3 Bucket: %s, Region: %s\n", Config.S3.BucketName, Config.S3.Region)
	// fmt.Printf("Google Drive Client ID: %s\n", Config.Drive.ClientID)
	// fmt.Printf("Google Drive Client Secret: %s\n", Config.Drive.ClientSecret)
}
