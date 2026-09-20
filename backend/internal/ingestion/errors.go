package ingestion

import "errors"

var (
	ErrUnsupportedSourceType = errors.New("unsupported source type")
	ErrInaccessibleSource    = errors.New("source could not be accessed")
	ErrEmptyContent          = errors.New("extracted content is empty")
	ErrInvalidURL            = errors.New("invalid url")
	ErrUnexpectedContentType = errors.New("unexpected content type")
)

// Why this file exists:
// Ingestion failures are different from ordinary CRUD validation failures.
// Keeping these errors in this package lets handlers and services later return
// useful messages when extraction cannot continue.
