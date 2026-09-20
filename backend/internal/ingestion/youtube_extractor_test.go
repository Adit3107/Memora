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
		{
			name:   "watch url",
			rawURL: "https://www.youtube.com/watch?v=abc123",
			want:   "abc123",
		},
		{
			name:   "shorts url",
			rawURL: "https://youtube.com/shorts/short123",
			want:   "short123",
		},
		{
			name:   "short url",
			rawURL: "https://youtu.be/shorturl123",
			want:   "shorturl123",
		},
		{
			name:   "embed url",
			rawURL: "https://www.youtube.com/embed/embed123",
			want:   "embed123",
		},
		{
			name:    "missing video id",
			rawURL:  "https://youtube.com/watch",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "non youtube url",
			rawURL:  "https://example.com/watch?v=abc",
			wantErr: ErrUnsupportedSourceType,
		},
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

func TestYouTubeExtractorExtractJSONTranscript(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oembed":
			if r.URL.Query().Get("format") != "json" {
				t.Fatalf("metadata format query = %q", r.URL.Query().Get("format"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"title":" Test Video ","author_name":" Memora Channel "}`))
		case "/timedtext":
			if r.URL.Query().Get("v") != "abc123" {
				t.Fatalf("transcript video id = %q", r.URL.Query().Get("v"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"events": [
					{"tStartMs": 0, "dDurationMs": 12000, "segs": [{"utf8":"Hello "},{"utf8":"world"}]},
					{"tStartMs": 12000, "dDurationMs": 13000, "segs": [{"utf8":"Second segment"}]}
				]
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	extractor := NewYouTubeExtractor(
		server.Client(),
		WithYouTubeMetadataEndpoint(server.URL+"/oembed"),
		WithYouTubeTranscriptEndpoint(server.URL+"/timedtext"),
	)

	got, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: "https://youtube.com/watch?v=abc123"})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if got.SourceType != SourceTypeVideo {
		t.Fatalf("SourceType = %q, want %q", got.SourceType, SourceTypeVideo)
	}
	if got.Title != "Test Video" {
		t.Fatalf("Title = %q", got.Title)
	}
	if got.Author != "Memora Channel" {
		t.Fatalf("Author = %q", got.Author)
	}
	if got.Metadata["video_id"] != "abc123" {
		t.Fatalf("video_id metadata = %q", got.Metadata["video_id"])
	}
	if len(got.Transcript) != 2 {
		t.Fatalf("Transcript length = %d, want 2", len(got.Transcript))
	}
	if got.Transcript[0].StartSeconds != 0 || got.Transcript[0].EndSeconds != 12 {
		t.Fatalf("first timestamp = %+v", got.Transcript[0])
	}
	if !strings.Contains(got.CleanText, "Hello world") || !strings.Contains(got.CleanText, "Second segment") {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
}

func TestYouTubeExtractorExtractXMLTranscript(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oembed":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"title":"XML Video","author_name":"Creator"}`))
		case "/timedtext":
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<transcript><text start="1.5" dur="2.25">Hello XML</text></transcript>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	extractor := NewYouTubeExtractor(
		server.Client(),
		WithYouTubeMetadataEndpoint(server.URL+"/oembed"),
		WithYouTubeTranscriptEndpoint(server.URL+"/timedtext"),
	)

	got, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: "https://youtu.be/xml123"})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if len(got.Transcript) != 1 {
		t.Fatalf("Transcript length = %d, want 1", len(got.Transcript))
	}
	if got.Transcript[0].StartSeconds != 1.5 || got.Transcript[0].EndSeconds != 3.75 {
		t.Fatalf("timestamp = %+v", got.Transcript[0])
	}
}

func TestYouTubeExtractorHandlesUnavailableTranscript(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oembed":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"title":"No Transcript","author_name":"Creator"}`))
		case "/timedtext":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"events":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	extractor := NewYouTubeExtractor(
		server.Client(),
		WithYouTubeMetadataEndpoint(server.URL+"/oembed"),
		WithYouTubeTranscriptEndpoint(server.URL+"/timedtext"),
	)

	_, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: "https://youtube.com/shorts/no-transcript"})
	if !errors.Is(err, ErrEmptyContent) {
		t.Fatalf("Extract() error = %v, want %v", err, ErrEmptyContent)
	}
}

func TestYouTubeExtractorHandlesUnavailableVideo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	extractor := NewYouTubeExtractor(
		server.Client(),
		WithYouTubeMetadataEndpoint(server.URL+"/oembed"),
		WithYouTubeTranscriptEndpoint(server.URL+"/timedtext"),
	)

	_, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: "https://youtube.com/watch?v=missing"})
	if !errors.Is(err, ErrInaccessibleSource) {
		t.Fatalf("Extract() error = %v, want %v", err, ErrInaccessibleSource)
	}
}
