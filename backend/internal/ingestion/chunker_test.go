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

func TestChunkIngestionResultReelProducesFullReelAndTimestampChunks(t *testing.T) {
	chunks, err := ChunkIngestionResult(IngestionResult{
		SourceType: SourceTypeReel,
		CleanText:  "First tip on goroutines. Second tip on channels.",
		Transcript: []TranscriptSegment{
			{StartSeconds: 0, EndSeconds: 5, Text: "First tip on goroutines."},
			{StartSeconds: 5, EndSeconds: 12, Text: "Second tip on channels."},
		},
	}, ChunkConfig{MaxCharacters: 30, OverlapCharacters: 0})
	if err != nil {
		t.Fatalf("ChunkIngestionResult() error = %v", err)
	}

	// Should have full reel chunk (index 0) + 2 smaller chunks
	if len(chunks) != 3 {
		t.Fatalf("len(chunks) = %d, want 3 (1 full reel + 2 timestamped chunks)", len(chunks))
	}
	if chunks[0].Metadata["chunk_scope"] != "full_reel" {
		t.Errorf("chunk[0] chunk_scope = %v, want full_reel", chunks[0].Metadata["chunk_scope"])
	}
	if *chunks[0].StartSeconds != 0 || *chunks[0].EndSeconds != 12 {
		t.Errorf("chunk[0] timestamps = %v-%v, want 0-12", *chunks[0].StartSeconds, *chunks[0].EndSeconds)
	}
	if *chunks[1].StartSeconds != 0 || *chunks[1].EndSeconds != 5 {
		t.Errorf("chunk[1] timestamps = %v-%v, want 0-5", *chunks[1].StartSeconds, *chunks[1].EndSeconds)
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
