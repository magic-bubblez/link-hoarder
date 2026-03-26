// internal/database/items.go — was links.go
package database

import (
	"context"
	"fmt"

	"github.com/magic_bubblez/link-hoarder/internal/models"
)

func CountItemsForBubble(ctx context.Context, uid string, bid string) (int, error) {
	var count int
	err := DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM items i
		JOIN bubbles b ON i.bid = b.id
		WHERE i.bid = $1 AND b.uid = $2
	`, bid, uid).Scan(&count)
	return count, err
}

func AddItem(ctx context.Context, uid string, bid string, req models.AddItemRequest) (models.Item, error) {
	var item models.Item

	query := `
		INSERT INTO items (bid, content)
		SELECT $1, $2
		FROM bubbles
		WHERE id = $1 AND uid = $3
		RETURNING id, bid, content, created_at
	`
	err := DB.QueryRow(ctx, query, bid, req.Content, uid).Scan(
		&item.ID, &item.BubbleID, &item.Content, &item.CreatedAt,
	)
	if err != nil {
		return item, fmt.Errorf("failed to add item (permission denied): %w", err)
	}

	return item, nil
}

func GetItemsForBubble(ctx context.Context, uid string, bid string) ([]models.Item, error) {
	items := []models.Item{}

	query := `
		SELECT i.id, i.bid, i.content, i.created_at
		FROM items i
		JOIN bubbles b ON i.bid = b.id
		WHERE i.bid = $1 AND b.uid = $2
		ORDER BY i.created_at DESC
	`
	rows, err := DB.Query(ctx, query, bid, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.BubbleID, &item.Content, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func DeleteItem(ctx context.Context, uid string, bid string, itemID string) error {
	query := `
		DELETE FROM items
		WHERE id = $1 AND bid = $2
		AND EXISTS (SELECT 1 FROM bubbles WHERE id = $2 AND uid = $3)
	`
	result, err := DB.Exec(ctx, query, itemID, bid, uid)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("item not found or permission denied")
	}
	return nil
}

func DeleteBubble(ctx context.Context, uid string, bid string) error {
	query := `DELETE FROM bubbles WHERE id = $1 AND uid = $2`
	result, err := DB.Exec(ctx, query, bid, uid)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("bubble not found or permission denied")
	}
	return nil
}

func UpgradeGuestToUser(ctx context.Context, userID string, email string) error {
	query := `
		UPDATE users 
		SET is_guest = false, email = $1 
		WHERE id = $2 AND is_guest = true
	`
	result, err := DB.Exec(ctx, query, email, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or already registered")
	}
	return nil
}

func GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, is_guest, created_at, email, public_slug FROM users WHERE email = $1`
	err := DB.QueryRow(ctx, query, email).Scan(&user.ID, &user.IsGuest, &user.CreatedAt, &user.Email, &user.PublicSlug)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func DeleteGuestUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE id = $1 AND is_guest = true`
	_, err := DB.Exec(ctx, query, userID)
	return err
}

func GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	query := `SELECT id, is_guest, created_at, email, public_slug FROM users WHERE id = $1`
	err := DB.QueryRow(ctx, query, userID).Scan(&user.ID, &user.IsGuest, &user.CreatedAt, &user.Email, &user.PublicSlug)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
