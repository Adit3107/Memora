package services

import (
	"testing"

	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

func TestNormalizeSearchInputDefaultsAndCapsLimit(t *testing.T) {
	input, err := normalizeSearchInput(SearchInput{
		UserID: " user-1 ",
		Query:  " Kafka consumer groups ",
		Limit:  100,
	})
	if err != nil {
		t.Fatalf("normalizeSearchInput returned error: %v", err)
	}
	if input.UserID != "user-1" {
		t.Fatalf("expected trimmed user id, got %q", input.UserID)
	}
	if input.Query != "Kafka consumer groups" {
		t.Fatalf("expected trimmed query, got %q", input.Query)
	}
	if input.Mode != SearchModeHybrid {
		t.Fatalf("expected default hybrid mode, got %q", input.Mode)
	}
	if input.Limit != maxSearchLimit {
		t.Fatalf("expected capped limit %d, got %d", maxSearchLimit, input.Limit)
	}
}

func TestNormalizeSearchInputRejectsInvalidContentType(t *testing.T) {
	badType := models.ContentType("audio")
	_, err := normalizeSearchInput(SearchInput{
		UserID:      "user-1",
		Query:       "Kafka",
		ContentType: &badType,
	})
	if err == nil {
		t.Fatal("expected invalid content type error")
	}
}

func TestReciprocalRankFusionDeduplicatesChunks(t *testing.T) {
	semantic := []repository.ChunkSearchResult{
		{ChunkID: "chunk-a", Title: "A", Score: 0.8},
		{ChunkID: "chunk-b", Title: "B", Score: 0.7},
	}
	keyword := []repository.ChunkSearchResult{
		{ChunkID: "chunk-b", Title: "B", Score: 0.4},
		{ChunkID: "chunk-c", Title: "C", Score: 0.3},
	}

	got := reciprocalRankFusion(semantic, keyword)
	if len(got) != 3 {
		t.Fatalf("expected 3 deduplicated results, got %d", len(got))
	}
	if got[0].ChunkID != "chunk-b" {
		t.Fatalf("expected chunk-b to rank first because it appears in both lists, got %s", got[0].ChunkID)
	}
}
