# Memora Backend

Go + Gin backend for Memora.

## Environment

Create a local `.env` file from `.env.example`:

```env
PORT=8080
DATABASE_URL=postgresql://USER:PASSWORD@HOST.neon.tech/DBNAME?sslmode=require
```

Use the Neon PostgreSQL connection string for `DATABASE_URL`. Keep `.env` out of
Git because it contains secrets.

Phase 3 uses PostgreSQL for structured application data. Original uploaded
files will later use object storage such as S3, not PostgreSQL.

On startup, the backend:

1. Loads `.env`
2. Opens a PostgreSQL connection using `DATABASE_URL`
3. Pings the database
4. Applies embedded SQL migrations from `internal/database/migrations`

## Run

```bash
go run ./cmd/server
```

## Current Persistence

- Users are persisted in PostgreSQL.
- Spaces are persisted in PostgreSQL.
- Content, tags, and content-tag links are still temporary in-memory data until
  the next Phase 3 persistence jobs.
