# Memora Backend

Go + Gin backend for Memora.

Go is the primary application backend. It owns public HTTP APIs, users, spaces,
content metadata, tags, ingestion orchestration, PostgreSQL, pgvector, S3,
processing status, chunk persistence, and responses to the frontend.

## Environment

Create a local `.env` file from `.env.example`:

```env
PORT=8080
DATABASE_URL=postgresql://USER:PASSWORD@HOST.neon.tech/DBNAME?sslmode=require

AWS_REGION=ap-south-1
AWS_S3_BUCKET=memora-original-files
AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY
AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY

AI_SERVICE_URL=http://localhost:8001
EMBEDDING_DIMENSION=384
EMBEDDING_MAX_CONCURRENCY=2
```

Use the Neon PostgreSQL connection string for `DATABASE_URL`. Keep `.env` out of
Git because it contains secrets, including AWS access keys.

`AI_SERVICE_URL` points to the internal FastAPI service. The frontend should not
use this URL.

`EMBEDDING_DIMENSION` must match both the Python embedding model output and the
pgvector column dimension in the active migration.

`EMBEDDING_MAX_CONCURRENCY` controls how many embedding batches Go may send to
Python at the same time. Keep this small for local development so the Python
model is not overloaded.

## Startup

On startup, the backend:

1. Loads `.env`
2. Opens a PostgreSQL connection using `DATABASE_URL`
3. Pings the database
4. Applies embedded SQL migrations from `internal/database/migrations`

Run:

```powershell
go run ./cmd/server
```

## Phase 5 Ingestion Flow

```text
Frontend
  -> Go /api/ingestion/*
  -> content type detection
  -> extraction in Go or Python
  -> Go cleaning
  -> Go chunking
  -> Go calls Python /embeddings
  -> Go validates vector shape
  -> PostgreSQL + pgvector
```

For YouTube:

```text
YouTube URL
  -> Go validates URL and extracts video ID
  -> Python /extract/youtube
  -> timestamped transcript
  -> Go chunks by transcript segments
  -> Python /embeddings
  -> content_chunks.embedding vector(384)
```

For documents:

```text
PDF/DOCX/PPTX/TXT/CSV/XLSX/Image
  -> extraction or OCR
  -> cleaned text
  -> page-aware chunks where page data exists
  -> Python /embeddings
  -> content_chunks.embedding vector(384)
```

## pgvector Storage

Migration `005_content_chunk_embeddings.sql` adds:

- `content_chunks.embedding vector(384)`
- `content_chunks.embedding_model`
- `content_chunks.embedded_at`

The dimension is `384` because the selected model is
`sentence-transformers/all-MiniLM-L6-v2`.

If you change the embedding model later, also update `EMBEDDING_DIMENSION`,
create a matching pgvector migration, and re-embed existing chunks.

Chunk metadata is preserved for future citations:

- `content_id`
- `chunk_index`
- `page_index`
- `start_seconds`
- `end_seconds`
- `metadata`

## Failure Handling

Ingestion uses the existing state machine:

```text
pending -> processing -> completed
                      -> failed
```

If extraction, cleaning, chunking, embedding generation, vector validation, or
database persistence fails, Go marks the ingestion result as `failed`.

Go does not mark ingestion `completed` unless every chunk has:

- non-empty text
- a `384`-dimension embedding
- an embedding model name

Database writes for the final ingestion result and chunks happen in one
transaction. That keeps partial embedding persistence from corrupting completed
content.

## Testing

```powershell
go test ./...
```

Manual checks:

1. Start the AI service on port `8001`.
2. Start the Go backend on port `8080`.
3. Ingest a YouTube URL with `POST /api/ingestion/url`.
4. Ingest a document with `POST /api/ingestion/file`.
5. Verify `content_chunks.embedding IS NOT NULL` in PostgreSQL.

## Phase Boundary

Phase 5 stores vectorized knowledge. Search ranking, hybrid search, RAG, chat,
summaries, recommendation systems, queues, and additional social ingestion
adapters belong to later phases.
