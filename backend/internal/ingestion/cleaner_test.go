package ingestion

import (
	"testing"
)

func TestCleanIngestionResultPreservesStructure(t *testing.T) {
	result := IngestionResult{
		SourceType:  SourceTypeDocument,
		Title:       "  My   Document  ",
		Description: "Intro\r\n\r\n\r\nDetails",
		RawText:     "Heading\n\n  first   paragraph \x00\n\n\n- item one",
		CleanText:   "",
		Metadata: map[string]string{
			" file_name ": " notes.pdf ",
			"empty":       "   ",
		},
		Pages: []DocumentPage{
			{Index: 1, Text: " Page   one\n\ntext "},
			{Index: 2, Text: "   "},
		},
	}

	got, err := CleanIngestionResult(result)
	if err != nil {
		t.Fatalf("CleanIngestionResult() error = %v", err)
	}

	if got.Title != "My Document" {
		t.Fatalf("Title = %q", got.Title)
	}
	if got.CleanText != "Heading\n\nfirst paragraph\n\n- item one" {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
	if got.Metadata["file_name"] != "notes.pdf" {
		t.Fatalf("metadata = %#v", got.Metadata)
	}
	if _, ok := got.Metadata["empty"]; ok {
		t.Fatalf("empty metadata was preserved: %#v", got.Metadata)
	}
	if len(got.Pages) != 1 || got.Pages[0].Index != 1 || got.Pages[0].Text != "Page one\n\ntext" {
		t.Fatalf("Pages = %#v", got.Pages)
	}
}

func TestCleanIngestionResultPreservesTranscriptTimestamps(t *testing.T) {
	result := IngestionResult{
		SourceType: SourceTypeVideo,
		CleanText:  " transcript ",
		Transcript: []TranscriptSegment{
			{StartSeconds: 1.5, EndSeconds: 2.5, Text: " hello   world "},
			{StartSeconds: 3, EndSeconds: 4, Text: "   "},
		},
	}

	got, err := CleanIngestionResult(result)
	if err != nil {
		t.Fatalf("CleanIngestionResult() error = %v", err)
	}

	if len(got.Transcript) != 1 {
		t.Fatalf("Transcript length = %d, want 1", len(got.Transcript))
	}
	if got.Transcript[0].StartSeconds != 1.5 || got.Transcript[0].EndSeconds != 2.5 {
		t.Fatalf("timestamp changed: %#v", got.Transcript[0])
	}
	if got.Transcript[0].Text != "hello world" {
		t.Fatalf("Text = %q", got.Transcript[0].Text)
	}
}

func TestCleanIngestionResultRejectsEmptyContent(t *testing.T) {
	_, err := CleanIngestionResult(IngestionResult{
		SourceType: SourceTypeArticle,
		CleanText:  " \x00 ",
	})
	if err != ErrEmptyContent {
		t.Fatalf("CleanIngestionResult() error = %v, want %v", err, ErrEmptyContent)
	}
}
