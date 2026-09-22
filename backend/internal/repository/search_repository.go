package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"memora-backend/internal/models"
)

type SearchRepository struct {
	db *sql.DB
}

type SemanticSearchResult struct {
	ContentID      string             `json:"content_id"`
	ChunkID        string             `json:"chunk_id"`
	ChunkIndex     int                `json:"chunk_index"`
	Title          string             `json:"title"`
	ContentType    models.ContentType `json:"content_type"`
	SourceURL      *string            `json:"source_url,omitempty"`
	Text           string             `json:"text"`
	Score          float64            `json:"score"`
	PageIndex      *int               `json:"page_index,omitempty"`
	StartSeconds   *float64           `json:"start_seconds,omitempty"`
	EndSeconds     *float64           `json:"end_seconds,omitempty"`
	EmbeddingModel string             `json:"embedding_model"`
	Metadata       map[string]string  `json:"metadata"`
}

func NewPostgresSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

func (r *SearchRepository) SemanticSearch(ctx context.Context, userID string, spaceID *string, embedding []float64, limit int) ([]SemanticSearchResult, error) {
	queryVector := vectorLiteralForSearch(embedding)
	spaceFilter := ""
	args := []any{userID, queryVector, limit}
	if spaceID != nil && strings.TrimSpace(*spaceID) != "" {
		spaceFilter = "AND c.space_id = $4"
		args = append(args, strings.TrimSpace(*spaceID))
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			c.id,
			cc.id,
			cc.chunk_index,
			c.title,
			c.type,
			c.source_url,
			cc.text,
			1 - (cc.embedding <=> $2::vector) AS score,
			cc.page_index,
			cc.start_seconds,
			cc.end_seconds,
			cc.embedding_model,
			cc.metadata
		FROM content_chunks cc
		JOIN content c ON c.id = cc.content_id
		WHERE c.user_id = $1
			AND cc.embedding IS NOT NULL
			`+spaceFilter+`
		ORDER BY cc.embedding <=> $2::vector
		LIMIT $3
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]SemanticSearchResult, 0)
	for rows.Next() {
		var result SemanticSearchResult
		var metadataJSON []byte
		if err := rows.Scan(
			&result.ContentID,
			&result.ChunkID,
			&result.ChunkIndex,
			&result.Title,
			&result.ContentType,
			&result.SourceURL,
			&result.Text,
			&result.Score,
			&result.PageIndex,
			&result.StartSeconds,
			&result.EndSeconds,
			&result.EmbeddingModel,
			&metadataJSON,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(metadataJSON, &result.Metadata); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func vectorLiteralForSearch(values []float64) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.FormatFloat(value, 'f', -1, 64))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// Why this file exists:
// Semantic search is a database concern over stored chunk embeddings. Keeping
// it in a repository preserves the backend's handler/service/repository split.
