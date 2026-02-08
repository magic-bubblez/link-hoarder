package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/magic_bubblez/link-hoarder/internal/models"
)

// data transfer object (DTO) pattern
func CreateBubble(ctx context.Context, uid string, req models.CreateBubbleRequest) (models.Bubble, error) {
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
			INSERT INTO bridge (bid, cid)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, bubble.ID, catID)

		if err != nil {
			return bubble, fmt.Errorf("failed to link tag: %w", err)
		}
		bubble.Tags = append(bubble.Tags, models.Category{ID: catID, Name: tag})
	}
	if err := tx.Commit(ctx); err != nil {
		return bubble, err
	}
	return bubble, nil
}

func GetAllBubbles(ctx context.Context, uid string) ([]models.Bubble, error) {
	bubbles := []models.Bubble{}
	query := `
        SELECT b.id, b.name, b.created_at,
        COALESCE(JSON_AGG(json_build_object(
            			'id', c.id, 
                        'uid', c.uid, 
                        'name', c.name
                    	)) FILTER (WHERE c.id IS NOT NULL), '[]') as tags
        FROM bubbles b
        LEFT JOIN bridge br ON b.id = br.bid
        LEFT JOIN categories c ON br.cid = c.id
        WHERE b.uid = $1 GROUP BY b.id
    `
	rows, err := DB.Query(ctx, query, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b models.Bubble
		var tags_json []byte // scan the raw json bytes first
		if err := rows.Scan(&b.ID, &b.Name, &b.CreatedAt, &tags_json); err != nil {
			return nil, err
		}

		if len(tags_json) > 0 {
			_ = json.Unmarshal(tags_json, &b.Tags)
		}
		bubbles = append(bubbles, b)
	}
	return bubbles, nil
}
