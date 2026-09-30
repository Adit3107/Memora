# 🧠 Mindshelf

<div align="center">

**Your Intelligent Multimodal Second Memory.**  
*Capture YouTube videos, Instagram Reels, PDFs, Office documents, articles, and images — transcribe, chunk, embed, and search with timestamp and page-level precision.*

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![FastAPI](https://img.shields.io/badge/FastAPI-0.118-009688?style=for-the-badge&logo=fastapi&logoColor=white)](https://fastapi.tiangolo.com)
[![Next.js](https://img.shields.io/badge/Next.js-16.3.5-000000?style=for-the-badge&logo=next.js&logoColor=white)](https://nextjs.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-pgvector-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)](https://github.com/pgvector/pgvector)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://docker.com)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](#)

[🌐 Live Demo (mindshelf.in)](https://mindshelf.in) • [🚀 Quick Start](#-quick-start) • [📐 Architecture](#-system-architecture) • [🤖 AI Pipeline](#-ai-models--pipeline) • [📡 API Reference](#-interactive-api-reference)

</div>

---

## 📑 Table of Contents

- [🌟 Overview](#-overview)
- [✨ Key Features](#-key-features)
- [📐 System Architecture](#-system-architecture)
- [🤖 AI Models & Pipeline](#-ai-models--pipeline)
- [🔍 Hybrid Search Engine (RRF)](#-hybrid-search-engine-rrf)
- [🛠️ Technology Stack](#-technology-stack)
- [🗄️ Database Schema & Vectors](#️-database-schema--vector-architecture)
- [📡 Interactive API Reference](#-interactive-api-reference)
- [🚀 Quick Start](#-quick-start)
- [⚙️ Environment Configuration](#️-environment-configuration)
- [🧪 Testing & Verification](#-testing--verification)
- [🌐 Production Deployment & Nginx](#-production-deployment--nginx-reverse-proxy)
- [📂 Project Structure](#-project-structure)

---

## 🌟 Overview

Every day, developers, students, and researchers consume dozens of YouTube tutorials, technical PDFs, documentation sites, and short-form video clips. Traditional bookmarking falls short:
- **Audio & Video are black boxes**: You can't `Ctrl + F` across 40 minutes of a conference talk or a fast reel.
- **Documents get siloed**: Key ideas in slide 32 of a presentation or page 14 of a research paper remain forgotten.
- **Keyword matching is brittle**: Searching for *"goroutine memory leakage"* misses documents talking about *"channel cleanup and hanging routines"*.

**Mindshelf solves this by serving as your external cognitive shelf.** You save any URL or upload any file. Behind the scenes, an orchestrated pipeline of **Go**, **FastAPI**, **Whisper**, **OCR**, and **Jina Embeddings** extracts the substance, splits it into semantic chunks with exact temporal and spatial anchors, and indexes it into **PostgreSQL + pgvector**. You can then search across everything via **Hybrid Search (Semantic + Keyword + Reciprocal Rank Fusion)** and jump straight to the exact second or page.

---

## ✨ Key Features

| Capability | What Mindshelf Does |
| :--- | :--- |
| **Multimodal Ingestion** | Ingests **YouTube videos**, **Instagram Reels**, **PDFs**, **DOCX**, **PPTX**, **TXT**, **CSV**, **XLSX**, and **Images (PNG, JPEG, WebP)**. |
| **Speech-to-Text & Transcripts** | Extracts official YouTube captions or transcribes raw audio using **Faster-Whisper** with millisecond timestamps. |
| **Vision & Keyframe OCR** | Extracts text from video frames via **RapidOCR** (ONNX-accelerated) and static image diagrams via **Tesseract OCR**. |
| **Asymmetric Embeddings** | Generates 256-dimensional Matryoshka embeddings via **Jina v5 Nano**, using specialized retrieval task prefixes for documents vs. queries. |
| **3-Tier Search Retrieval** | Choose between **Semantic Vector Search** (cosine similarity), **Full-Text Keyword Search** (`tsvector`), or **Hybrid Search** with **Reciprocal Rank Fusion (RRF)**. |
| **AI Summarization** | Instant, high-signal bullet-point takeaways powered by **Google Gemini (1.5 Flash)**. |
| **Spaces & Knowledge Taxonomies** | Group content into dedicated topic rooms (*System Design*, *Go*, *AI Research*, *DSA*) with custom tags. |
| **Enterprise Authentication** | Frictionless authentication via **Clerk**, automatically synced to internal PostgreSQL user models. |
| **Modern Responsive UI** | Built with **Next.js 16 (Turbopack)**, **Tailwind CSS v4**, rich dark/light modes, keyboard shortcuts (`Ctrl + K`), and fluid micro-animations. |

---

## 📐 System Architecture

Mindshelf enforces a strict separation of concerns across a modern microservices architecture:

```mermaid
graph TD
    Client["Client Browser<br/>(Next.js 16 + Tailwind CSS)"]
    
    subgraph Gateway ["Application Gateway & Core Backend (Go + Gin)"]
        GoAPI["Go HTTP Server (:8080)"]
        Router["Route Handlers & Auth Sync"]
        IngestManager["Ingestion Orchestrator"]
        Chunker["Semantic & Timestamp Chunker"]
        SearchEngine["Hybrid Search Service (RRF)"]
        GeminiClient["Gemini 1.5 Flash Client"]
    end
    
    subgraph AIService ["Internal AI Microservice (Python + FastAPI)"]
        PyAPI["FastAPI Server (:8001)"]
        Whisper["Faster-Whisper & Transcript API"]
        RapidOCR["RapidOCR + OpenCV Keyframes"]
        DocExtract["PyPDF / python-docx / openpyxl"]
        JinaModel["Jina Embeddings v5 Nano (256d)"]
    end
    
    subgraph DataStorage ["Data & Persistence Layer"]
        PG[("Neon PostgreSQL<br/>+ pgvector")]
        S3[("AWS S3<br/>Original Files & Media")]
    end

    Client -->|"HTTP / REST (Clerk Auth)"| GoAPI
    GoAPI --> Router
    Router --> IngestManager
    Router --> SearchEngine
    Router --> GeminiClient
    
    IngestManager -->|"Internal HTTP (:8001)"| PyAPI
    SearchEngine -->|"Query Embedding (:8001)"| PyAPI
    
    PyAPI --> Whisper
    PyAPI --> RapidOCR
    PyAPI --> DocExtract
    PyAPI --> JinaModel
    
    IngestManager --> Chunker
    Chunker -->|"Batch Embeddings"| JinaModel
    
    IngestManager -->|"Store Chunks & 256d Vectors"| PG
    IngestManager -->|"Store Raw Blobs"| S3
    SearchEngine -->|"Cosine Distance (<=>) & FTS"| PG
```

### Architectural Guardrails
1. **Frontend calls ONLY Go**: Next.js communicates solely with the Go backend (`:8080`).
2. **Python is strictly internal**: The Python FastAPI service (`:8001`) is never exposed directly to the public web or frontend clients.
3. **Go owns application state**: Go manages users, spaces, database transactions, S3 uploads, ingestion status state machines, and chunk persistence.
4. **Python owns compute-heavy AI**: Python focuses on document extraction, OCR, audio transcription, and vector embedding inference.

---

## 🤖 AI Models & Pipeline

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant Web as Next.js Frontend
    participant Go as Go Backend (:8080)
    participant Py as Python AI Service (:8001)
    participant DB as PostgreSQL + pgvector
    participant S3 as AWS S3

    User->>Web: Submits YouTube URL / Uploads Document
    Web->>Go: POST /api/ingestion/url OR /file
    Go->>Go: Validate input & detect content type
    opt File Upload
        Go->>S3: Upload raw document/asset
    end
    Go->>Py: POST /extract/youtube OR /extract/document
    Py->>Py: Transcribe audio / Run OCR / Parse text
    Py-->>Go: Structured text with timestamps / page numbers
    Go->>Go: Clean, normalize & generate semantic chunks
    Go->>Py: POST /embeddings (Batch chunk texts, input_type="document")
    Py->>Py: Generate 256-dim Jina vectors
    Py-->>Go: Normalized 256-dim embedding vectors
    Go->>Go: Validate vector dimensions (len == 256)
    Go->>DB: Atomic Transaction (Insert content + chunks + vectors)
    Go-->>Web: Ingestion Completed
    Web-->>User: Content indexed & searchable
```

### Model Specifications

| Component | Technology / Model | Dimensions / Specs | Primary Purpose |
| :--- | :--- | :--- | :--- |
| **Embeddings** | `jinaai/jina-embeddings-v5-text-nano` | **256 dimensions** (Matryoshka) | Asymmetric dense text vector embeddings with multilingual support. |
| **Audio Transcription** | `faster-whisper` (CTranslate2) | Small/Base Whisper weights | Local, high-throughput audio transcription with start/end timestamps. |
| **YouTube Captions** | `youtube-transcript-api` | Direct JSON caption stream | Zero-compute, millisecond-accurate transcript extraction. |
| **Keyframe OCR** | `rapidocr-onnxruntime` + OpenCV | ONNX Runtime | Keyframe extraction and visual text recognition for short-form video reels. |
| **Static Document OCR** | `pytesseract` + `Pillow` | Tesseract 5 Engine | High-precision optical character recognition for diagrams and screenshots. |
| **Document Parsers** | `pypdf`, `python-docx`, `python-pptx`, `openpyxl` | Structured page-level AST | Extracts page-tagged text, table rows, and presentation slides. |
| **Video Ingestion** | `yt-dlp` | Audio stream demuxing | Extracts audio streams and video metadata from supported media links. |
| **Summarization** | Google Gemini (`gemini-1.5-flash`) | Context-aware LLM API | Generates executive summaries and key bullet takeaways for indexed items. |

---

## 🔍 Hybrid Search Engine (RRF)

Mindshelf implements a production-grade **Hybrid Retrieval Pipeline** that outperforms standalone vector or keyword search:

```mermaid
graph LR
    Query["Search Query"] --> GoEngine["Go Search Service"]
    
    GoEngine -->|"Query Text"| Keyword["PostgreSQL Full-Text Search<br/>(to_tsvector & websearch_to_tsquery)"]
    GoEngine -->|"Python :8001"| Dense["Jina Embeddings<br/>(Task: retrieval.query)"]
    
    Dense --> VectorSearch["pgvector Cosine Distance<br/>(jina_embedding <=> query_vec)"]
    
    Keyword --> RankA["Rank Keyword Matches (1..N)"]
    VectorSearch --> RankB["Rank Semantic Matches (1..N)"]
    
    RankA --> RRF["Reciprocal Rank Fusion (RRF)<br/>Score = Σ (1 / (60 + Rank))"]
    RankB --> RRF
    
    RRF --> Final["Ranked, Deduplicated Results<br/>with Timestamps & Page Numbers"]
```

### The Math: Reciprocal Rank Fusion
Instead of attempting to calibrate and blend vector distance floats (0.0 to 2.0) with PostgreSQL BM25 text rank floats (arbitrary scales), Mindshelf applies **Reciprocal Rank Fusion (RRF)** with a constant $k = 60$:

$$RRF(d) = \sum_{m \in \{semantic, keyword\}} \frac{1}{60 + \text{rank}_m(d)}$$

- **Robust against outliers**: High-confidence keyword matches (e.g. unique error codes or method names) complement abstract conceptual matches (e.g. *"how does memory cleanup work"*).
- **Zero score normalization drift**: Vector scores and text scores do not need constant manual tuning.

---

## 🛠️ Technology Stack

| Layer | Technology | Key Libraries & Tools |
| :--- | :--- | :--- |
| **Frontend** | Next.js 16.3.5 (App Router, Turbopack) | React 19, TypeScript, Tailwind CSS v4, Lucide React, Framer Motion, `@clerk/nextjs` |
| **Backend Gateway** | Go 1.23+ | Gin Gonic 1.12, `pgx/v5`, AWS SDK for Go v2 (S3), `godotenv`, embedded SQL migrations |
| **AI Microservice** | Python 3.11+ | FastAPI 0.118, Uvicorn, Sentence-Transformers 5.1, Faster-Whisper, RapidOCR, PyPDF, `yt-dlp` |
| **Database** | Neon PostgreSQL 16 | `pgvector` (Vector similarity search), GiST/GIN indexes, Full-Text Search (`tsvector`) |
| **Object Storage** | Amazon AWS S3 | Encrypted bucket storage for raw documents, presentations, and uploaded media assets |
| **Authentication** | Clerk Authentication | Modern session management, JWT verification, and automated background user synchronization |
| **DevOps & Infra** | Docker & Docker Compose | Multi-stage Docker builds, Nginx reverse proxy, Certbot SSL, systemd automation |

---

## 🗄️ Database Schema & Vector Architecture

```mermaid
erDiagram
    users ||--o{ spaces : owns
    users ||--o{ contents : saves
    users ||--o{ tags : creates
    spaces ||--o{ contents : categorizes
    contents ||--o{ content_chunks : splits_into
    contents ||--o{ content_tags : labeled_with
    tags ||--o{ content_tags : applies_to
    contents ||--o{ ingestion_results : tracks

    users {
        uuid id PK
        string email
        string display_name
        string clerk_id UK
        timestamp created_at
        timestamp updated_at
    }

    spaces {
        uuid id PK
        uuid user_id FK
        string name
        string description
        timestamp created_at
    }

    contents {
        uuid id PK
        uuid user_id FK
        uuid space_id FK
        string type "video | document | article | image"
        string source "youtube | reel | pdf | docx | etc"
        string source_url
        string title
        string description
        string s3_bucket
        string s3_key
        string status "pending | processing | completed | failed"
        jsonb metadata
        timestamp created_at
    }

    content_chunks {
        uuid id PK
        uuid content_id FK
        int chunk_index
        int page_index "Null for videos"
        float start_seconds "Null for docs"
        float end_seconds "Null for docs"
        text text
        vector jina_embedding "vector(256)"
        string embedding_model
        timestamp embedded_at
        jsonb metadata
    }

    tags {
        uuid id PK
        uuid user_id FK
        string name
        timestamp created_at
    }

    content_tags {
        uuid content_id PK,FK
        uuid tag_id PK,FK
    }

    ingestion_results {
        uuid id PK
        uuid user_id FK
        uuid content_id FK
        string status "pending | processing | completed | failed"
        text error
        timestamp created_at
    }
```

---

## 📡 Interactive API Reference

<details>
<summary><b>1. Ingestion Endpoints (URLs & Files)</b></summary>

### Ingest URL (YouTube, Reels, Articles)
`POST /api/ingestion/url`

**Request:**
```bash
curl -X POST http://localhost:8080/api/ingestion/url \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "usr_99812401",
    "space_id": "spc_1204812",
    "url": "https://www.youtube.com/watch?v=0ZJgIjIuY7U",
    "title": "Kafka Architecture Crash Course"
  }'
```

**Response (200 OK):**
```json
{
  "content_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "ingestion_id": "c1f73b88-1293-4e2b-bb29-fa8430b8e721",
  "status": "processing",
  "message": "Content accepted for processing"
}
```

---

### Ingest Document (PDF, DOCX, PPTX, XLSX, Images)
`POST /api/ingestion/file`

**Request:**
```bash
curl -X POST http://localhost:8080/api/ingestion/file \
  -F "user_id=usr_99812401" \
  -F "space_id=spc_1204812" \
  -F "title=System Design Guide" \
  -F "file=@/path/to/system_design.pdf"
```

---

### Check Ingestion Status
`GET /api/ingestion/:id`

**Response:**
```json
{
  "id": "c1f73b88-1293-4e2b-bb29-fa8430b8e721",
  "content_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "status": "completed",
  "error": "",
  "created_at": "2026-10-01T02:00:00Z"
}
```
</details>

<details>
<summary><b>2. Search Retrieval Endpoints</b></summary>

### Search Knowledge Base
`POST /api/search`

**Request:**
```bash
curl -X POST http://localhost:8080/api/search \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "usr_99812401",
    "query": "How do consumer group rebalances occur?",
    "mode": "hybrid",
    "limit": 5,
    "offset": 0
  }'
```

**Supported Modes:**
- `hybrid`: Reciprocal Rank Fusion combining semantic similarity and keyword ranking.
- `semantic`: Pure cosine vector distance against pgvector `jina_embedding`.
- `keyword`: PostgreSQL full-text search across titles and chunk bodies.

**Response:**
```json
{
  "mode": "hybrid",
  "query": "How do consumer group rebalances occur?",
  "total": 1,
  "results": [
    {
      "chunk_id": "d82f718a-9214-4112-9c19-15881cba9012",
      "content_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
      "title": "Kafka Architecture Crash Course",
      "content_type": "video",
      "source_url": "https://www.youtube.com/watch?v=0ZJgIjIuY7U",
      "start_seconds": 182.5,
      "end_seconds": 245.0,
      "page_index": null,
      "text": "When a new consumer joins or leaves the group, the group coordinator initiates a rebalance...",
      "score": 0.0322
    }
  ]
}
```
</details>

<details>
<summary><b>3. Spaces, Tags & Summaries</b></summary>

### List User Spaces
`GET /api/spaces?user_id=usr_99812401`

### Generate Instant AI Summary (Gemini 1.5 Flash)
`GET /api/content/:id/summary`

**Response:**
```json
{
  "summary": "This video breaks down event streaming and consumer partition reassignments in Apache Kafka.",
  "bullets": [
    "Explains how group coordinators detect consumer heartbeat dropouts.",
    "Compares eager rebalance protocols against cooperative sticky rebalances.",
    "Shows best practices for configuring session timeouts."
  ],
  "source": "gemini-1.5-flash"
}
```
</details>

---

## 🚀 Quick Start

### Prerequisites
- [Docker & Docker Compose](https://www.docker.com/) (recommended for containerized run)
- [Go 1.23+](https://golang.org/dl/) (if running backend natively)
- [Python 3.11+](https://www.python.org/downloads/) (if running AI service natively)
- [Node.js 20+](https://nodejs.org/) & `npm` (if running frontend natively)
- A Neon PostgreSQL instance with the `pgvector` extension enabled.

---

### Option A: 1-Click Launch with Docker Compose (Recommended)

1. **Clone the repository:**
   ```bash
   git clone https://github.com/Adit3107/Mindshelf.git
   cd Mindshelf
   ```

2. **Configure your environment files:**
   ```bash
   # Root Compose Environment
   cp .env.compose.example .env

   # Service-specific Environments
   cp backend/.env.example backend/.env
   cp ai-service/.env.example ai-service/.env
   cp frontend/.env.example frontend/.env
   ```
   *(Fill in your PostgreSQL `DATABASE_URL` in `backend/.env` and your Clerk keys in `frontend/.env`)*.

3. **Start all services:**
   ```bash
   docker compose up -d --build
   ```

4. **Access your apps:**
   - **Frontend UI**: [http://localhost:3000](http://localhost:3000) (or port 80 if configured)
   - **Go Backend API**: [http://localhost:8080/api/health](http://localhost:8080/api/health)
   - **FastAPI AI Service**: [http://localhost:8001/health](http://localhost:8001/health)

---

### Option B: Local Native Development

<details>
<summary><b>Click to expand step-by-step native instructions</b></summary>

#### 1. Start the Python AI Service
```powershell
cd ai-service
python -m venv .venv
# On Windows:
.\.venv\Scripts\Activate.ps1
# On Linux/macOS:
# source .venv/bin/activate

pip install -r requirements.txt
uvicorn app.main:app --reload --port 8001
```

#### 2. Start the Go Backend
```powershell
cd backend
go run ./cmd/server
# Automatically connects to Neon PG, pings, and executes embedded migrations
```

#### 3. Start the Next.js Frontend
```powershell
cd frontend
npm install
npm run dev
# Open http://localhost:3000 in your browser
```
</details>

---

## ⚙️ Environment Configuration

<details>
<summary><b>Backend (<code>backend/.env</code>)</b></summary>

```env
PORT=8080
DATABASE_URL=postgresql://user:password@ep-sample-12345.neon.tech/mindshelf?sslmode=require

# AWS S3 Storage (Optional for local test, required for file uploads)
AWS_REGION=ap-south-1
AWS_S3_BUCKET=mindshelf-assets
AWS_ACCESS_KEY_ID=YOUR_AWS_ACCESS_KEY
AWS_SECRET_ACCESS_KEY=YOUR_AWS_SECRET_KEY

# Internal AI Microservice Connection
AI_SERVICE_URL=http://localhost:8001
EMBEDDING_DIMENSION=256
EMBEDDING_MAX_CONCURRENCY=2

# Gemini AI Summarization
GEMINI_API_KEY=your_gemini_api_key
GEMINI_MODEL=gemini-1.5-flash
```
</details>

<details>
<summary><b>AI Service (<code>ai-service/.env</code>)</b></summary>

```env
HOST=127.0.0.1
PORT=8001
EMBEDDING_MODEL_NAME=jinaai/jina-embeddings-v5-text-nano
EMBEDDING_DIMENSION=256
ALLOW_HASH_EMBEDDINGS=false
TESSERACT_CMD=tesseract
```
</details>

<details>
<summary><b>Frontend (<code>frontend/.env</code>)</b></summary>

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api

# Clerk Authentication Keys
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_sample_clerk_key
CLERK_SECRET_KEY=sk_test_sample_clerk_key
NEXT_PUBLIC_CLERK_SIGN_IN_URL=/login
NEXT_PUBLIC_CLERK_SIGN_UP_URL=/signup
NEXT_PUBLIC_CLERK_SIGN_IN_FALLBACK_REDIRECT_URL=/app
NEXT_PUBLIC_CLERK_SIGN_UP_FALLBACK_REDIRECT_URL=/app
```
</details>

---

## 🧪 Testing & Verification

Mindshelf maintains comprehensive test suites across both backend and AI microservices:

```bash
# 1. Run all Go Backend Unit & Integration Tests
cd backend
go test ./...

# 2. Run Python AI Service Unit Tests
cd ai-service
python -m unittest discover -s tests

# 3. Verify Next.js Compilation & TypeScript Types
cd frontend
npm run build
npm run lint
```

---

## 🌐 Production Deployment & Nginx Reverse Proxy

In production (e.g. AWS EC2 with domain `mindshelf.in`), Nginx handles public SSL termination and proxies requests to Docker containers:

```mermaid
graph LR
    UserPublic["User Browser<br/>(HTTPS: 443)"] --> Nginx["Nginx Reverse Proxy<br/>(Let's Encrypt SSL)"]
    Nginx -->|"Proxy /"| DockerFront["Frontend Container<br/>(Port 3000)"]
    Nginx -->|"Proxy /api/"| DockerBack["Backend Container<br/>(Port 8080)"]
    DockerBack -->|"Internal Network"| DockerAI["AI Service Container<br/>(Port 8001)"]
```

### Production Nginx Configuration (`/etc/nginx/sites-available/mindshelf`)
```nginx
server {
    listen 80;
    server_name mindshelf.in www.mindshelf.in;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name mindshelf.in www.mindshelf.in;

    ssl_certificate /etc/letsencrypt/live/mindshelf.in/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/mindshelf.in/privkey.pem;

    client_max_body_size 50M;

    # API Proxy to Go Backend
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Frontend Proxy to Next.js
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

---

## 📂 Project Structure

```text
Mindshelf/
├── .env.compose.example      # Root environment template for Docker Compose
├── docker-compose.yml        # Multi-container orchestration (Frontend, Backend, AI)
├── README.md                 # Project documentation
│
├── backend/                  # Go Application Gateway & API Server
│   ├── cmd/server/           # Application entrypoint (main.go)
│   ├── internal/
│   │   ├── config/           # Environment loader & config parsing
│   │   ├── database/         # PostgreSQL connection & embedded SQL migrations
│   │   │   └── migrations/   # pgvector schema, indexes, chunking tables
│   │   ├── handlers/         # Gin HTTP handlers (Ingestion, Search, Users, Spaces)
│   │   ├── ingestion/        # Cleaning, chunking, and Python HTTP client
│   │   ├── models/           # Domain entity definitions
│   │   ├── repository/       # PostgreSQL data access layer (pgvector queries)
│   │   ├── routes/           # REST endpoint mapping
│   │   ├── services/         # Business logic, Hybrid Search (RRF), Gemini AI summaries
│   │   └── storage/          # AWS S3 integration
│   ├── Dockerfile
│   └── go.mod
│
├── ai-service/               # Python AI & Multimodal Microservice
│   ├── app/
│   │   ├── config/           # Model & environment settings
│   │   ├── routes/           # FastAPI endpoints (/embeddings, /extract/*, /health)
│   │   ├── services/
│   │   │   ├── documents.py  # PDF, DOCX, PPTX, XLSX parsers
│   │   │   ├── embedding_service.py # Jina v5 text nano inference (256d)
│   │   │   ├── fusion.py     # Multimodal audio + OCR alignment
│   │   │   ├── images.py     # Tesseract OCR for uploaded graphics
│   │   │   ├── reels.py      # Instagram Reels extraction & keyframes
│   │   │   ├── video_ocr.py  # RapidOCR ONNX-runtime keyframe reader
│   │   │   └── videos.py     # yt-dlp & faster-whisper speech-to-text
│   │   └── main.py           # FastAPI application startup
│   ├── tests/                # AI microservice unit tests
│   ├── Dockerfile
│   └── requirements.txt
│
└── frontend/                 # Next.js 16 Web Application
    ├── src/
    │   ├── app/              # App router (Landing, Dashboard, Search, Spaces, Settings)
    │   ├── components/       # UI design system & interactive widgets
    │   │   ├── auth/         # Clerk authentication & user sync
    │   │   ├── content/      # Library browser & content detail views
    │   │   ├── dashboard/    # Hero greeting, quick actions, overview cards
    │   │   ├── save/         # URL & document upload forms with progress
    │   │   ├── search/       # Hybrid search browser & Ctrl+K command palette
    │   │   └── theme/        # Dark/light mode theme provider
    │   ├── lib/              # API client & auth-sync helpers
    │   └── types/            # TypeScript schemas & contracts
    ├── package.json
    └── next.config.ts
```

---

<div align="center">

Crafted with ❤️ by the **Mindshelf Team**.  
*Never lose a video insight, presentation slide, or research snippet again.*

</div>
