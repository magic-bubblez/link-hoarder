// internal/models/models.go
package models

import (
	"time"
)

// Users table
type User struct {
	ID         string    `json:"id"`
	IsGuest    bool      `json:"is_guest"`
	CreatedAt  time.Time `json:"created_at"`
	Email      *string   `json:"email"`
	PublicSlug *string   `json:"public_slug,omitempty"`
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
	Tags      []Category `json:"tags,omitempty"`
	ItemCount int        `json:"item_count"`
}

// items table
type Item struct {
	ID        string    `json:"id"`
	BubbleID  string    `json:"bid"`
	Content   string    `json:"content"`
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

// add item payload
type AddItemRequest struct {
	Content string `json:"content"`
}
