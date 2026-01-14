package main

import (
	"fmt"
	"log"

	"net/http"

	"github.com/magic_bubblez/link-hoarder/internal/auth"
	"github.com/magic_bubblez/link-hoarder/internal/database"
)

func main() {
	pool, err := database.Connection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	database.DB = pool
	defer database.DB.Close() 

	// req multiplexer. func is to match incoming request against the list
	// of registered patterns and call the handler for it
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(string)
		
		fmt.Fprintf(w, "Welcome to Bubbles!\nYour User ID is: %s", userID)
	})

	//wrapping mux with guestmiddleware, next in chain
	authHandler := auth.GuestMiddleware(mux)    

	fmt.Println("server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", authHandler); err != nil {
		log.Fatal(err)
	}
}