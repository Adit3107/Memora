package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"memora-backend/internal/models"
)

type SearchRepository struct {
	db *sql.DB
}

type SearchFilters struct {
	UserID      string
	SpaceID     *string
	ContentIDs  []string
	ContentType *models.ContentType
	SourceType  *string
	TagIDs      []string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type ChunkSearchResult struct {
	ContentID      string             `json:"content_id"`
	ChunkID        string             `json:"chunk_id"`
	ChunkIndex     int                `json:"chunk_index"`
	Title          string             `json:"title"`
	ContentType    models.ContentType `json:"content_type"`
	SourceURL      *string            `json:"source_url,omitempty"`
	ThumbnailURL   *string            `json:"thumbnail_url,omitempty"`
	Text           string             `json:"text"`
	Score          float64            `json:"score"`
	PageIndex      *int               `json:"page_index,omitempty"`
	StartSeconds   *float64           `json:"start_seconds,omitempty"`
	EndSeconds     *float64           `json:"end_seconds,omitempty"`
	SourceType     string             `json:"source_type"`
	EmbeddingModel string             `json:"embedding_model,omitempty"`
	Metadata       map[string]string  `json:"metadata"`
	Tags           []string           `json:"tags"`
}

type ContentMetadataResult struct {
	ContentID    string             `json:"content_id"`
	UserID       string             `json:"user_id"`
	Name         string             `json:"name"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	ContentType  models.ContentType `json:"content_type"`
	SourceURL    *string            `json:"source_url,omitempty"`
	ThumbnailURL *string            `json:"thumbnail_url,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Metadata     map[string]string  `json:"metadata"`
	MinSeconds   *float64           `json:"min_seconds,omitempty"`
	MaxSeconds   *float64           `json:"max_seconds,omitempty"`
	ChunkCount   int                `json:"chunk_count"`
}

func NewPostgresSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

func (r *SearchRepository) SemanticSearch(ctx context.Context, filters SearchFilters, embedding []float64, limit int, offset int) ([]ChunkSearchResult, error) {
	vectorArg := vectorLiteralForSearch(embedding)
	args := []any{vectorArg}
	whereSQL, whereArgs := searchWhereClause(filters, len(args)+1)
	args = append(args, whereArgs...)
	limitPlaceholder := len(args) + 1
	args = append(args, limit)
	offsetPlaceholder := len(args) + 1
	args = append(args, offset)

	return r.queryResults(ctx, fmt.Sprintf(`
		SELECT
			c.id,
			cc.id,
			cc.chunk_index,
			COALESCE(NULLIF(c.name, ''), c.title),
			c.type,
			c.source_url,
			c.thumbnail_url,
			cc.text,
			GREATEST(0, 1 - (cc.embedding <=> $1::vector)) AS score,
			cc.page_index,
			cc.start_seconds,
			cc.end_seconds,
			cc.source_type,
			cc.embedding_model,
			cc.metadata,
			%s
		FROM content_chunks cc
		JOIN content c ON c.id = cc.content_id
		WHERE cc.embedding IS NOT NULL
			%s
		ORDER BY cc.embedding <=> $1::vector
		LIMIT $%d OFFSET $%d
	`, tagsSubquery(), whereSQL, limitPlaceholder, offsetPlaceholder), args...)
}

func (r *SearchRepository) KeywordSearch(ctx context.Context, filters SearchFilters, query string, limit int, offset int) ([]ChunkSearchResult, error) {
	args := []any{query}
	whereSQL, whereArgs := searchWhereClause(filters, len(args)+1)
	args = append(args, whereArgs...)
	limitPlaceholder := len(args) + 1
	args = append(args, limit)
	offsetPlaceholder := len(args) + 1
	args = append(args, offset)

	document := `setweight(to_tsvector('english', COALESCE(NULLIF(c.name, ''), c.title) || ' ' || c.title || ' ' || c.description), 'A') || to_tsvector('english', cc.text)`
	tsQuery := `plainto_tsquery('english', $1)`

	return r.queryResults(ctx, fmt.Sprintf(`
		SELECT
			c.id,
			cc.id,
			cc.chunk_index,
			COALESCE(NULLIF(c.name, ''), c.title),
			c.type,
			c.source_url,
			c.thumbnail_url,
			cc.text,
			ts_rank_cd(%s, %s) AS score,
			cc.page_index,
			cc.start_seconds,
			cc.end_seconds,
			cc.source_type,
			cc.embedding_model,
			cc.metadata,
			%s
		FROM content_chunks cc
		JOIN content c ON c.id = cc.content_id
		WHERE %s @@ %s
			%s
		ORDER BY score DESC, c.created_at DESC
		LIMIT $%d OFFSET $%d
	`, document, tsQuery, tagsSubquery(), document, tsQuery, whereSQL, limitPlaceholder, offsetPlaceholder), args...)
}

func (r *SearchRepository) ContentChunks(ctx context.Context, filters SearchFilters, limit int) ([]ChunkSearchResult, error) {
	if limit <= 0 {
		limit = 50
	}
	args := []any{}
	whereSQL, whereArgs := searchWhereClause(filters, len(args)+1)
	args = append(args, whereArgs...)
	limitPlaceholder := len(args) + 1
	args = append(args, limit)

	return r.queryResults(ctx, fmt.Sprintf(`
		SELECT
			c.id,
			cc.id,
			cc.chunk_index,
			COALESCE(NULLIF(c.name, ''), c.title),
			c.type,
			c.source_url,
			c.thumbnail_url,
			cc.text,
			1.0 AS score,
			cc.page_index,
			cc.start_seconds,
			cc.end_seconds,
			cc.source_type,
			cc.embedding_model,
			cc.metadata,
			%s
		FROM content_chunks cc
		JOIN content c ON c.id = cc.content_id
		WHERE TRUE
			%s
		ORDER BY c.created_at DESC, cc.chunk_index ASC
		LIMIT $%d
	`, tagsSubquery(), whereSQL, limitPlaceholder), args...)
}

func (r *SearchRepository) ContentMetadata(ctx context.Context, userID string, contentID string) (ContentMetadataResult, error) {
	var result ContentMetadataResult
	var metadataJSON []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT
			c.id,
			c.user_id,
			c.name,
			c.title,
			c.description,
			c.type,
			c.source_url,
			c.thumbnail_url,
			c.created_at,
			c.updated_at,
			COALESCE(ir.metadata, '{}'::jsonb),
			MIN(cc.start_seconds),
			MAX(cc.end_seconds),
			COUNT(cc.id)
		FROM content c
		LEFT JOIN ingestion_results ir ON ir.content_id = c.id
		LEFT JOIN content_chunks cc ON cc.content_id = c.id
		WHERE c.user_id = $1
			AND c.id = $2
		GROUP BY c.id, c.user_id, c.name, c.title, c.description, c.type,
			c.source_url, c.thumbnail_url, c.created_at, c.updated_at, ir.metadata
	`, userID, contentID).Scan(
		&result.ContentID,
		&result.UserID,
		&result.Name,
		&result.Title,
		&result.Description,
		&result.ContentType,
		&result.SourceURL,
		&result.ThumbnailURL,
		&result.CreatedAt,
		&result.UpdatedAt,
		&metadataJSON,
		&result.MinSeconds,
		&result.MaxSeconds,
		&result.ChunkCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ContentMetadataResult{}, ErrNotFound
	}
	if err != nil {
		return ContentMetadataResult{}, err
	}
	if err := json.Unmarshal(metadataJSON, &result.Metadata); err != nil {
		return ContentMetadataResult{}, err
	}

	return result, nil
}

func (r *SearchRepository) queryResults(ctx context.Context, query string, args ...any) ([]ChunkSearchResult, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ChunkSearchResult, 0)
	for rows.Next() {
		var result ChunkSearchResult
		var metadataJSON []byte
		var tagsJSON []byte
		if err := rows.Scan(
			&result.ContentID,
			&result.ChunkID,
			&result.ChunkIndex,
			&result.Title,
			&result.ContentType,
			&result.SourceURL,
			&result.ThumbnailURL,
			&result.Text,
			&result.Score,
			&result.PageIndex,
			&result.StartSeconds,
			&result.EndSeconds,
			&result.SourceType,
			&result.EmbeddingModel,
			&metadataJSON,
			&tagsJSON,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(metadataJSON, &result.Metadata); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(tagsJSON, &result.Tags); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func searchWhereClause(filters SearchFilters, startIndex int) (string, []any) {
	clauses := []string{"c.user_id = $" + strconv.Itoa(startIndex)}
	args := []any{filters.UserID}
	next := startIndex + 1

	if filters.SpaceID != nil && strings.TrimSpace(*filters.SpaceID) != "" {
		clauses = append(clauses, "c.space_id = $"+strconv.Itoa(next))
		args = append(args, strings.TrimSpace(*filters.SpaceID))
		next++
	}
	if len(filters.ContentIDs) > 0 {
		placeholders := make([]string, 0, len(filters.ContentIDs))
		for _, contentID := range filters.ContentIDs {
			contentID = strings.TrimSpace(contentID)
			if contentID == "" {
				continue
			}
			placeholders = append(placeholders, "$"+strconv.Itoa(next))
			args = append(args, contentID)
			next++
		}
		if len(placeholders) > 0 {
			clauses = append(clauses, "c.id IN ("+strings.Join(placeholders, ", ")+")")
		}
	}
	if filters.ContentType != nil && strings.TrimSpace(string(*filters.ContentType)) != "" {
		clauses = append(clauses, "c.type = $"+strconv.Itoa(next))
		args = append(args, string(*filters.ContentType))
		next++
	}
	if filters.SourceType != nil && strings.TrimSpace(*filters.SourceType) != "" {
		clauses = append(clauses, "(cc.source_type = $"+strconv.Itoa(next)+" OR cc.metadata->>'provider' = $"+strconv.Itoa(next)+" OR cc.metadata->>'content_type' = $"+strconv.Itoa(next)+")")
		args = append(args, strings.TrimSpace(*filters.SourceType))
		next++
	}
	if filters.CreatedFrom != nil {
		clauses = append(clauses, "c.created_at >= $"+strconv.Itoa(next))
		args = append(args, *filters.CreatedFrom)
		next++
	}
	if filters.CreatedTo != nil {
		clauses = append(clauses, "c.created_at <= $"+strconv.Itoa(next))
		args = append(args, *filters.CreatedTo)
		next++
	}
	if len(filters.TagIDs) > 0 {
		placeholders := make([]string, 0, len(filters.TagIDs))
		for _, tagID := range filters.TagIDs {
			placeholders = append(placeholders, "$"+strconv.Itoa(next))
			args = append(args, strings.TrimSpace(tagID))
			next++
		}
		clauses = append(clauses, fmt.Sprintf(`
			EXISTS (
				SELECT 1
				FROM content_tags ct
				JOIN tags t ON t.id = ct.tag_id
				WHERE ct.content_id = c.id
					AND t.user_id = c.user_id
					AND ct.tag_id IN (%s)
			)
		`, strings.Join(placeholders, ", ")))
	}

	return "AND " + strings.Join(clauses, " AND "), args
}

func tagsSubquery() string {
	return `(
		SELECT COALESCE(jsonb_agg(t.name ORDER BY t.name), '[]'::jsonb)
		FROM content_tags ct
		JOIN tags t ON t.id = ct.tag_id
		WHERE ct.content_id = c.id
	) AS tags`
}

func vectorLiteralForSearch(values []float64) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.FormatFloat(value, 'f', -1, 64))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// Why this file exists:
// Search is a database concern over content, chunks, tags, full-text indexes,
// and pgvector. The repository owns SQL while services own orchestration.
