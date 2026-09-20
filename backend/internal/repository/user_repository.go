package repository

import (
	"context"
	"database/sql"
	"errors"

	"memora-backend/internal/models"
)

type UserRepository struct {
	// *sql.DB is a concurrency-safe connection pool, not one single connection.
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user models.User) (models.User, error) {
	ctx := context.Background()

	// RETURNING lets Postgres send back generated/default fields such as id.
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (name, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, created_at, updated_at
	`, user.Name, user.Email, user.CreatedAt, user.UpdatedAt).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *UserRepository) List() ([]models.User, error) {
	ctx := context.Background()

	// QueryContext returns rows because SELECT can return many users.
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, email, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		// Scan copies the current SQL row into the Go struct fields.
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			continue
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) GetByID(id string) (models.User, error) {
	ctx := context.Background()

	var user models.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Translate database-specific "no row" into our app-level not found error.
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *UserRepository) Update(id string, user models.User) (models.User, error) {
	ctx := context.Background()

	err := r.db.QueryRowContext(ctx, `
		UPDATE users
		SET name = $1, email = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, name, email, created_at, updated_at
	`, user.Name, user.Email, user.UpdatedAt, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *UserRepository) Delete(id string) error {
	ctx := context.Background()

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM users WHERE id = $1
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
// This repository is the only place that knows SQL details for users.
// Handlers/services call methods like Create and GetByID instead of writing SQL.
// Parameter placeholders ($1, $2, ...) protect us from SQL injection by separating SQL from values.
