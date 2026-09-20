package ingestion

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRedditExtractorExtract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ".json") {
			t.Fatalf("path = %q, want .json suffix", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"data": {
					"children": [
						{
							"data": {
								"title": "Learning Go with Memora",
								"author": "gopher",
								"subreddit": "golang",
								"permalink": "/r/golang/comments/abc/learning_go_with_memora/",
								"selftext": "This post has useful body text.",
								"created_utc": 1700000000,
								"score": 42,
								"num_comments": 7
							}
						}
					]
				}
			}
		]`))
	}))
	defer server.Close()

	extractor := NewRedditExtractor(server.Client(), WithRedditJSONBaseURL(server.URL))
	got, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: "https://reddit.com/r/golang/comments/abc/title"})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if got.SourceType != SourceTypeArticle {
		t.Fatalf("SourceType = %q, want %q", got.SourceType, SourceTypeArticle)
	}
	if got.Title != "Learning Go with Memora" {
		t.Fatalf("Title = %q", got.Title)
	}
	if got.Author != "gopher" {
		t.Fatalf("Author = %q", got.Author)
	}
	if got.Metadata["subreddit"] != "golang" {
		t.Fatalf("subreddit metadata = %q", got.Metadata["subreddit"])
	}
	if got.Metadata["created_utc"] == "" {
		t.Fatal("created_utc metadata was empty")
	}
	if !strings.Contains(got.CleanText, "useful body text") {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
}

func TestRedditExtractorRejectsNonRedditURL(t *testing.T) {
	extractor := NewRedditExtractor(nil)

	_, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: "https://example.com/post"})
	if !errors.Is(err, ErrUnsupportedSourceType) {
		t.Fatalf("Extract() error = %v, want %v", err, ErrUnsupportedSourceType)
	}
}

func TestRedditExtractorHandlesUnavailableAndEmptyPosts(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
	}{
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			body:       `{"message":"not found"}`,
			wantErr:    ErrInaccessibleSource,
		},
		{
			name:       "empty listing",
			statusCode: http.StatusOK,
			body:       `[{"data":{"children":[]}}]`,
			wantErr:    ErrEmptyContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			extractor := NewRedditExtractor(server.Client(), WithRedditJSONBaseURL(server.URL))
			_, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: "https://reddit.com/r/golang/comments/abc/title"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Extract() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
