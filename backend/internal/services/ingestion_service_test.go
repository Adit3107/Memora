package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"memora-backend/internal/ingestion"
)

func TestEmbedChunksPreservesChunkMetadataAndAddsVectors(t *testing.T) {
	server := embeddingServer(t, func(texts []string) pythonEmbeddingTestResponse {
		embeddings := make([][]float64, len(texts))
		for index := range embeddings {
			embedding := make([]float64, ingestion.DefaultEmbeddingDimension)
			embedding[0] = float64(index + 1)
			embeddings[index] = embedding
		}
		return pythonEmbeddingTestResponse{
			Success:    true,
			Model:      "sentence-transformers/all-MiniLM-L6-v2",
			Dimension:  ingestion.DefaultEmbeddingDimension,
			Embeddings: embeddings,
		}
	})
	defer server.Close()

	page := 3
	start := 12.5
	end := 18.25
	chunks := []ingestion.ContentChunk{
		{
			Index:      0,
			Text:       "document chunk",
			SourceType: ingestion.SourceTypeDocument,
			PageIndex:  &page,
		},
		{
			Index:        1,
			Text:         "youtube transcript chunk",
			SourceType:   ingestion.SourceTypeVideo,
			StartSeconds: &start,
			EndSeconds:   &end,
		},
	}

	service := &IngestionService{
		embeddingClient:         ingestion.NewPythonEmbeddingClient(server.URL, server.Client(), ingestion.DefaultEmbeddingDimension),
		embeddingDimension:      ingestion.DefaultEmbeddingDimension,
		embeddingMaxConcurrency: 1,
	}
	if err := service.embedChunks(context.Background(), chunks); err != nil {
		t.Fatalf("embedChunks returned error: %v", err)
	}

	if len(chunks[0].Embedding) != ingestion.DefaultEmbeddingDimension || chunks[0].EmbeddingModel == "" {
		t.Fatal("expected first chunk to have embedding and model")
	}
	if chunks[0].PageIndex == nil || *chunks[0].PageIndex != page {
		t.Fatal("expected page metadata to be preserved")
	}
	if chunks[1].StartSeconds == nil || *chunks[1].StartSeconds != start || chunks[1].EndSeconds == nil || *chunks[1].EndSeconds != end {
		t.Fatal("expected timestamp metadata to be preserved")
	}
}

func TestEmbedChunksUsesBoundedConcurrentBatches(t *testing.T) {
	var mu sync.Mutex
	active := 0
	maxActive := 0

	server := embeddingServer(t, func(texts []string) pythonEmbeddingTestResponse {
		mu.Lock()
		active++
		if active > maxActive {
			maxActive = active
		}
		mu.Unlock()

		time.Sleep(20 * time.Millisecond)

		mu.Lock()
		active--
		mu.Unlock()

		embeddings := make([][]float64, len(texts))
		for index := range embeddings {
			embeddings[index] = make([]float64, ingestion.DefaultEmbeddingDimension)
		}
		return pythonEmbeddingTestResponse{
			Success:    true,
			Model:      "sentence-transformers/all-MiniLM-L6-v2",
			Dimension:  ingestion.DefaultEmbeddingDimension,
			Embeddings: embeddings,
		}
	})
	defer server.Close()

	chunks := make([]ingestion.ContentChunk, ingestion.MaxEmbeddingBatchSize*3)
	for index := range chunks {
		chunks[index] = ingestion.ContentChunk{Index: index, Text: "chunk text", SourceType: ingestion.SourceTypeText}
	}

	service := &IngestionService{
		embeddingClient:         ingestion.NewPythonEmbeddingClient(server.URL, server.Client(), ingestion.DefaultEmbeddingDimension),
		embeddingDimension:      ingestion.DefaultEmbeddingDimension,
		embeddingMaxConcurrency: 2,
	}
	if err := service.embedChunks(context.Background(), chunks); err != nil {
		t.Fatalf("embedChunks returned error: %v", err)
	}

	if maxActive != 2 {
		t.Fatalf("expected max concurrency 2, got %d", maxActive)
	}
	for _, chunk := range chunks {
		if len(chunk.Embedding) != ingestion.DefaultEmbeddingDimension {
			t.Fatal("expected every chunk to receive an embedding")
		}
	}
}

func TestEmbedChunksStopsOnEmbeddingFailure(t *testing.T) {
	server := embeddingServer(t, func(texts []string) pythonEmbeddingTestResponse {
		return pythonEmbeddingTestResponse{
			Success: false,
			Error:   "embedding model is unavailable",
		}
	})
	defer server.Close()

	service := &IngestionService{
		embeddingClient:         ingestion.NewPythonEmbeddingClient(server.URL, server.Client(), ingestion.DefaultEmbeddingDimension),
		embeddingDimension:      ingestion.DefaultEmbeddingDimension,
		embeddingMaxConcurrency: 1,
	}
	err := service.embedChunks(context.Background(), []ingestion.ContentChunk{{Index: 0, Text: "chunk text"}})
	if err == nil {
		t.Fatal("expected embedding failure")
	}
}

type pythonEmbeddingTestResponse struct {
	Success    bool        `json:"success"`
	Model      string      `json:"model"`
	Dimension  int         `json:"dimension"`
	Embeddings [][]float64 `json:"embeddings"`
	Error      string      `json:"error"`
}

func embeddingServer(t *testing.T, handler func([]string) pythonEmbeddingTestResponse) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Texts []string `json:"texts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode embedding request: %v", err)
		}

		_ = json.NewEncoder(w).Encode(handler(request.Texts))
	}))
}
