package repository

import (
	"context"
	"database/sql"
	"errors"

	"memora-backend/internal/models"
)

type ContentRepository struct {
	// *sql.DB is shared by all repositories. It manages a pool of DB connections.
	db *sql.DB
}

func NewPostgresContentRepository(db *sql.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

func (r *ContentRepository) Create(content models.Content) (models.Content, error) {
	ctx := context.Background()

	// RETURNING asks Postgres to send back generated values like id.
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO content (
			user_id, space_id, name, title, description, type,
			source_url, thumbnail_url, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, user_id, space_id, name, title, description, type,
			source_url, thumbnail_url, created_at, updated_at
	`,
		content.UserID,
		content.SpaceID,
		content.Name,
		content.Title,
		content.Description,
		content.Type,
		content.SourceURL,
		content.ThumbnailURL,
		content.CreatedAt,
		content.UpdatedAt,
	).Scan(
		&content.ID,
		&content.UserID,
		&content.SpaceID,
		&content.Name,
		&content.Title,
		&content.Description,
		&content.Type,
		&content.SourceURL,
		&content.ThumbnailURL,
		&content.CreatedAt,
		&content.UpdatedAt,
	)
	if err != nil {
		return models.Content{}, err
	}

	return content, nil
}

func (r *ContentRepository) List() ([]models.Content, error) {
	ctx := context.Background()

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, space_id, name, title, description, type,
			source_url, thumbnail_url, created_at, updated_at
		FROM content
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contents := make([]models.Content, 0)
	for rows.Next() {
		var content models.Content
		if err := scanContent(rows, &content); err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return contents, nil
}

func (r *ContentRepository) ListByUserID(userID string) ([]models.Content, error) {
	ctx := context.Background()

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, space_id, name, title, description, type,
			source_url, thumbnail_url, created_at, updated_at
		FROM content
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contents := make([]models.Content, 0)
	for rows.Next() {
		var content models.Content
		if err := scanContent(rows, &content); err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return contents, nil
}

func (r *ContentRepository) GetByID(id string) (models.Content, error) {
	ctx := context.Background()

	var content models.Content
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, space_id, name, title, description, type,
			source_url, thumbnail_url, created_at, updated_at
		FROM content
		WHERE id = $1
	`, id).Scan(
		&content.ID,
		&content.UserID,
		&content.SpaceID,
		&content.Name,
		&content.Title,
		&content.Description,
		&content.Type,
		&content.SourceURL,
		&content.ThumbnailURL,
		&content.CreatedAt,
		&content.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Content{}, ErrNotFound
	}
	if err != nil {
		return models.Content{}, err
	}

	return content, nil
}

func (r *ContentRepository) Update(id string, content models.Content) (models.Content, error) {
	ctx := context.Background()

	err := r.db.QueryRowContext(ctx, `
		UPDATE content
		SET user_id = $1, space_id = $2, name = $3, title = $4, description = $5,
			type = $6, source_url = $7, thumbnail_url = $8, updated_at = $9
		WHERE id = $10
		RETURNING id, user_id, space_id, name, title, description, type,
			source_url, thumbnail_url, created_at, updated_at
	`,
		content.UserID,
		content.SpaceID,
		content.Name,
		content.Title,
		content.Description,
		content.Type,
		content.SourceURL,
		content.ThumbnailURL,
		content.UpdatedAt,
		id,
	).Scan(
		&content.ID,
		&content.UserID,
		&content.SpaceID,
		&content.Name,
		&content.Title,
		&content.Description,
		&content.Type,
		&content.SourceURL,
		&content.ThumbnailURL,
		&content.CreatedAt,
		&content.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Content{}, ErrNotFound
	}
	if err != nil {
		return models.Content{}, err
	}

	return content, nil
}

func (r *ContentRepository) Delete(id string) error {
	ctx := context.Background()

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM content WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func scanContent(rows *sql.Rows, content *models.Content) error {
	// Scan copies SQL column values into Go struct fields in the same order as SELECT.
	return rows.Scan(
		&content.ID,
		&content.UserID,
		&content.SpaceID,
		&content.Name,
		&content.Title,
		&content.Description,
		&content.Type,
		&content.SourceURL,
		&content.ThumbnailURL,
		&content.CreatedAt,
		&content.UpdatedAt,
	)
}

// Why this file exists:
// This repository persists content metadata in PostgreSQL.
// It stores only metadata, not original files, transcripts, chunks, embeddings, or OCR output.
// SQL placeholders like $1 keep user values separate from SQL text and protect against injection.
