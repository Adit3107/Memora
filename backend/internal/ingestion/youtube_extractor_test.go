package ingestion

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestYouTubeVideoID(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr error
	}{
		{name: "watch url", rawURL: "https://www.youtube.com/watch?v=abc123", want: "abc123"},
		{name: "shorts url", rawURL: "https://youtube.com/shorts/short123", want: "short123"},
		{name: "short url", rawURL: "https://youtu.be/shorturl123", want: "shorturl123"},
		{name: "embed url", rawURL: "https://www.youtube.com/embed/embed123", want: "embed123"},
		{name: "missing video id", rawURL: "https://youtube.com/watch", wantErr: ErrInvalidURL},
		{name: "non youtube url", rawURL: "https://example.com/watch?v=abc", wantErr: ErrUnsupportedSourceType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := YouTubeVideoID(tt.rawURL)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("YouTubeVideoID() error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("YouTubeVideoID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestYouTubeExtractorUsesPythonTranscriptService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/extract/youtube" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"success": true,
			"title": "Transcript Video",
			"transcript_text": "Hello world second line",
			"combined_text": "Hello world second line",
			"transcript": [
				{"start_seconds": 0, "end_seconds": 1.5, "text": "Hello world"},
				{"start_seconds": 1.5, "end_seconds": 3, "text": "second line"}
			],
			"metadata": {"youtube_transcript_api": "true"}
		}`))
	}))
	defer server.Close()

	extractor := NewYouTubeExtractor(NewPythonExtractionClient(server.URL, server.Client()))
	got, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: "https://youtu.be/abc123"})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if got.SourceType != SourceTypeVideo {
		t.Fatalf("SourceType = %q, want %q", got.SourceType, SourceTypeVideo)
	}
	if got.Title != "Transcript Video" {
		t.Fatalf("Title = %q", got.Title)
	}
	if got.Metadata["video_id"] != "abc123" {
		t.Fatalf("video_id metadata = %q", got.Metadata["video_id"])
	}
	if got.Metadata["extraction_service"] != "python" {
		t.Fatalf("extraction_service metadata = %q", got.Metadata["extraction_service"])
	}
	if len(got.Transcript) != 2 {
		t.Fatalf("Transcript length = %d, want 2", len(got.Transcript))
	}
	if got.Transcript[0].StartSeconds != 0 || got.Transcript[0].EndSeconds != 1.5 {
		t.Fatalf("first timestamp = %+v", got.Transcript[0])
	}
	if !strings.Contains(got.CleanText, "Hello world") || !strings.Contains(got.CleanText, "second line") {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
}

func TestYouTubeExtractorHandlesUnavailableTranscript(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false,"error":"youtube transcript is unavailable"}`))
	}))
	defer server.Close()

	extractor := NewYouTubeExtractor(NewPythonExtractionClient(server.URL, server.Client()))
	_, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: "https://youtube.com/shorts/missing"})
	if !errors.Is(err, ErrTranscriptUnavailable) {
		t.Fatalf("Extract() error = %v, want %v", err, ErrTranscriptUnavailable)
	}
}
