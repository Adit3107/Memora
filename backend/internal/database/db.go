package database

import (
	"context" // Context is used for timeout on pinging the database.
	"database/sql" // database/sql is the standard library's generic SQL interface.
	"errors" // errors is used to create error values.
	"time" // time is used for setting connection pool lifetimes and ping timeouts.

	// Blank import registers the pgx driver with database/sql.
	// We do not call pgx directly here; database/sql discovers it by driver name.
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(databaseURL string) (*sql.DB, error) { // Open creates a new database connection pool to the given database URL.
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	// sql.Open creates a connection pool handle; it does not prove the DB is reachable yet.
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	// These settings keep the pool small for development and avoid opening too many Neon connections.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func Ping(ctx context.Context, db *sql.DB) error {
	// Context timeout prevents startup from hanging forever if Neon/network is unavailable.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return db.PingContext(pingCtx)
}

// Why this file exists:
// This file owns the low-level PostgreSQL connection setup for Memora.
// The rest of the backend should not know driver names, pool settings, or ping logic.
// database/sql gives us a reusable connection pool, while pgx is the actual Postgres driver.
