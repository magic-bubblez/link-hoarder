// cmd/api/main.go - the entry point for api server
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"

	"github.com/magic_bubblez/link-hoarder/internal/auth"
	"github.com/magic_bubblez/link-hoarder/internal/database"
	"github.com/magic_bubblez/link-hoarder/internal/middleware"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env vars")
	}

	pool, err := database.Connection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	database.DB = pool
	defer database.DB.Close()

	auth.InitGoogleAuth()

	// Start the worker pool: 50 workers, buffer for 1000 pending jobs
	StartScrapeWorkers(50, 1000)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(string)

		fmt.Fprintf(w, "Welcome to Bubbles!\nYour User ID is: %s", userID)
	})

	// Bubble routes
	mux.HandleFunc("POST /bubbles", CreateBubbleHandler)
	mux.HandleFunc("GET /bubbles", GetAllBubblesHandler)
	mux.HandleFunc("DELETE /bubbles/{bid}", DeleteBubbleHandler)

	// Link routes
	mux.HandleFunc("POST /bubbles/{bid}/links", AddLinkHandler)
	mux.HandleFunc("GET /bubbles/{bid}/links", GetLinksForBubbleHandler)
	mux.HandleFunc("DELETE /bubbles/{bid}/links/{lid}", DeleteLinkHandler)

	// Auth routes (Google OAuth)
	mux.HandleFunc("GET /auth/google/login", auth.GoogleLoginHandler)
	mux.HandleFunc("GET /auth/google/callback", auth.GoogleCallbackHandler)

	limiter := middleware.NewRateLimiter(rate.Limit(5), 10)
	handler := limiter.Limit(auth.GuestMiddleware(mux))

	fmt.Println("server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
