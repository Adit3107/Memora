package ingestion

import "net/http"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Why this file exists:
// Extractors need HTTP, but tests should not call real external services.
// This tiny interface lets production use http.Client while tests use local fakes.
