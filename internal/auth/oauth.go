// internal/auth/oauth.go
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/magic_bubblez/link-hoarder/internal/database"
)

// response from Google's userinfo endpoint
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string) //get userID from guest middleware
	// we need userid in callback to upgrade the guest
	url := GoogleConfig.AuthCodeURL(userID)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	guestUserID := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange auth code for access token
	token, err := GoogleConfig.Exchange(ctx, code)
	if err != nil {
		fmt.Printf("OAuth exchange failed: %v\n", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	userInfo, err := fetchGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		fmt.Printf("Failed to fetch user info: %v\n", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	existingUser, err := database.GetUserByEmail(ctx, userInfo.Email)

	var finalUserID string

	if err == nil && existingUser != nil {
		finalUserID = existingUser.ID
		fmt.Printf("Returning user logged in: %s (%s)\n", userInfo.Email, finalUserID)

		if guestUserID != finalUserID {
			_ = database.DeleteGuestUser(ctx, guestUserID)
		}
	} else {
		err = database.UpgradeGuestToUser(ctx, guestUserID, userInfo.Email)
		if err != nil {
			fmt.Printf("Failed to upgrade guest: %v\n", err)
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}
		finalUserID = guestUserID
		fmt.Printf("Guest upgraded to user: %s (%s)\n", userInfo.Email, finalUserID)
	}

	// session cookie with the final user ID
	http.SetCookie(w, &http.Cookie{
		Name:     "session_pass",
		Value:    finalUserID,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// fetchGoogleUserInfo calls Google's API to get user profile
func fetchGoogleUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google API returned status %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
