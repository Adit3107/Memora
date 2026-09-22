# Memora Embedding Contract

The Go backend calls the Python AI service over internal HTTP. The frontend
must not call this endpoint directly.

## Endpoint

```http
POST /embeddings
Content-Type: application/json
```

## Request

```json
{
  "texts": [
    "first chunk",
    "second chunk"
  ]
}
```

Rules:

- `texts` is required.
- Batch size must be 1 through 128.
- Every text item must contain non-whitespace content.
- Go sends already-created chunks; Python does not chunk content.
- The frontend must never call this endpoint directly.

## Success Response

```json
{
  "success": true,
  "model": "sentence-transformers/all-MiniLM-L6-v2",
  "dimension": 384,
  "embeddings": [
    [0.12, -0.03],
    [0.04, 0.19]
  ],
  "error": ""
}
```

Rules:

- `embeddings.length` must equal `texts.length`.
- Every embedding must have exactly `dimension` values.
- For V1, Go defaults to requiring `dimension` to equal `384`.
- `model` identifies the embedding model used for stored chunks.

## Failure Response

Application-level failures use:

```json
{
  "success": false,
  "model": "",
  "dimension": 0,
  "embeddings": [],
  "error": "embedding model is unavailable"
}
```

FastAPI/Pydantic validation errors may return HTTP `422` for malformed input.
Go treats non-2xx responses as embedding failures.

## Go Validation

The Go client validates:

- Python service returned success.
- number of vectors equals number of input texts.
- `dimension` equals the backend `EMBEDDING_DIMENSION`.
- every vector length equals the backend `EMBEDDING_DIMENSION`.
- `model` is non-empty.

If validation fails, Go treats the embedding step as failed and does not mark
ingestion completed.

## Model Choice

Default model:

```text
sentence-transformers/all-MiniLM-L6-v2
```

Why:

- Free local model with no paid API requirement.
- Small enough for student/dev machines compared with larger embedding models.
- Produces 384-dimensional vectors, which keeps pgvector storage modest.
- Good enough for V1 semantic search over notes, transcripts, and documents.

Tradeoffs:

- Local inference uses CPU/RAM and the first run downloads model files.
- Quality is good for V1, but paid or larger models may improve retrieval later.
- pgvector schema uses `vector(384)`, so changing to a different dimension later
  requires a migration and re-embedding existing chunks.

If the model changes later, update both services:

- Python: `EMBEDDING_MODEL_NAME` and `EMBEDDING_DIMENSION`
- Go: `EMBEDDING_DIMENSION`
- PostgreSQL: a migration that matches the new vector dimension

Development fallback:

```text
hash-dev-384
```

This exists only so local endpoint tests can run when the real model dependency
or download is unavailable. It is deterministic, but it is not real semantic
search quality.

## Persistence

Go stores vectors in:

```text
content_chunks.embedding vector(384)
```

alongside:

- `content_id`
- `chunk_index`
- `page_index`
- `start_seconds`
- `end_seconds`
- `metadata`
- `embedding_model`
- `embedded_at`
