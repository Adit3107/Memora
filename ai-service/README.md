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

The first real embedding request may download the local SentenceTransformer
model. After that, the model is cached by the normal Hugging Face tooling.

## Environment

```env
HOST=127.0.0.1
PORT=8001
TESSERACT_CMD=tesseract
EMBEDDING_MODEL_NAME=sentence-transformers/all-MiniLM-L6-v2
EMBEDDING_DIMENSION=384
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
sentence-transformers/all-MiniLM-L6-v2
```

Dimension:

```text
384
```

Why this model:

- Free and local, so no paid API is required.
- Small enough for a student/dev machine.
- Good enough for V1 semantic retrieval over documents and transcripts.
- Produces 384-dimensional vectors, which keeps pgvector storage modest.

Python returns vectors to Go. It does not connect to PostgreSQL and does not own
chunk persistence.

## Embedding Contract

The contract is documented in `EMBEDDING_CONTRACT.md`.

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

Rules:

- batch size is 1 through 128
- empty text is rejected
- one vector is returned for each input text
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

Expected: `success: true`, `dimension: 384`, and two embedding arrays.

## Phase Boundary

This service does not own users, auth, PostgreSQL, S3, frontend APIs, chunk
persistence, search ranking, or RAG. Go orchestrates those application concerns.
