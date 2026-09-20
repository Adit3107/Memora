package repository

import (
	"context"
	"database/sql"
)

type ContentTagRepository struct {
	db *sql.DB
}

func NewPostgresContentTagRepository(db *sql.DB) *ContentTagRepository {
	return &ContentTagRepository{db: db}
}

func (r *ContentTagRepository) SetTags(contentID string, tagIDs []string) error {
	ctx := context.Background()

	// Transaction keeps the delete+insert replacement as one safe unit of work.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM content_tags WHERE content_id = $1
	`, contentID); err != nil {
		return err
	}

	for _, tagID := range tagIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO content_tags (content_id, tag_id)
			VALUES ($1, $2)
			ON CONFLICT (content_id, tag_id) DO NOTHING
		`, contentID, tagID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ContentTagRepository) ListTagIDs(contentID string) ([]string, error) {
	ctx := context.Background()

	rows, err := r.db.QueryContext(ctx, `
		SELECT tag_id
		FROM content_tags
		WHERE content_id = $1
		ORDER BY created_at ASC
	`, contentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagIDs := make([]string, 0)
	for rows.Next() {
		var tagID string
		if err := rows.Scan(&tagID); err != nil {
			return nil, err
		}
		tagIDs = append(tagIDs, tagID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tagIDs, nil
}

func (r *ContentTagRepository) RemoveTag(contentID string, tagID string) error {
	ctx := context.Background()

	_, err := r.db.ExecContext(ctx, `
		DELETE FROM content_tags
		WHERE content_id = $1 AND tag_id = $2
	`, contentID, tagID)
	return err
}

func (r *ContentTagRepository) ListContentIDsByTag(tagID string) ([]string, error) {
	ctx := context.Background()

	rows, err := r.db.QueryContext(ctx, `
		SELECT content_id
		FROM content_tags
		WHERE tag_id = $1
		ORDER BY created_at DESC
	`, tagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contentIDs := make([]string, 0)
	for rows.Next() {
		var contentID string
		if err := rows.Scan(&contentID); err != nil {
			return nil, err
		}
		contentIDs = append(contentIDs, contentID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return contentIDs, nil
}

func (r *ContentTagRepository) RemoveTagEverywhere(tagID string) error {
	ctx := context.Background()

	_, err := r.db.ExecContext(ctx, `
		DELETE FROM content_tags WHERE tag_id = $1
	`, tagID)
	return err
}

func (r *ContentTagRepository) RemoveContent(contentID string) error {
	ctx := context.Background()

	_, err := r.db.ExecContext(ctx, `
		DELETE FROM content_tags WHERE content_id = $1
	`, contentID)
	return err
}

// Why this file exists:
// content_tags is a join table because content and tags are many-to-many.
// One content item can have many tags, and one tag can belong to many content items.
// The database primary key prevents duplicate tag links for the same content item.
