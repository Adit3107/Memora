# Memora Backend

Go + Gin backend for Memora.

## Environment

Create a local `.env` file from `.env.example`:

```env
PORT=8080
DATABASE_URL=postgresql://USER:PASSWORD@HOST.neon.tech/DBNAME?sslmode=require
AWS_REGION=ap-south-1
AWS_S3_BUCKET=memora-original-files
AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY
AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
```

Use the Neon PostgreSQL connection string for `DATABASE_URL`. Keep `.env` out of
Git because it contains secrets, including AWS access keys.

Phase 3 uses PostgreSQL for structured application data. Original uploaded
files use object storage such as S3, not PostgreSQL.

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
- Content metadata is persisted in PostgreSQL.
- Tags and content-tag links are persisted in PostgreSQL.
- Original binary files should be stored in S3 through `internal/storage`.

## S3 Notes

S3 is object storage: a bucket stores objects, and each object is addressed by a
key. Memora keeps metadata in PostgreSQL and stores original files like PDFs,
DOCX files, CSVs, and images in S3.

Use a private bucket. Do not enable public access unless a later product
requirement needs it. The IAM user or role should have only the minimum actions
needed for this backend:

- `s3:PutObject`
- `s3:GetObject`
- `s3:DeleteObject`

Limit those actions to the Memora bucket ARN, not every bucket in the AWS
account.
