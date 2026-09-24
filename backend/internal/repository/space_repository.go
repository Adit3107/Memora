package repository

import (
	"context"
	"database/sql"
	"errors"

	"memora-backend/internal/models"
)

type SpaceRepository struct {
	// Repositories share the app-wide database pool instead of opening their own connections.
	db *sql.DB
}

func NewPostgresSpaceRepository(db *sql.DB) *SpaceRepository {
	return &SpaceRepository{db: db}
}

func (r *SpaceRepository) Create(space models.Space) (models.Space, error) {
	ctx := context.Background()

	// user_id is a foreign key in the database, so Postgres enforces that the owner exists.
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO spaces (user_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, name, description, created_at, updated_at
	`, space.UserID, space.Name, space.Description, space.CreatedAt, space.UpdatedAt).Scan(
		&space.ID,
		&space.UserID,
		&space.Name,
		&space.Description,
		&space.CreatedAt,
		&space.UpdatedAt,
	)
	if err != nil {
		return models.Space{}, err
	}

	return space, nil
}

func (r *SpaceRepository) List() ([]models.Space, error) {
	ctx := context.Background()

	// For now this lists all spaces; auth/user filtering comes later when auth exists.
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM spaces
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spaces := make([]models.Space, 0)
	for rows.Next() {
		var space models.Space
		// Scan maps SQL columns in SELECT order into the Space struct.
		if err := rows.Scan(
			&space.ID,
			&space.UserID,
			&space.Name,
			&space.Description,
			&space.CreatedAt,
			&space.UpdatedAt,
		); err != nil {
			continue
		}
		spaces = append(spaces, space)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return spaces, nil
}

func (r *SpaceRepository) ListByUserID(userID string) ([]models.Space, error) {
	ctx := context.Background()

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM spaces
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spaces := make([]models.Space, 0)
	for rows.Next() {
		var space models.Space
		if err := rows.Scan(
			&space.ID,
			&space.UserID,
			&space.Name,
			&space.Description,
			&space.CreatedAt,
			&space.UpdatedAt,
		); err != nil {
			continue
		}
		spaces = append(spaces, space)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return spaces, nil
}

func (r *SpaceRepository) GetByID(id string) (models.Space, error) {
	ctx := context.Background()

	var space models.Space
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM spaces
		WHERE id = $1
	`, id).Scan(
		&space.ID,
		&space.UserID,
		&space.Name,
		&space.Description,
		&space.CreatedAt,
		&space.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		// Keep repository errors consistent across in-memory and Postgres implementations.
		return models.Space{}, ErrNotFound
	}
	if err != nil {
		return models.Space{}, err
	}

	return space, nil
}

func (r *SpaceRepository) Update(id string, space models.Space) (models.Space, error) {
	ctx := context.Background()

	err := r.db.QueryRowContext(ctx, `
		UPDATE spaces
		SET user_id = $1, name = $2, description = $3, updated_at = $4
		WHERE id = $5
		RETURNING id, user_id, name, description, created_at, updated_at
	`, space.UserID, space.Name, space.Description, space.UpdatedAt, id).Scan(
		&space.ID,
		&space.UserID,
		&space.Name,
		&space.Description,
		&space.CreatedAt,
		&space.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Space{}, ErrNotFound
	}
	if err != nil {
		return models.Space{}, err
	}

	return space, nil
}

func (r *SpaceRepository) Delete(id string) error {
	ctx := context.Background()

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM spaces WHERE id = $1
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
// This repository persists Spaces in PostgreSQL.
// A Space belongs to a user through user_id, and the database foreign key protects that relationship.
// Keeping SQL here means services can focus on Memora rules instead of database syntax.
