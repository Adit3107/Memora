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

## YouTube Transcript Ingestion

The current video ingestion path is intentionally narrow:

```text
YouTube URL
-> Go validates URL and extracts video ID
-> Python /extract/youtube
-> youtube-transcript-api
-> Go cleaning/chunking/persistence
```

Restart `go run ./cmd/server`, then send **POST**
`http://localhost:8080/api/ingestion/url` in Postman, Body -> raw -> JSON:

```json
{
  "user_id": "67a79aff-376d-48a7-af69-f087d46d313e",
  "space_id": "2e96706f-8134-448b-918d-979aeb0500bc",
  "url": "https://youtu.be/nlz9j-r0U9U"
}
```

Use the returned `ingestion_id` with `GET /api/ingestion/:id` to inspect
`clean_text`, `transcript`, and metadata.

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
