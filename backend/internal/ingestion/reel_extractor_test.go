package ingestion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReelExtractorInstagramSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/extract/reel" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"success": true,
			"title": "Learn Go Channels",
			"description": "Goroutines and channels in 30 seconds",
			"uploader": "golang_daily",
			"thumbnail_url": "https://instagram.com/p/thumb.jpg",
			"duration_seconds": 30.0,
			"transcript": [
				{"start_seconds": 0, "end_seconds": 3.5, "text": "Goroutines run concurrently."},
				{"start_seconds": 3.5, "end_seconds": 7.0, "text": "Channels pass data safely."}
			],
			"metadata": {
				"source_platform": "instagram",
				"uploader": "golang_daily"
			}
		}`))
	}))
	defer server.Close()

	client := NewPythonExtractionClient(server.URL, server.Client())
	extractor := NewReelExtractor(client)

	res, err := extractor.Extract(context.Background(), ExtractInput{
		SourceURL: "https://www.instagram.com/reel/C8xyz123/",
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if res.SourceType != SourceTypeReel {
		t.Errorf("SourceType = %v, want %v", res.SourceType, SourceTypeReel)
	}
	if res.Title != "Learn Go Channels" {
		t.Errorf("Title = %q, want %q", res.Title, "Learn Go Channels")
	}
	if res.Author != "golang_daily" {
		t.Errorf("Author = %q, want %q", res.Author, "golang_daily")
	}
	if len(res.Transcript) != 2 {
		t.Fatalf("len(Transcript) = %d, want 2", len(res.Transcript))
	}
	if res.Metadata["source_platform"] != "instagram" {
		t.Errorf("metadata[source_platform] = %q, want instagram", res.Metadata["source_platform"])
	}
	if res.Metadata["reel_id"] != "C8xyz123" {
		t.Errorf("metadata[reel_id] = %q, want C8xyz123", res.Metadata["reel_id"])
	}
}

func TestReelExtractorFacebookSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/extract/reel" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"success": true,
			"title": "System Design Tips",
			"description": "Kafka vs RabbitMQ",
			"uploader": "tech_architect",
			"thumbnail_url": "https://facebook.com/thumb.jpg",
			"duration_seconds": 45.0,
			"transcript": [
				{"start_seconds": 0, "end_seconds": 5.0, "text": "Kafka is a distributed log."}
			]
		}`))
	}))
	defer server.Close()

	client := NewPythonExtractionClient(server.URL, server.Client())
	extractor := NewReelExtractor(client)

	res, err := extractor.Extract(context.Background(), ExtractInput{
		SourceURL: "https://www.facebook.com/reel/9876543210",
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if res.Metadata["source_platform"] != "facebook" {
		t.Errorf("metadata[source_platform] = %q, want facebook", res.Metadata["source_platform"])
	}
	if res.Metadata["reel_id"] != "9876543210" {
		t.Errorf("metadata[reel_id] = %q, want 9876543210", res.Metadata["reel_id"])
	}
}

func TestReelExtractorRejectsNonReelURL(t *testing.T) {
	extractor := NewReelExtractor(NewPythonExtractionClient("http://localhost:8001", nil))
	_, err := extractor.Extract(context.Background(), ExtractInput{
		SourceURL: "https://example.com/article",
	})
	if err == nil {
		t.Fatal("expected error for non-reel URL, got nil")
	}
}
