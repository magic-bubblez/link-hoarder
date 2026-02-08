package main

import (
	"encoding/json"
	"net/http"

	"github.com/magic_bubblez/link-hoarder/internal/auth"
	"github.com/magic_bubblez/link-hoarder/internal/database"
	"github.com/magic_bubblez/link-hoarder/internal/models"
)

func CreateBubbleHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)

	var req models.CreateBubbleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	bubble, err := database.CreateBubble(r.Context(), userID, req)
	if err != nil {
		http.Error(w, "Failed to create bubble: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bubble)
}

func AddLinkHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	var req models.AddLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	bid := r.PathValue("bid")

	link, err := database.AddLink(r.Context(), userID, bid, req)
	if err != nil {
		http.Error(w, "Failed to add link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Queue the scrape job for a worker to pick up (non-blocking if buffer has room)
	ScrapeJobs <- ScrapeJob{URL: link.URL, LinkID: link.ID}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(link)
}

func GetAllBubblesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	bubbles, err := database.GetAllBubbles(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get bubbles: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bubbles)
}

func GetLinksForBubbleHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	bid := r.PathValue("bid")

	links, err := database.GetLinksForBubble(r.Context(), userID, bid)
	if err != nil {
		http.Error(w, "Failed to get links: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

func DeleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	bid := r.PathValue("bid")
	linkID := r.PathValue("lid")

	err := database.DeleteLink(r.Context(), userID, bid, linkID)
	if err != nil {
		http.Error(w, "Failed to delete link: "+err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteBubbleHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	bid := r.PathValue("bid")

	err := database.DeleteBubble(r.Context(), userID, bid)
	if err != nil {
		http.Error(w, "Failed to delete bubble: "+err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
