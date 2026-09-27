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
  ],
  "input_type": "document"
}
```

Rules:

- `texts` is required.
- Batch size must be 1 through 128.
- Every text item must contain non-whitespace content.
- `input_type` is `document` or `query`; it defaults to `document`.
- Go sends already-created chunks; Python does not chunk content.
- The frontend must never call this endpoint directly.

## Success Response

```json
{
  "success": true,
  "model": "jinaai/jina-embeddings-v5-text-nano",
  "dimension": 256,
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
- Go defaults to requiring `dimension` to equal `256`.
- `model` identifies the embedding model used for stored chunks.
- `document` and `query` select Jina's retrieval passage and query prompts.

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

- Default model: `jinaai/jina-embeddings-v5-text-nano`.
- Output dimension: `256`.
- `document` and `query` select Jina's retrieval document and query prompts.
- The model supports multilingual retrieval and Hindi. Romanized or mixed-script
  Hinglish should be evaluated with representative queries before relying on it.
- Local inference uses CPU/RAM and downloads model files on first use.
- The weights are licensed CC BY-NC 4.0; obtain appropriate permission before
  commercial use.
- Active Jina vectors use `vector(256)`. Legacy 384-dimensional vectors remain
  in `embedding` during migration and are excluded from Jina semantic search.
- Run `go run ./cmd/reembed` from `backend/` to populate vectors for old chunks.
  The command is safe to resume after interruption.

Development fallback:

```text
hash-dev-256
```

This exists only so local endpoint tests can run when the real model dependency
or download is unavailable. It is deterministic, but it is not real semantic
search quality.

## Persistence

Go stores active Jina vectors in:

```text
content_chunks.jina_embedding vector(256)
```

The previous `content_chunks.embedding vector(384)` column is retained as
legacy data during the transition.

alongside:

- `content_id`
- `chunk_index`
- `page_index`
- `start_seconds`
- `end_seconds`
- `metadata`
- `embedding_model`
- `embedded_at`
