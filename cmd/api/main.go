// cmd/api/main.go - the entry point for api server
package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/time/rate"

	"github.com/magic_bubblez/link-hoarder/internal/auth"
	"github.com/magic_bubblez/link-hoarder/internal/database"
	"github.com/magic_bubblez/link-hoarder/internal/middleware"
)

func main() {
	pool, err := database.Connection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	database.DB = pool
	defer database.DB.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(string)

		fmt.Fprintf(w, "Welcome to Bubbles!\nYour User ID is: %s", userID)
	})
	mux.HandleFunc("POST /bubbles", CreateBubbleHandler)
	mux.HandleFunc("POST /bubbles/{bid}/links", AddLinkHandler)
	mux.HandleFunc("GET /bubbles", GetAllBubblesHandler)

	limiter := middleware.NewRateLimiter(rate.Limit(5), 10)
	handler := limiter.Limit(auth.GuestMiddleware(mux))

	fmt.Println("server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
