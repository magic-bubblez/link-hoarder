// internal/models/models.go
package models

import (
	"time"
)

// users table
type User struct {
	ID        string    `json:"id"`
	IsGuest   bool      `json:"is_guest"`
	CreatedAt time.Time `json:"created_at"`
	Email     *string   `json:"email"`
}

// categories table
type Category struct {
	ID     string `json:"id"`
	UserID string `json:"uid"`
	Name   string `json:"name"`
}

// bubbles table
type Bubble struct {
	ID        string     `json:"id"`
	UserID    string     `json:"uid"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	Tags      []Category `json:"tags, omitempty"`
}

// links table
type Link struct {
	ID        string    `json:"id"`
	BubbleID  string    `json:"bid"`
	URL       string    `json:"url"`
	Title     *string   `json:"title"`
	IconURL   *string   `json:"icon_url"`
	CreatedAt time.Time `json:"created_at"`
}

// bridge table (many to many relationship)
type Bridge struct {
	BubbleID   string `json:"bid"`
	CategoryID string `json:"cid"`
}

// create bubble request payload
type CreateBubbleRequest struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// add link payload
type AddLinkRequest struct {
	Link string `json:"link"`
}
