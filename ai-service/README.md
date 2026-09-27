# Memora AI Service

This FastAPI service is the Python side of Memora's internal AI architecture.

Go remains the primary backend. Python is used for work where Python libraries
are the better fit: document extraction, image OCR, YouTube transcript
retrieval, and embedding generation.

The frontend should never call this service directly.

## Local Setup

```powershell
cd ai-service
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8001
```

The first real embedding request downloads the Jina model weights. After that,
the model is cached by the normal Hugging Face tooling.

## Environment

```env
HOST=127.0.0.1
PORT=8001
TESSERACT_CMD=tesseract
EMBEDDING_MODEL_NAME=jinaai/jina-embeddings-v5-text-nano
EMBEDDING_DIMENSION=256
ALLOW_HASH_EMBEDDINGS=true
```

`ALLOW_HASH_EMBEDDINGS=true` exists for local development only. It lets endpoint
tests keep working if the real model dependency or download is unavailable. The
hash fallback is deterministic, but it is not real semantic search quality.

## Endpoints

```http
GET /health
POST /extract/document
POST /extract/image
POST /extract/youtube
POST /embeddings
```

## Embedding Model

Default model:

```text
jinaai/jina-embeddings-v5-text-nano
```

Dimension:

```text
256
```

Why this model:

- Local inference supports multilingual retrieval, including Hindi.
- Produces 256-dimensional vectors using Jina's supported Matryoshka size.
- Query and document inputs use distinct retrieval prompts.

The model weights are licensed CC BY-NC 4.0. Check the license and obtain
permission before using the weights commercially.

Python returns vectors to Go. It does not connect to PostgreSQL and does not own
chunk persistence.

## Embedding Contract

The contract is documented in `EMBEDDING_CONTRACT.md`.

Request:

```json
{
  "texts": ["first chunk", "second chunk"],
  "input_type": "document"
}
```

Response:

```json
{
  "success": true,
  "model": "jinaai/jina-embeddings-v5-text-nano",
  "dimension": 256,
  "embeddings": [[0.12, -0.03]],
  "error": ""
}
```

Rules:

- batch size is 1 through 128
- empty text is rejected
- one vector is returned for each input text
- `input_type` is `document` or `query` (defaults to `document`)
- vectors must match the configured dimension

## Testing

```powershell
..\ai-service\.venv\Scripts\python.exe -m unittest discover -s tests
```

Manual Postman check:

```http
POST http://localhost:8001/embeddings
Content-Type: application/json
```

```json
{
  "texts": [
    "I am learning Go backend development.",
    "Memora stores chunks in pgvector."
  ]
}
```

Expected: `success: true`, `dimension: 256`, and two embedding arrays.

## Phase Boundary

This service does not own users, auth, PostgreSQL, S3, frontend APIs, chunk
persistence, search ranking, or RAG. Go orchestrates those application concerns.
