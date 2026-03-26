package main

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"

	"github.com/magic_bubblez/link-hoarder/internal/auth"
	"github.com/magic_bubblez/link-hoarder/internal/database"
	"github.com/magic_bubblez/link-hoarder/internal/models"
)

// Guest limits
const (
	GuestMaxBubbles        = 3
	GuestMaxItemsPerBubble = 10
)

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func generateSlug() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	rand.Read(b)
	for i := range b {
		b[i] = chars[b[i]%byte(len(chars))]
	}
	return string(b)
}

func CreateBubbleHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)

	user, err := database.GetUserByID(r.Context(), userID)
	if err != nil {
		jsonError(w, "failed to verify user", http.StatusInternalServerError)
		return
	}
	if user.IsGuest {
		count, err := database.CountBubblesForUser(r.Context(), userID)
		if err != nil {
			jsonError(w, "failed to check bubble count", http.StatusInternalServerError)
			return
		}
		if count >= GuestMaxBubbles {
			jsonError(w, "guest limit reached — sign up to create more bubbles!", http.StatusForbidden)
			return
		}
	}

	var req models.CreateBubbleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON format", http.StatusBadRequest)
		return
	}
	bubble, err := database.CreateBubble(r.Context(), userID, req)
	if err != nil {
		log.Printf("CreateBubble error: %v", err)
		jsonError(w, "failed to create bubble", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bubble)
}

func AddItemHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	bid := r.PathValue("bid")

	user, err := database.GetUserByID(r.Context(), userID)
	if err != nil {
		jsonError(w, "failed to verify user", http.StatusInternalServerError)
		return
	}
	if user.IsGuest {
		count, err := database.CountItemsForBubble(r.Context(), userID, bid)
		if err != nil {
			jsonError(w, "failed to check item count", http.StatusInternalServerError)
			return
		}
		if count >= GuestMaxItemsPerBubble {
			jsonError(w, "guest limit reached — sign up to add more items!", http.StatusForbidden)
			return
		}
	}

	var req models.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON format", http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		jsonError(w, "content is required", http.StatusBadRequest)
		return
	}

	item, err := database.AddItem(r.Context(), userID, bid, req)
	if err != nil {
		log.Printf("AddItem error: %v", err)
		jsonError(w, "failed to add item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func GetAllBubblesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	bubbles, err := database.GetAllBubbles(r.Context(), userID)
	if err != nil {
		log.Printf("GetAllBubbles error: %v", err)
		jsonError(w, "failed to get bubbles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bubbles)
}

func GetItemsForBubbleHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	bid := r.PathValue("bid")

	items, err := database.GetItemsForBubble(r.Context(), userID, bid)
	if err != nil {
		log.Printf("GetItemsForBubble error: %v", err)
		jsonError(w, "failed to get items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	bid := r.PathValue("bid")
	itemID := r.PathValue("iid")

	err := database.DeleteItem(r.Context(), userID, bid, itemID)
	if err != nil {
		log.Printf("DeleteItem error: %v", err)
		jsonError(w, "failed to delete item", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteBubbleHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	bid := r.PathValue("bid")

	err := database.DeleteBubble(r.Context(), userID, bid)
	if err != nil {
		log.Printf("DeleteBubble error: %v", err)
		jsonError(w, "failed to delete bubble", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func UpdateBubbleHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	bid := r.PathValue("bid")

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON format", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	err := database.UpdateBubbleName(r.Context(), userID, bid, req.Name)
	if err != nil {
		log.Printf("UpdateBubble error: %v", err)
		jsonError(w, "failed to update bubble", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetMeHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)

	user, err := database.GetUserByID(r.Context(), userID)
	if err != nil {
		jsonError(w, "failed to get user info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ── Visibility toggle (auth required, non-guest only) ──

func ToggleVisibilityHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)

	user, err := database.GetUserByID(r.Context(), userID)
	if err != nil {
		jsonError(w, "failed to verify user", http.StatusInternalServerError)
		return
	}
	if user.IsGuest {
		jsonError(w, "sign up to share bubbles publicly", http.StatusForbidden)
		return
	}

	var req struct {
		Public bool `json:"public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	var slug *string
	if req.Public {
		s := generateSlug()
		slug = &s
	}

	result, err := database.SetUserVisibility(r.Context(), userID, slug)
	if err != nil {
		log.Printf("ToggleVisibility error: %v", err)
		jsonError(w, "failed to update visibility", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]*string{"public_slug": result})
}

// ── Public endpoints (no auth) ──

func GetPublicBubblesHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	_, bubbles, err := database.GetPublicUserBubbles(r.Context(), slug)
	if err != nil {
		jsonError(w, "page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bubbles)
}

func GetPublicItemsHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	bid := r.PathValue("bid")

	// Verify the slug is valid (user exists and is public)
	uid, _, err := database.GetPublicUserBubbles(r.Context(), slug)
	if err != nil {
		jsonError(w, "page not found", http.StatusNotFound)
		return
	}

	// Use the normal GetItemsForBubble but with the public user's ID
	items, err := database.GetItemsForBubble(r.Context(), uid, bid)
	if err != nil {
		jsonError(w, "failed to get items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
