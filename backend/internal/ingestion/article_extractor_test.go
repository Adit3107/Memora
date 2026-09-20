package ingestion

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebArticleExtractorExtract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`
			<!doctype html>
			<html>
				<head>
					<title> Example Article </title>
					<meta name="description" content="A useful article for testing.">
					<meta name="author" content="Ada Lovelace">
				</head>
				<body>
					<nav>Home Pricing Login</nav>
					<article>
						<h1>Example Article</h1>
						<p>This is the first useful paragraph.</p>
						<p>This is the second useful paragraph.</p>
					</article>
					<footer>Copyright footer text</footer>
					<script>console.log("ignore me")</script>
				</body>
			</html>
		`))
	}))
	defer server.Close()

	extractor := NewWebArticleExtractor(server.Client())
	got, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: server.URL + "/article"})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if got.SourceType != SourceTypeArticle {
		t.Fatalf("SourceType = %q, want %q", got.SourceType, SourceTypeArticle)
	}
	if got.Title != "Example Article" {
		t.Fatalf("Title = %q", got.Title)
	}
	if got.Description != "A useful article for testing." {
		t.Fatalf("Description = %q", got.Description)
	}
	if got.Author != "Ada Lovelace" {
		t.Fatalf("Author = %q", got.Author)
	}
	if !strings.Contains(got.CleanText, "first useful paragraph") {
		t.Fatalf("CleanText did not include article text: %q", got.CleanText)
	}
	if strings.Contains(got.CleanText, "Pricing") || strings.Contains(got.CleanText, "Copyright") {
		t.Fatalf("CleanText included boilerplate: %q", got.CleanText)
	}
}

func TestWebArticleExtractorRejectsNonArticleURLTypes(t *testing.T) {
	extractor := NewWebArticleExtractor(nil)

	_, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: "https://youtube.com/watch?v=abc"})
	if !errors.Is(err, ErrUnsupportedSourceType) {
		t.Fatalf("Extract() error = %v, want %v", err, ErrUnsupportedSourceType)
	}
}

func TestWebArticleExtractorHandlesStatusAndContentType(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		contentType string
		body        string
		wantErr     error
	}{
		{
			name:        "forbidden",
			statusCode:  http.StatusForbidden,
			contentType: "text/html",
			body:        "forbidden",
			wantErr:     ErrInaccessibleSource,
		},
		{
			name:        "non html",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body:        "{}",
			wantErr:     ErrUnexpectedContentType,
		},
		{
			name:        "empty page",
			statusCode:  http.StatusOK,
			contentType: "text/html",
			body:        "<html><head><title>Only title</title></head><body><script>empty</script></body></html>",
			wantErr:     ErrEmptyContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			extractor := NewWebArticleExtractor(server.Client())
			_, err := extractor.Extract(context.Background(), ExtractInput{SourceURL: server.URL})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Extract() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
