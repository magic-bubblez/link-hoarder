package main

import (
	"encoding/json"
	"net/http"

	"github.com/magic-bubblez/link-hoarder/internal/auth"
	"github.com/magic-bubblez/link-hoarder/internal/database"
	"github.com/magic-bubblez/link-hoarder/internal/models"
)

func CreateBubbleHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)

	// Parse the JSON req body
	var req models.CreateBubbleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	bubble, err := database.CreateBubble(userID, req)
	if err != nil {
		http.Error(w, "Failed to create bubble: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bubble)
}