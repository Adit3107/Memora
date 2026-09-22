package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPythonEmbeddingClientGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var request pythonEmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(request.Texts) != 2 {
			t.Fatalf("expected 2 texts, got %d", len(request.Texts))
		}

		embedding := make([]float64, DefaultEmbeddingDimension)
		response := pythonEmbeddingResponse{
			Success:    true,
			Model:      "sentence-transformers/all-MiniLM-L6-v2",
			Dimension:  DefaultEmbeddingDimension,
			Embeddings: [][]float64{embedding, embedding},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewPythonEmbeddingClient(server.URL, server.Client(), DefaultEmbeddingDimension)
	got, err := client.Generate(context.Background(), []string{"first chunk", "second chunk"})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if got.Model == "" {
		t.Fatal("expected model identifier")
	}
	if got.Dimension != DefaultEmbeddingDimension {
		t.Fatalf("expected dimension %d, got %d", DefaultEmbeddingDimension, got.Dimension)
	}
	if len(got.Embeddings) != 2 {
		t.Fatalf("expected 2 embeddings, got %d", len(got.Embeddings))
	}
}

func TestPythonEmbeddingClientRejectsEmptyInput(t *testing.T) {
	client := NewPythonEmbeddingClient("http://localhost:8001", nil, DefaultEmbeddingDimension)
	_, err := client.Generate(context.Background(), nil)
	if !errors.Is(err, ErrEmptyContent) {
		t.Fatalf("expected ErrEmptyContent, got %v", err)
	}
}

func TestPythonEmbeddingClientRejectsInvalidResponseShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := pythonEmbeddingResponse{
			Success:    true,
			Model:      "bad-model",
			Dimension:  3,
			Embeddings: [][]float64{{0.1, 0.2, 0.3}},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewPythonEmbeddingClient(server.URL, server.Client(), DefaultEmbeddingDimension)
	_, err := client.Generate(context.Background(), []string{"query"})
	if !errors.Is(err, ErrEmbeddingFailed) {
		t.Fatalf("expected ErrEmbeddingFailed, got %v", err)
	}
}
