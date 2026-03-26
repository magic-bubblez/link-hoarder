package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/magic_bubblez/link-hoarder/internal/database"
	"golang.org/x/time/rate"
)

type ContextKey string

const UserIDKey ContextKey = "user_id"

var guestCreationLimiter = rate.NewLimiter(rate.Every(time.Minute/30), 5)

func GuestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_pass")
		var userID string

		if err == nil {
			userID = cookie.Value

			// Verify the user still exists in the database
			_, dbErr := database.GetUserByID(r.Context(), userID)
			if dbErr != nil {
				// Cookie points to a deleted/nonexistent user — create a fresh guest
				fmt.Println("Stale cookie detected, creating new guest for old ID:", userID)
				userID, err = createGuestUser(w, r)
				if err != nil {
					return
				}
				fmt.Println("New Guest Created (replaced stale):", userID)
			} else {
				fmt.Println("User exists. ID:", userID)
			}
		} else {
			userID, err = createGuestUser(w, r)
			if err != nil {
				return
			}
			fmt.Println("New Guest Created:", userID)
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func createGuestUser(w http.ResponseWriter, r *http.Request) (string, error) {
	if !guestCreationLimiter.Allow() {
		http.Error(w, "Too many guests being created, try again shortly", http.StatusTooManyRequests)
		return "", fmt.Errorf("guest creation rate limit exceeded")
	}

	userID := uuid.New().String()

	_, dbErr := database.DB.Exec(context.Background(),
		`INSERT INTO users (id, is_guest, created_at) VALUES ($1, true, NOW())`,
		userID,
	)
	if dbErr != nil {
		fmt.Printf("Error saving guest to DB: %v\n", dbErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return "", dbErr
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_pass",
		Value:    userID,
		Path:     "/",
		Expires:  time.Now().Add(5 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	return userID, nil
}
