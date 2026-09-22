# Memora Postman Collection

Import these two files into Postman:

- `memora-phase-5.postman_collection.json`
- `memora-local.postman_environment.json`

Select the `Memora Local` environment before running requests.

## Required Servers

Start the AI service:

```powershell
cd ai-service
.\.venv\Scripts\Activate.ps1
uvicorn app.main:app --reload --port 8001
```

Start the Go backend:

```powershell
cd backend
go run ./cmd/server
```

## Suggested Order

1. Backend - Setup / Health Check
2. AI Service - Direct Tests / AI Health Check
3. Backend - Setup / Create User
4. Backend - Setup / Create Space
5. AI Service - Direct Tests / Generate Embeddings
6. Backend - Ingestion / Ingest YouTube URL
7. Backend - Ingestion / Get Ingestion Result
8. Backend - Ingestion / Ingest File

`Create User`, `Create Space`, and ingestion requests save returned IDs into the
environment automatically.

## File Upload Variables

Set these environment variables manually before file upload requests:

- `document_file_path`
- `image_file_path`

Use absolute paths if the Postman extension does not resolve relative paths.

## Phase Boundary

The collection includes `Backend - Optional Search Route` because the route
exists in the current codebase. Treat it as a Phase 6 preview. Phase 5 is about
creating and storing vectorized chunks, not building search or RAG UX.
