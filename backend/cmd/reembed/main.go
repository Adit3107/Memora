package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"memora-backend/internal/config"
	"memora-backend/internal/database"
	"memora-backend/internal/ingestion"
)

const embeddingBatchSize = 128

type contentChunk struct {
	id   string
	text string
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	if err := database.Ping(ctx, db); err != nil {
		return err
	}

	embeddingClient := ingestion.NewPythonEmbeddingClient(
		cfg.AIServiceURL,
		nil,
		cfg.EmbeddingDimension,
	)
	lastID := ""
	processed := 0

	for {
		chunks, err := loadNextBatch(ctx, db, lastID)
		if err != nil {
			return err
		}
		if len(chunks) == 0 {
			break
		}

		texts := make([]string, len(chunks))
		for index, chunk := range chunks {
			texts[index] = chunk.text
		}
		embeddings, err := embeddingClient.GenerateDocuments(ctx, texts)
		if err != nil {
			return fmt.Errorf("generate Jina embeddings after chunk %s: %w", lastID, err)
		}
		if err := saveBatch(ctx, db, chunks, embeddings); err != nil {
			return err
		}

		lastID = chunks[len(chunks)-1].id
		processed += len(chunks)
		log.Printf("Jina embeddings generated for %d chunks", processed)
	}

	log.Printf("Jina embedding backfill complete: %d chunks", processed)
	return nil
}

func loadNextBatch(ctx context.Context, db *sql.DB, lastID string) ([]contentChunk, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, text
		FROM content_chunks
		WHERE jina_embedding IS NULL AND id > $1
		ORDER BY id
		LIMIT $2
	`, lastID, embeddingBatchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chunks := make([]contentChunk, 0, embeddingBatchSize)
	for rows.Next() {
		var chunk contentChunk
		if err := rows.Scan(&chunk.id, &chunk.text); err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chunks, nil
}

func saveBatch(ctx context.Context, db *sql.DB, chunks []contentChunk, embeddings ingestion.EmbeddingBatch) error {
	if len(chunks) != len(embeddings.Embeddings) {
		return fmt.Errorf("embedding count mismatch: got %d vectors for %d chunks", len(embeddings.Embeddings), len(chunks))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	for index, chunk := range chunks {
		if _, err := tx.ExecContext(ctx, `
			UPDATE content_chunks
			SET jina_embedding = $1::vector, embedding_model = $2, embedded_at = $3
			WHERE id = $4 AND jina_embedding IS NULL
		`, vectorLiteral(embeddings.Embeddings[index]), embeddings.Model, now, chunk.id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func vectorLiteral(values []float64) string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = strconv.FormatFloat(value, 'f', -1, 64)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
