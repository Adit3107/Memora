package ingestion

import (
	"errors"
	"testing"
)

func TestDetectURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    DetectedContentType
		wantErr error
	}{
		{
			name:   "youtube watch url",
			rawURL: "https://youtube.com/watch?v=abc",
			want:   DetectedContentTypeYouTube,
		},
		{
			name:   "youtube short url",
			rawURL: "https://www.youtube.com/shorts/abc",
			want:   DetectedContentTypeYouTube,
		},
		{
			name:   "youtu be url",
			rawURL: "https://youtu.be/abc",
			want:   DetectedContentTypeYouTube,
		},
		{
			name:   "markdown youtube url",
			rawURL: "[https://youtu.be/abc](https://youtu.be/abc)",
			want:   DetectedContentTypeYouTube,
		},
		{
			name:   "angle wrapped article url",
			rawURL: "<https://example.com/article>",
			want:   DetectedContentTypeWebArticle,
		},
		{
			name:   "web article url",
			rawURL: "https://example.com/article",
			want:   DetectedContentTypeWebArticle,
		},
		{
			name:    "not a url",
			rawURL:  "not-a-url",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "unsupported scheme",
			rawURL:  "ftp://example.com/file",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "empty url",
			rawURL:  "",
			wantErr: ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectURL(tt.rawURL)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("DetectURL() error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("DetectURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		mimeType string
		want     DetectedContentType
		wantErr  error
	}{
		{
			name:     "pdf by mime type",
			fileName: "paper.bin",
			mimeType: "application/pdf",
			want:     DetectedContentTypePDF,
		},
		{
			name:     "docx by extension",
			fileName: "document.docx",
			mimeType: "application/octet-stream",
			want:     DetectedContentTypeDOCX,
		},
		{
			name:     "pptx by mime type",
			fileName: "slides",
			mimeType: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
			want:     DetectedContentTypePPTX,
		},
		{
			name:     "txt by mime type",
			fileName: "notes",
			mimeType: "text/plain; charset=utf-8",
			want:     DetectedContentTypeTXT,
		},
		{
			name:     "csv by extension",
			fileName: "data.csv",
			mimeType: "",
			want:     DetectedContentTypeCSV,
		},
		{
			name:     "xlsx by mime type",
			fileName: "sheet",
			mimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			want:     DetectedContentTypeXLSX,
		},
		{
			name:     "image by mime type",
			fileName: "upload",
			mimeType: "image/png",
			want:     DetectedContentTypeImage,
		},
		{
			name:     "image by extension",
			fileName: "photo.webp",
			mimeType: "application/octet-stream",
			want:     DetectedContentTypeImage,
		},
		{
			name:     "unsupported file",
			fileName: "archive.zip",
			mimeType: "application/zip",
			wantErr:  ErrUnsupportedSourceType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectFile(tt.fileName, tt.mimeType)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("DetectFile() error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("DetectFile() = %q, want %q", got, tt.want)
			}
		})
	}
}
