package repository

import (
	"errors"
	"testing"

	"memora-backend/internal/ingestion"
)

func TestValidateEmbeddedChunksRequiresVectorsBeforeCompletion(t *testing.T) {
	err := validateEmbeddedChunks([]ingestion.ContentChunk{
		{
			Index:      0,
			Text:       "chunk without embedding",
			SourceType: ingestion.SourceTypeText,
		},
	}, ingestion.DefaultEmbeddingDimension)
	if !errors.Is(err, ingestion.ErrEmbeddingFailed) {
		t.Fatalf("expected ErrEmbeddingFailed, got %v", err)
	}
}

func TestValidateEmbeddedChunksAcceptsFullyEmbeddedChunks(t *testing.T) {
	err := validateEmbeddedChunks([]ingestion.ContentChunk{
		{
			Index:          0,
			Text:           "embedded chunk",
			SourceType:     ingestion.SourceTypeText,
			Embedding:      make([]float64, ingestion.DefaultEmbeddingDimension),
			EmbeddingModel: "sentence-transformers/all-MiniLM-L6-v2",
		},
	}, ingestion.DefaultEmbeddingDimension)
	if err != nil {
		t.Fatalf("expected embedded chunk to validate, got %v", err)
	}
}
