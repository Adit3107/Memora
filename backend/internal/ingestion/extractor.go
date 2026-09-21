package ingestion

import (
	"context"
	"io"
)

type ExtractInput struct {
	SourceURL   string
	FileName    string
	ContentType string
	Body        io.Reader
	FilePath    string
}

type Extractor interface {
	Extract(ctx context.Context, input ExtractInput) (IngestionResult, error)
}

// Why this file exists:
// Extractors hide source-specific retrieval details behind one behavior. Some
// extractors run locally in Go, while future document/OCR extractors may be thin
// Go clients that call the internal Python extraction service. Both approaches
// return IngestionResult so cleaning, chunking, and persistence stay in Go.
