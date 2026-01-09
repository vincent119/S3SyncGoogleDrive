package googlesdk

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vincent119/S3SyncGoogleDrive/internal/configs"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// FileIO interface for file operations
type FileIO interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
}

// RealFileIO implements FileIO using os package
type RealFileIO struct{}

func (r *RealFileIO) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (r *RealFileIO) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// AuthManager handles Google authentication
type AuthManager struct {
	fs FileIO
}

// NewAuthManager creates a new AuthManager
func NewAuthManager(fs FileIO) *AuthManager {
	return &AuthManager{fs: fs}
}

// NewDefaultAuthManager creates an AuthManager with RealFileIO
func NewDefaultAuthManager() *AuthManager {
	return NewAuthManager(&RealFileIO{})
}

func (a *AuthManager) GoogleConnect() {
	ctx := context.Background()

	conf := &oauth2.Config{
		ClientID:     configs.Config.Drive.ClientID,
		ClientSecret: configs.Config.Drive.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/drive"},
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}

	authURL := conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Visit the following URL, authorize the app, and paste the authorization code below:")
	fmt.Println(authURL)

	var code string
	fmt.Print("Enter the authorization code: ")
	fmt.Scan(&code)

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Token exchange failed: %v", err)
	}

	fmt.Println("Access Token:", token.AccessToken)
	fmt.Println("Refresh Token (save this):", token.RefreshToken)

	a.saveRefreshToken(token.RefreshToken)
}

func (a *AuthManager) saveRefreshToken(refreshToken string) {
	err := a.fs.WriteFile("config/refresh_token.txt", []byte(refreshToken), 0644)
	if err != nil {
		log.Println("Failed to save refresh token:", err)
	} else {
		log.Println("Refresh token saved to refresh_token.txt")
	}
}

func (a *AuthManager) GetAccessTokenFromRefresh() string {
	ctx := context.Background()

	// 讀取 refresh_token
	refreshTokenBytes, err := a.fs.ReadFile("config/refresh_token.txt")
	if err != nil {
		log.Fatalf("Failed to read refresh_token.txt: %v", err)
	}
	refreshToken := string(refreshTokenBytes)

	conf := &oauth2.Config{
		ClientID:     configs.Config.Drive.ClientID,
		ClientSecret: configs.Config.Drive.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/drive"},
	}

	// 用 refresh_token 換 access_token
	token := &oauth2.Token{RefreshToken: refreshToken}
	tokenSource := conf.TokenSource(ctx, token)

	newToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalf("Failed to get access token from refresh token: %v", err)
	}

	fmt.Println("New Access Token:", newToken.AccessToken)
	return newToken.AccessToken
}

func (a *AuthManager) GetDriveService() *drive.Service {
	ctx := context.Background()

	refreshToken, err := a.fs.ReadFile("config/refresh_token.txt")
	if err != nil {
		log.Fatalf("Failed to read refresh_token.txt: %v", err)
	}

	conf := &oauth2.Config{
		ClientID:     configs.Config.Drive.ClientID,
		ClientSecret: configs.Config.Drive.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/drive"},
	}

	token := &oauth2.Token{RefreshToken: string(refreshToken)}
	tokenSource := conf.TokenSource(ctx, token)

	srv, err := drive.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		log.Fatalf("Failed to create Drive service: %v", err)
	}
	return srv
}

// Global Wrappers for backward compatibility
func GoogleConnect() {
	NewDefaultAuthManager().GoogleConnect()
}

func GetAccessTokenFromRefresh() string {
	return NewDefaultAuthManager().GetAccessTokenFromRefresh()
}

func GetDriveService() *drive.Service {
	return NewDefaultAuthManager().GetDriveService()
}

