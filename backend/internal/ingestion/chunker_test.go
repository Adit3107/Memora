package ingestion

import (
	"strings"
	"testing"
)

func TestChunkIngestionResultShortContent(t *testing.T) {
	chunks, err := ChunkIngestionResult(IngestionResult{
		SourceType: SourceTypeArticle,
		CleanText:  "Short content.",
	}, ChunkConfig{MaxCharacters: 100, OverlapCharacters: 10})
	if err != nil {
		t.Fatalf("ChunkIngestionResult() error = %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("len(chunks) = %d, want 1", len(chunks))
	}
	if chunks[0].Text != "Short content." {
		t.Fatalf("Text = %q", chunks[0].Text)
	}
}

func TestChunkIngestionResultLongContent(t *testing.T) {
	text := strings.Join([]string{
		"Paragraph one has enough words to stand alone.",
		"Paragraph two also has enough words to trigger chunking.",
		"Paragraph three finishes the example cleanly.",
	}, "\n\n")

	chunks, err := ChunkIngestionResult(IngestionResult{
		SourceType: SourceTypeArticle,
		CleanText:  text,
	}, ChunkConfig{MaxCharacters: 80, OverlapCharacters: 20})
	if err != nil {
		t.Fatalf("ChunkIngestionResult() error = %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("len(chunks) = %d, want at least 2", len(chunks))
	}
	if chunks[0].SourceType != SourceTypeArticle {
		t.Fatalf("SourceType = %q", chunks[0].SourceType)
	}
}

func TestChunkIngestionResultTranscriptPreservesTimestamps(t *testing.T) {
	chunks, err := ChunkIngestionResult(IngestionResult{
		SourceType: SourceTypeVideo,
		CleanText:  "video transcript",
		Transcript: []TranscriptSegment{
			{StartSeconds: 0, EndSeconds: 3, Text: "first segment"},
			{StartSeconds: 3, EndSeconds: 6, Text: "second segment"},
		},
	}, ChunkConfig{MaxCharacters: 100, OverlapCharacters: 0})
	if err != nil {
		t.Fatalf("ChunkIngestionResult() error = %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("len(chunks) = %d, want 1", len(chunks))
	}
	if *chunks[0].StartSeconds != 0 || *chunks[0].EndSeconds != 6 {
		t.Fatalf("timestamps = %v-%v", *chunks[0].StartSeconds, *chunks[0].EndSeconds)
	}
}

func TestChunkIngestionResultPagesPreservePageIndex(t *testing.T) {
	chunks, err := ChunkIngestionResult(IngestionResult{
		SourceType: SourceTypeDocument,
		CleanText:  "document",
		Pages: []DocumentPage{
			{Index: 5, Text: "page five text"},
		},
	}, ChunkConfig{MaxCharacters: 100, OverlapCharacters: 0})
	if err != nil {
		t.Fatalf("ChunkIngestionResult() error = %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("len(chunks) = %d, want 1", len(chunks))
	}
	if chunks[0].PageIndex == nil || *chunks[0].PageIndex != 5 {
		t.Fatalf("PageIndex = %#v, want 5", chunks[0].PageIndex)
	}
}

func TestChunkIngestionResultEmptyContent(t *testing.T) {
	_, err := ChunkIngestionResult(IngestionResult{
		SourceType: SourceTypeArticle,
		CleanText:  "   ",
	}, ChunkConfig{})
	if err != ErrEmptyContent {
		t.Fatalf("ChunkIngestionResult() error = %v, want %v", err, ErrEmptyContent)
	}
}
