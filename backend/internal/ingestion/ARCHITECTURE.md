# Memora Ingestion and Embedding Architecture

Memora uses Go as the primary backend and a Python service for specialized
extraction and AI work.

The goal is not to move the backend to Python. The goal is to keep ownership
clear and make the system easy to explain.

## System Diagram

```text
Next.js
   |
   v
Go API
   |
   +------ PostgreSQL + pgvector
   |
   +------ S3
   |
   +------ Python AI Service
              |
              +-- Embeddings
              +-- OCR/processing support
              +-- Future AI services
```

The frontend calls only Go. Python is an internal service.

```text
Next.js -> Go -> Python
```

## Phase 5 Flow

```text
Content
  -> Extract
  -> Clean
  -> Chunk
  -> Go calls Python /embeddings
  -> Python returns vectors
  -> Go validates vectors
  -> Go persists chunks + vectors
```

For YouTube:

```text
YouTube URL
  -> transcript
  -> timestamped chunks
  -> embeddings
  -> pgvector
```

For documents:

```text
Document
  -> text extraction or OCR
  -> page-aware chunks where page data exists
  -> embeddings
  -> pgvector
```

## Go Responsibilities

Go remains the application backend. It owns:

- HTTP API routes
- request validation
- URL parsing
- content type detection
- ingestion orchestration
- users, spaces, content, and tags
- PostgreSQL migrations and writes
- S3 coordination
- ingestion status transitions
- generic cleaning and normalization
- chunking
- embedding client calls to Python
- vector shape validation
- chunk and vector persistence
- frontend response shape

Go should not call Python just because Python exists. If Go can handle a source
cleanly, keep it in Go.

## Python Responsibilities

Python is a specialized internal AI service. It handles:

- richer PDF extraction
- DOCX extraction
- PPTX extraction
- image OCR through Tesseract
- YouTube transcript retrieval through youtube-transcript-api
- embedding generation through SentenceTransformers

Python does not own:

- user-facing API routes
- authentication
- PostgreSQL writes
- S3 ownership
- ingestion status
- chunk persistence

## Internal API Contracts

Extraction endpoints return structured text, pages, transcript segments, and
metadata.

Embedding endpoint:

```http
POST /embeddings
```

Request:

```json
{
  "texts": ["first chunk", "second chunk"]
}
```

Response:

```json
{
  "success": true,
  "model": "sentence-transformers/all-MiniLM-L6-v2",
  "dimension": 384,
  "embeddings": [[0.12, -0.03]],
  "error": ""
}
```

The full embedding contract lives in `ai-service/EMBEDDING_CONTRACT.md`.

## Embedding Model

Default model:

```text
sentence-transformers/all-MiniLM-L6-v2
```

Dimension:

```text
384
```

This model is local and free to run, which fits the project goal of being
student-friendly. The tradeoff is that the first run downloads model files and
local CPU inference is slower than paid hosted APIs.

Changing the embedding dimension later requires a database migration and
re-embedding existing chunks.

The Go backend reads `EMBEDDING_DIMENSION` from the environment and uses it to
validate Python responses before writing chunks. This makes model swaps explicit:
change the Python model, change the expected dimension, migrate pgvector, then
re-embed existing content.

## pgvector Schema

Phase 5 keeps the existing content/chunk model and adds only what embeddings
need:

```text
content
   |
   +-- content_chunks
          |
          +-- text
          +-- chunk_index
          +-- page_index
          +-- start_seconds
          +-- end_seconds
          +-- metadata
          +-- embedding vector(384)
          +-- embedding_model
          +-- embedded_at
```

The vector index uses cosine distance through pgvector.

## Concurrency

The ingestion pipeline remains sequential through extraction, cleaning, and
chunking because each stage depends on the previous stage.

Embedding generation can be batched and parallelized because each batch of chunk
texts is independent. Go uses a bounded worker pool controlled by
`EMBEDDING_MAX_CONCURRENCY`.

Database writes remain sequential and transactional. This avoids races and
prevents partially embedded chunks from being marked completed.

Context cancellation is passed through Go HTTP requests to Python. If the client
cancels or a timeout occurs, workers stop and ingestion is marked failed.

## Failure Handling

Ingestion states:

```text
pending -> processing -> completed
                      -> failed
```

V1 behavior is simple:

- extraction failure marks ingestion failed
- empty content marks ingestion failed
- embedding service failure marks ingestion failed
- invalid embedding dimensions mark ingestion failed
- database failure marks ingestion failed when possible

Go does not save a completed ingestion unless every chunk has a valid vector and
model name.

## Tests

Backend tests cover:

- chunking
- extraction helpers
- Python embedding client success and invalid responses
- chunk-to-embedding integration
- bounded embedding concurrency
- validation that completed chunks require vectors

AI service tests cover:

- YouTube transcript behavior
- embedding fallback behavior
- empty text handling

Run:

```powershell
cd backend
go test ./...
```

```powershell
cd ai-service
..\ai-service\.venv\Scripts\python.exe -m unittest discover -s tests
```

## Phase Boundary

Phase 5 ends when extracted content can be chunked, embedded, and stored in
PostgreSQL + pgvector.

Do not add these to Phase 5:

- semantic search UI
- search ranking
- hybrid search
- RAG
- chat
- Redis
- Kafka
- background distributed workers
- additional social platform ingestion
- recommendations
- AI-generated tags
- summaries
- production Kubernetes deployment
