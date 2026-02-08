// internal/database/links.go
package database

import (
	"context"
	"fmt"

	"github.com/magic_bubblez/link-hoarder/internal/models"
)

func AddLink(ctx context.Context, uid string, bid string, req models.AddLinkRequest) (models.Link, error) {
	var link models.Link

	query := `
		INSERT INTO links (bid, url)
		SELECT $1, $2
		FROM bubbles
		WHERE id = $1 AND uid = $3
		RETURNING id, bid, url, created_at
	`
	err := DB.QueryRow(ctx, query, bid, req.Link, uid).Scan(
		&link.ID, &link.BubbleID, &link.URL, &link.CreatedAt,
	)
	if err != nil {
		return link, fmt.Errorf("failed to add link (permission denied): %w", err)
	}

	return link, nil
}

func UpdateLinkData(ctx context.Context, linkID string, title string, imagePreview string) error {
	query := `UPDATE links SET title = $1, icon_url = $2 WHERE id = $3`
	_, err := DB.Exec(ctx, query, title, imagePreview, linkID)
	return err
}

func GetLinksForBubble(ctx context.Context, uid string, bid string) ([]models.Link, error) {
	links := []models.Link{}

	query := `
		SELECT l.id, l.bid, l.url, l.title, l.icon_url, l.created_at
		FROM links l
		JOIN bubbles b ON l.bid = b.id
		WHERE l.bid = $1 AND b.uid = $2
		ORDER BY l.created_at DESC
	`
	rows, err := DB.Query(ctx, query, bid, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var link models.Link
		if err := rows.Scan(&link.ID, &link.BubbleID, &link.URL, &link.Title, &link.IconURL, &link.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func DeleteLink(ctx context.Context, uid string, bid string, linkID string) error {
	query := `
		DELETE FROM links
		WHERE id = $1 AND bid = $2
		AND EXISTS (SELECT 1 FROM bubbles WHERE id = $2 AND uid = $3)
	`
	result, err := DB.Exec(ctx, query, linkID, bid, uid)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("link not found or permission denied")
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
	query := `SELECT id, is_guest, created_at, email FROM users WHERE email = $1`
	err := DB.QueryRow(ctx, query, email).Scan(&user.ID, &user.IsGuest, &user.CreatedAt, &user.Email)
	if err != nil {
		return nil, err // includes pgx.ErrNoRows if not found
	}
	return &user, nil
}

func DeleteGuestUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE id = $1 AND is_guest = true`
	_, err := DB.Exec(ctx, query, userID)
	return err
}
