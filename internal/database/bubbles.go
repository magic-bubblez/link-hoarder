package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/magic_bubblez/link-hoarder/internal/models"
)

func CountBubblesForUser(ctx context.Context, uid string) (int, error) {
	var count int
	err := DB.QueryRow(ctx, `SELECT COUNT(*) FROM bubbles WHERE uid = $1`, uid).Scan(&count)
	return count, err
}

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
                    	)) FILTER (WHERE c.id IS NOT NULL), '[]') as tags,
        (SELECT COUNT(*) FROM items i WHERE i.bid = b.id) as item_count
        FROM bubbles b
        LEFT JOIN bridge br ON b.id = br.bid
        LEFT JOIN categories c ON br.cid = c.id
        WHERE b.uid = $1 GROUP BY b.id
        ORDER BY b.created_at DESC
    `
	rows, err := DB.Query(ctx, query, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b models.Bubble
		var tags_json []byte
		if err := rows.Scan(&b.ID, &b.Name, &b.CreatedAt, &tags_json, &b.ItemCount); err != nil {
			return nil, err
		}

		if len(tags_json) > 0 {
			_ = json.Unmarshal(tags_json, &b.Tags)
		}
		bubbles = append(bubbles, b)
	}
	return bubbles, nil
}

func UpdateBubbleName(ctx context.Context, uid string, bid string, name string) error {
	query := `UPDATE bubbles SET name = $1 WHERE id = $2 AND uid = $3`
	result, err := DB.Exec(ctx, query, name, bid, uid)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("bubble not found or permission denied")
	}
	return nil
}

func SetUserVisibility(ctx context.Context, uid string, slug *string) (*string, error) {
	query := `UPDATE users SET public_slug = $1 WHERE id = $2 RETURNING public_slug`
	var result *string
	err := DB.QueryRow(ctx, query, slug, uid).Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to update visibility: %w", err)
	}
	return result, nil
}

func GetPublicUserBubbles(ctx context.Context, slug string) (string, []models.Bubble, error) {
	// First find the user by slug
	var uid string
	err := DB.QueryRow(ctx, `SELECT id FROM users WHERE public_slug = $1`, slug).Scan(&uid)
	if err != nil {
		return "", nil, err
	}

	// Then get all their bubbles (reuse the same query shape)
	bubbles, err := GetAllBubbles(ctx, uid)
	if err != nil {
		return "", nil, err
	}
	return uid, bubbles, nil
}
