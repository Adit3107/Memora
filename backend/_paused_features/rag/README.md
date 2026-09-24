# Paused Feature: RAG Service (Retrieval-Augmented Generation)

> **Status:** PAUSED / ON HOLD  
> **Target Release:** Upcoming Phase

This directory isolates all RAG and LLM backend components so the active codebase only exposes working and ready-for-use features (Content Ingestion, pgvector Embedding storage, Spaces, Tags, and Hybrid Search).

## Contained Files

- **`rag_service.go`**: Core RAG orchestration logic including intent classification, context window formatting, grounding prompt design, citation assembly with timestamps and page numbers, and grounded fallback logic.
- **`rag_service_test.go`**: Unit test suite for RAG service.
- **`llm_service.go`**: Google Gemini API client integration (`GeminiLLMService`).
- **`rag_handler.go`**: HTTP controller for `POST /api/rag`.

## How to Resume & Re-enable

1. Move `rag_service.go`, `rag_service_test.go`, and `llm_service.go` back to `backend/internal/services/`.
2. Move `rag_handler.go` back to `backend/internal/handlers/`.
3. Uncomment the file contents (remove the outer `/* ... */` comment blocks).
4. In `backend/internal/routes/routes.go`:
   - Uncomment the initialization of `ragHandler := handlers.NewRAGHandler(...)`.
   - Uncomment the `api.POST("/rag", ragHandler.Ask)` route registration.
5. Run verification tests:
   ```powershell
   go test ./...
   ```
