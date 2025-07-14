package googlesdk

import (
	"context"
	"fmt"
	"log"
	"os"

	"S3SyncGoogleDrive/internal/configs"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func GoogleConnect() {
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

	saveRefreshToken(token.RefreshToken)
}

func saveRefreshToken(refreshToken string) {
	err := os.WriteFile("config/refresh_token.txt", []byte(refreshToken), 0644)
	if err != nil {
		log.Println("Failed to save refresh token:", err)
	} else {
		log.Println("Refresh token saved to refresh_token.txt")
	}
}

func GetAccessTokenFromRefresh() string {
	ctx := context.Background()

	// 讀取 refresh_token
	refreshTokenBytes, err := os.ReadFile("config/fresh_token.txt")
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

func GetDriveService() *drive.Service {
	ctx := context.Background()

	refreshToken, err := os.ReadFile("config/refresh_token.txt")
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
