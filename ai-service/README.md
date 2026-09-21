# Memora AI / Extraction Service

This service is the Python side of Memora's Phase 4 ingestion architecture.

Go remains the primary backend. This FastAPI service only handles extraction tasks where Python's ecosystem is a better fit, such as PDF/DOCX/PPTX parsing and OCR. Go should keep simple extraction work such as TXT, CSV, and current basic XLSX handling when that is enough.

## Local Setup

```powershell
cd ai-service
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8001
```

## Endpoints

```http
GET /health
POST /extract/document
POST /extract/image
```

The Go backend calls this service. The frontend should not call it directly.

## Phase Boundary

This service does not generate embeddings, call Gemini, run RAG, or connect to PostgreSQL. Those responsibilities belong to later phases or to the Go backend.
