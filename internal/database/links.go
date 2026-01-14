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

// Update db with links for a bubble
func UpdateLinkData(ctx context.Context, linkID string, title string, imagePreview string) error {
	query := `UPDATE links SET title = $1, icon_url = $2 WHERE id = $3`
	_, err := DB.Exec(ctx, query, title, imagePreview, linkID)
	return err
}
