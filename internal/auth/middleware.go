package auth

import (
	"context" //holds data for passing it down the chain of handlers
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/magic_bubblez/link-hoarder/internal/database"
)

type ContextKey string

const UserIDKey ContextKey = "user_id"

func GuestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //w is an interface, anything written to it goes to client
		// r is struct pointer. contains incoming request data
		cookie, err := r.Cookie("session_pass")
		var userID string

		if err == nil {
			userID = cookie.Value
			fmt.Println("User exists. ID: ", userID)
		} else {
			userID = uuid.New().String()

			_, dbErr := database.DB.Exec(context.Background(),
				`INSERT INTO users (id, is_guest, created_at) VALUES ($1, true, NOW())`,
				userID,
			)
			if dbErr != nil {
				fmt.Printf("Error saving guest to DB: %v\n", dbErr)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session_pass",
				Value:    userID,
				Path:     "/",
				Expires:  time.Now().Add(5 * 24 * time.Hour), // 5 Days
				HttpOnly: true,                               // Security: js cannot read this
			})

			fmt.Println("New Guest Created:", userID)
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
