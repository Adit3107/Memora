package storage

import (
	"context"
	"io"
)

type PutObjectInput struct {
	Key         string
	Body        io.Reader
	ContentType string
	Size        int64
}

type ObjectMetadata struct {
	Bucket      string
	Key         string
	ContentType string
	Size        int64
}

type GetObjectOutput struct {
	Body        io.ReadCloser
	ContentType string
	Size        int64
}

type ObjectStore interface {
	PutObject(ctx context.Context, input PutObjectInput) (ObjectMetadata, error)
	GetObject(ctx context.Context, key string) (GetObjectOutput, error)
	DeleteObject(ctx context.Context, key string) error
}

// Why this file exists:
// This file defines the small storage contract the rest of Memora can depend on.
// An interface in Go describes behavior without choosing the concrete implementation.
// Later services can depend on ObjectStore without knowing whether files are stored in S3 or somewhere else.
