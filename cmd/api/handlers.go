package main

import (
	"encoding/json"
	"net/http"
	"context"
	"fmt"

	"github.com/magic_bubblez/link-hoarder/internal/scraper"
	"github.com/magic_bubblez/link-hoarder/internal/auth"
	"github.com/magic_bubblez/link-hoarder/internal/database"
	"github.com/magic_bubblez/link-hoarder/internal/models"
)

//to create new bubble
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

//to add new links to a bubble 
func AddLinkHandler(w http.ResponseWriter, r *http.Request){
	userID := r.Context().Value(auth.UserIDKey).(string)
	var req models.AddLinkRequest
	if err:= json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	bid := r.PathValue("bid")

	link, err := database.AddLink(r.Context(), userID, bid, req)
	if err != nil{
		http.Error(w, "Failed to add link: "+err.Error(), http.StatusInternalServerError)
		return
	}
	//Scrape the title and preview image from link
	go func(url string, linkID string){
		bgctx := context.Background() 
		title, image, err := scraper.ScrapeTitle(url)
		if err != nil{
			fmt.Printf("Failed to scrape title for %s: %v\n", url, err)
			return
		}

		var titleStr, imageStr string
		if title != nil {
			titleStr = *title
		}
		if image != nil {
			imageStr = *image
		}
		if titleStr != "" || imageStr != "" {       //loose checking here
			err := database.UpdateLinkData(bgctx, linkID, titleStr, imageStr)
			if err != nil{
				fmt.Printf("DB update failed: %v\n", err)
			} else {
				fmt.Printf("Title/Image updated for link %s\n", linkID)
			}
		}
	}(link.URL, link.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(link)
}

//to get all bubbles for user 
func GetAllBubblesHandler(w http.ResponseWriter, r *http.Request){
	userID := r.Context().Value(auth.UserIDKey).(string)
	bubbles, err := database.GetAllBubbles(r.Context(), userID)
	if err != nil{
		http.Error(w, "Failed to get bubbles: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bubbles)
}

