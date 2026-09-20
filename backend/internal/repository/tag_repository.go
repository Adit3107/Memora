package repository

import (
	"context"
	"database/sql"
	"errors"

	"memora-backend/internal/models"
)

type TagRepository struct {
	db *sql.DB
}

func NewPostgresTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(tag models.Tag) (models.Tag, error) {
	ctx := context.Background()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO tags (user_id, name, created_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, created_at
	`, tag.UserID, tag.Name, tag.CreatedAt).Scan(
		&tag.ID,
		&tag.UserID,
		&tag.Name,
		&tag.CreatedAt,
	)
	if err != nil {
		return models.Tag{}, err
	}

	return tag, nil
}

func (r *TagRepository) List() ([]models.Tag, error) {
	ctx := context.Background()

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, created_at
		FROM tags
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]models.Tag, 0)
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

func (r *TagRepository) GetByID(id string) (models.Tag, error) {
	ctx := context.Background()

	var tag models.Tag
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, created_at
		FROM tags
		WHERE id = $1
	`, id).Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Tag{}, ErrNotFound
	}
	if err != nil {
		return models.Tag{}, err
	}

	return tag, nil
}

func (r *TagRepository) Update(id string, tag models.Tag) (models.Tag, error) {
	ctx := context.Background()

	err := r.db.QueryRowContext(ctx, `
		UPDATE tags
		SET user_id = $1, name = $2
		WHERE id = $3
		RETURNING id, user_id, name, created_at
	`, tag.UserID, tag.Name, id).Scan(
		&tag.ID,
		&tag.UserID,
		&tag.Name,
		&tag.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Tag{}, ErrNotFound
	}
	if err != nil {
		return models.Tag{}, err
	}

	return tag, nil
}

func (r *TagRepository) Delete(id string) error {
	ctx := context.Background()

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM tags WHERE id = $1
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

// Why this file exists:
// This repository persists normal user-created tags such as RAG, AI, Go, or DSA.
// The database unique constraint prevents one user from creating duplicate tag names.
// Tag recommendation and AI-generated tags are intentionally not part of this file.
