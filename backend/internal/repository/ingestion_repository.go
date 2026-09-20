package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"memora-backend/internal/ingestion"
)

type IngestionStatus string

const (
	IngestionStatusPending    IngestionStatus = "pending"
	IngestionStatusProcessing IngestionStatus = "processing"
	IngestionStatusCompleted  IngestionStatus = "completed"
	IngestionStatusFailed     IngestionStatus = "failed"
)

type IngestionRepository struct {
	db *sql.DB
}

func NewPostgresIngestionRepository(db *sql.DB) *IngestionRepository {
	return &IngestionRepository{db: db}
}

func (r *IngestionRepository) CreatePending(ctx context.Context, contentID string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO ingestion_results (content_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		RETURNING id
	`, contentID, IngestionStatusPending, time.Now().UTC()).Scan(&id)
	return id, err
}

func (r *IngestionRepository) MarkProcessing(ctx context.Context, ingestionID string) error {
	return r.setStatus(ctx, ingestionID, IngestionStatusProcessing, "")
}

func (r *IngestionRepository) MarkFailed(ctx context.Context, ingestionID string, message string) error {
	return r.setStatus(ctx, ingestionID, IngestionStatusFailed, message)
}

func (r *IngestionRepository) SaveCompleted(ctx context.Context, ingestionID string, contentID string, result ingestion.IngestionResult, chunks []ingestion.ContentChunk) error {
	metadataJSON, err := json.Marshal(result.Metadata)
	if err != nil {
		return err
	}
	transcriptJSON, err := json.Marshal(result.Transcript)
	if err != nil {
		return err
	}
	pagesJSON, err := json.Marshal(result.Pages)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
		UPDATE ingestion_results
		SET status = $1, error_message = '', raw_text = $2, clean_text = $3,
			metadata = $4, transcript = $5, pages = $6, updated_at = $7
		WHERE id = $8
	`, IngestionStatusCompleted, result.RawText, result.CleanText, metadataJSON, transcriptJSON, pagesJSON, now, ingestionID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM content_chunks WHERE content_id = $1
	`, contentID); err != nil {
		return err
	}

	for _, chunk := range chunks {
		chunkMetadataJSON, err := json.Marshal(chunk.Metadata)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO content_chunks (
				content_id, chunk_index, text, source_type, page_index,
				start_seconds, end_seconds, metadata, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
			contentID,
			chunk.Index,
			chunk.Text,
			chunk.SourceType,
			chunk.PageIndex,
			chunk.StartSeconds,
			chunk.EndSeconds,
			chunkMetadataJSON,
			now,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *IngestionRepository) setStatus(ctx context.Context, ingestionID string, status IngestionStatus, message string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE ingestion_results
		SET status = $1, error_message = $2, updated_at = $3
		WHERE id = $4
	`, status, message, time.Now().UTC(), ingestionID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Why this file exists:
// Repositories are the database boundary in this backend.
// This one stores ingestion state, extracted text, metadata, transcript/pages,
// and chunks without mixing SQL into handlers or extraction code.
