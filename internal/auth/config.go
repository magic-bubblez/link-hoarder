// internal/auth/config.go - Google OAuth configuration
package auth

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleConfig *oauth2.Config

func InitGoogleAuth() {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	GoogleConfig = &oauth2.Config{
		RedirectURL:  baseURL + "/auth/google/callback",
		ClientID:     os.Getenv("GCLIENT_ID"),
		ClientSecret: os.Getenv("GCLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}
