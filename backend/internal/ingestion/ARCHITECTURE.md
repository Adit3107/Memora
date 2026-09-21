# Memora Phase 4 Ingestion Architecture

Memora uses Go as the primary backend and a Python service for specialized extraction work.

The goal is not to move the backend to Python. The goal is to keep ownership clear:

- Go receives frontend requests.
- Go validates input and detects content type.
- Go decides which extraction path to use.
- Go owns PostgreSQL, S3, ingestion status, chunking, and API responses.
- Python handles extraction tasks where Python libraries are a better fit.

## Request Flow

```text
Next.js
  -> Go Backend
  -> content type detection
  -> local Go extractor OR Python extraction service
  -> Go cleaner
  -> Go chunker
  -> PostgreSQL
  -> Go API response
```

The frontend should call only Go. Python is an internal service.

## Go Responsibilities

Go remains the application backend. It owns:

- HTTP API routes
- request validation
- URL parsing
- content type detection
- ingestion orchestration
- deciding whether Python is needed
- database writes
- S3 coordination
- ingestion status transitions
- cleaning and normalization when generic
- chunking
- frontend response shape

Go should not call Python just because Python exists. If Go can handle the source cleanly, keep it in Go.

## Python Responsibilities

Python is a specialized extraction service. It should handle:

- PDF extraction
- DOCX extraction
- PPTX extraction
- OCR through Tesseract
- YouTube transcript retrieval through youtube-transcript-api
- future AI utilities in later phases

Python should not own:

- user-facing API routes
- authentication
- PostgreSQL writes
- S3 ownership
- ingestion status
- chunk persistence

## Current Phase 4 Split

Prefer Go for:

- content type detection
- web/article extraction
- YouTube URL validation and orchestration
- transcript normalization
- TXT extraction
- CSV extraction
- basic XLSX extraction when simple sheet/cell text is enough
- cleaning
- chunking
- persistence

Prefer Python for:

- image OCR
- richer PDF extraction
- richer DOCX extraction
- richer PPTX extraction
- YouTube transcript retrieval

TXT and CSV can stay in Go because the standard library handles them simply.

XLSX stays in Go for the current basic sheet-text extraction. Python may be used later if richer spreadsheet handling becomes important.

## Service Boundary

Python should return normalized extraction data to Go. The response should be stable across extractor types.

Conceptual response:

```json
{
  "success": true,
  "content_type": "pdf",
  "title": "Example",
  "text": "Extracted text",
  "metadata": {},
  "pages": []
}
```

For images:

```json
{
  "success": true,
  "content_type": "image",
  "text": "OCR text",
  "metadata": {
    "image_format": "png"
  }
}
```

Python errors should be structured and safe:

```json
{
  "success": false,
  "error": "could not extract text from PDF"
}
```

Go should convert Python errors into Memora application errors. Do not expose Python stack traces to the frontend.

## When Go Should Call Python

Go should call Python when:

- the content type is an image that needs OCR
- a document format needs richer parsing than the current Go extractor provides
- the source is a YouTube URL and transcript retrieval uses youtube-transcript-api
- a future Phase 5+ AI operation requires Python libraries

Go should not call Python when:

- the source is a normal web article and Go can parse useful HTML
- the source is TXT, CSV, or basic XLSX
- the work is persistence, status tracking, or chunking

## Responsibility Table

| Responsibility | Go | Python |
|---|---|---|
| REST API | yes | no |
| Main backend | yes | no |
| Users, spaces, content, tags | yes | no |
| Authentication and authorization integration | yes | no |
| PostgreSQL and migrations | yes | no |
| S3 ownership | yes | no |
| Ingestion orchestration | yes | no |
| URL detection | yes | no |
| YouTube URL detection | yes | no |
| YouTube transcript normalization | yes | retrieval only |
| YouTube transcript retrieval | orchestration only | yes, through youtube-transcript-api |
| Web/article ingestion | yes | optional only with a specific technical reason |
| PDF extraction | no, except fallback/basic tests | yes |
| DOCX extraction | no, except fallback/basic tests | yes |
| PPTX extraction | no, except fallback/basic tests | yes |
| TXT extraction | yes | no |
| CSV extraction | yes | optional only if requirements become data-heavy |
| XLSX extraction | yes for basic sheet text | optional for richer spreadsheet processing |
| OCR | no | yes |
| Cleaning | yes | extraction-specific cleanup only |
| Chunking | yes | no |
| Embeddings | no, Phase 5 | yes, Phase 5 |
| Gemini/LLM/RAG | no, except API orchestration later | yes, later phases |
| Search API | yes | no |

## Node Analogy

If this were a Node app, Go is like the Express or NestJS backend:

```text
controller -> service -> repository -> database
```

Python is like a private internal microservice:

```text
documentExtractorClient.extractPDF(file)
ocrClient.extractImageText(file)
```

The Node/Go backend still owns the user request and database transaction. The Python service only returns extracted content.

## Phase Boundary

Phase 4 ends at structured, cleaned, chunked content.

Do not add:

- embeddings
- Gemini
- LangChain
- semantic search
- RAG
- queues
- background workers
- direct Python database access
