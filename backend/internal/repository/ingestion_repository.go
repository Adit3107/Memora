package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
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
	db                 *sql.DB
	embeddingDimension int
}

type StoredIngestionResult struct {
	ID           string                        `json:"id"`
	ContentID    string                        `json:"content_id"`
	Status       IngestionStatus               `json:"status"`
	ErrorMessage string                        `json:"error_message"`
	RawText      string                        `json:"raw_text"`
	CleanText    string                        `json:"clean_text"`
	Metadata     map[string]string             `json:"metadata"`
	Transcript   []ingestion.TranscriptSegment `json:"transcript"`
	Pages        []ingestion.DocumentPage      `json:"pages"`
	CreatedAt    time.Time                     `json:"created_at"`
	UpdatedAt    time.Time                     `json:"updated_at"`
}

func NewPostgresIngestionRepository(db *sql.DB, embeddingDimension int) *IngestionRepository {
	if embeddingDimension <= 0 {
		embeddingDimension = ingestion.DefaultEmbeddingDimension
	}

	return &IngestionRepository{
		db:                 db,
		embeddingDimension: embeddingDimension,
	}
}

func (r *IngestionRepository) GetByID(ctx context.Context, ingestionID string) (StoredIngestionResult, error) {
	var result StoredIngestionResult
	var metadataJSON []byte
	var transcriptJSON []byte
	var pagesJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, content_id, status, error_message, raw_text, clean_text,
			metadata, transcript, pages, created_at, updated_at
		FROM ingestion_results
		WHERE id = $1
	`, ingestionID).Scan(
		&result.ID,
		&result.ContentID,
		&result.Status,
		&result.ErrorMessage,
		&result.RawText,
		&result.CleanText,
		&metadataJSON,
		&transcriptJSON,
		&pagesJSON,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return StoredIngestionResult{}, ErrNotFound
	}
	if err != nil {
		return StoredIngestionResult{}, err
	}

	if err := json.Unmarshal(metadataJSON, &result.Metadata); err != nil {
		return StoredIngestionResult{}, err
	}
	if err := json.Unmarshal(transcriptJSON, &result.Transcript); err != nil {
		return StoredIngestionResult{}, err
	}
	if err := json.Unmarshal(pagesJSON, &result.Pages); err != nil {
		return StoredIngestionResult{}, err
	}

	return result, nil
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
	if err := validateEmbeddedChunks(chunks, r.embeddingDimension); err != nil {
		return err
	}

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
				start_seconds, end_seconds, metadata, embedding,
				embedding_model, embedded_at, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::vector, $10, $11, $12)
		`,
			contentID,
			chunk.Index,
			chunk.Text,
			chunk.SourceType,
			chunk.PageIndex,
			chunk.StartSeconds,
			chunk.EndSeconds,
			chunkMetadataJSON,
			vectorLiteral(chunk.Embedding),
			chunk.EmbeddingModel,
			embeddedAt(chunk.Embedding),
			now,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func validateEmbeddedChunks(chunks []ingestion.ContentChunk, embeddingDimension int) error {
	if embeddingDimension <= 0 {
		embeddingDimension = ingestion.DefaultEmbeddingDimension
	}

	if len(chunks) == 0 {
		return ingestion.ErrEmptyContent
	}

	for _, chunk := range chunks {
		if strings.TrimSpace(chunk.Text) == "" {
			return ingestion.ErrEmptyContent
		}
		if len(chunk.Embedding) != embeddingDimension {
			return ingestion.ErrEmbeddingFailed
		}
		if strings.TrimSpace(chunk.EmbeddingModel) == "" {
			return ingestion.ErrEmbeddingFailed
		}
	}

	return nil
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

func vectorLiteral(values []float64) *string {
	if len(values) == 0 {
		return nil
	}

	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.FormatFloat(value, 'f', -1, 64))
	}
	literal := "[" + strings.Join(parts, ",") + "]"
	return &literal
}

func embeddedAt(values []float64) *time.Time {
	if len(values) == 0 {
		return nil
	}

	now := time.Now().UTC()
	return &now
}

// Why this file exists:
// Repositories are the database boundary in this backend.
// This one stores ingestion state, extracted text, metadata, transcript/pages,
// and chunks without mixing SQL into handlers or extraction code.
