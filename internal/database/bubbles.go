package database

import (
	"context"
	"fmt"
	"github.com/magic_bubblez/link-hoarder/internal/models"
)

func CreateBubble(uid string, req models.CreateBubbleRequest) (models.Bubble, error) {
	ctx := context.Background()
	var bubble models.Bubble

	tx, err := DB.Begin(ctx)
	if err != nil {
		return bubble, err
	}
	defer tx.Rollback(ctx)


	err = tx.QueryRow(ctx, `
		INSERT INTO bubbles (uid, name)
		VALUES ($1, $2)
		RETURNING id, uid, name, created_at
	`, uid, req.Name).Scan(&bubble.ID, &bubble.UserID, &bubble.Name, &bubble.CreatedAt)

	if err != nil {
		return bubble, fmt.Errorf("failed to insert bubble: %w", err)
	}

	for _, tag := range req.Tags {
		var catID string

		// Create category if not exists
		err := tx.QueryRow(ctx, `
			WITH s AS (SELECT id FROM categories WHERE uid=$1 AND name=$2),
			i AS (INSERT INTO categories (uid, name) SELECT $1, $2 WHERE NOT EXISTS (SELECT 1 FROM s) RETURNING id)
			SELECT id FROM i UNION ALL SELECT id FROM s
		`, uid, tag).Scan(&catID)

		if err != nil {
			return bubble, fmt.Errorf("failed on tag '%s': %w", tag, err)
		}

		// Link in Bridge Table
		_, err = tx.Exec(ctx, `
			INSERT INTO bubble_categories (bubble_id, category_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, bubble.ID, catID)

		if err != nil {
			return bubble, fmt.Errorf("failed to link tag: %w", err)
		}
		
		// add to respose object
		bubble.Tags = append(bubble.Tags, models.Category{ID: catID, Name: tag})
	}
	
	
	if err := tx.Commit(ctx); err != nil {
		return bubble, err
	}

	return bubble, nil
}