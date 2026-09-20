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
// Extractors hide source-specific retrieval details behind one behavior.
// A YouTube extractor, PDF extractor, and article extractor can all return
// IngestionResult while using very different logic internally.
