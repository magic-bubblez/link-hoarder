package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"

	"github.com/magic_bubblez/link-hoarder/internal/auth"
	"github.com/magic_bubblez/link-hoarder/internal/database"
	"github.com/magic_bubblez/link-hoarder/internal/middleware"
)

func main() {
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

	go startGuestCleanup()

	// Authenticated routes (behind GuestMiddleware)
	authMux := http.NewServeMux()
	authMux.HandleFunc("GET /api/me", GetMeHandler)
	authMux.HandleFunc("POST /api/bubbles", CreateBubbleHandler)
	authMux.HandleFunc("GET /api/bubbles", GetAllBubblesHandler)
	authMux.HandleFunc("DELETE /api/bubbles/{bid}", DeleteBubbleHandler)
	authMux.HandleFunc("PATCH /api/bubbles/{bid}", UpdateBubbleHandler)
	authMux.HandleFunc("PATCH /api/visibility", ToggleVisibilityHandler)
	authMux.HandleFunc("POST /api/bubbles/{bid}/items", AddItemHandler)
	authMux.HandleFunc("GET /api/bubbles/{bid}/items", GetItemsForBubbleHandler)
	authMux.HandleFunc("DELETE /api/bubbles/{bid}/items/{iid}", DeleteItemHandler)

	// Auth routes (need guest middleware for user ID)
	authMux.HandleFunc("GET /auth/google/login", auth.GoogleLoginHandler)
	authMux.HandleFunc("GET /auth/google/callback", auth.GoogleCallbackHandler)

	// Top-level mux: public routes + auth-wrapped routes
	topMux := http.NewServeMux()

	// Public API (no auth)
	topMux.HandleFunc("GET /api/public/{slug}", GetPublicBubblesHandler)
	topMux.HandleFunc("GET /api/public/{slug}/items/{bid}", GetPublicItemsHandler)

	// Public slug page — serve dashboard.html
	topMux.HandleFunc("GET /b/{slug}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/dashboard.html")
	})

	// Serve header images
	topMux.HandleFunc("GET /images/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if !strings.HasSuffix(name, ".webp") && !strings.HasSuffix(name, ".png") {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "frontend/images/"+name)
	})

	// Static files - catch-all without method to avoid conflict with /api/
	topMux.Handle("/", http.FileServer(http.Dir("frontend")))

	// Wrap auth routes with GuestMiddleware
	topMux.Handle("/api/", auth.GuestMiddleware(authMux))
	topMux.Handle("/auth/", auth.GuestMiddleware(authMux))

	limiter := middleware.NewRateLimiter(rate.Limit(5), 10)
	handler := middleware.CORS(limiter.Limit(topMux))

	fmt.Println("server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

func startGuestCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		result, err := database.DB.Exec(context.Background(),
			`DELETE FROM users WHERE is_guest = true AND created_at < NOW() - INTERVAL '5 days'`,
		)
		if err != nil {
			log.Println("Guest cleanup error:", err)
		} else {
			log.Printf("Guest cleanup: removed %d expired guests\n", result.RowsAffected())
		}
	}
}
