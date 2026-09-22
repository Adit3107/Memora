ALTER TABLE content_chunks
	ADD COLUMN IF NOT EXISTS embedding vector(384),
	ADD COLUMN IF NOT EXISTS embedding_model TEXT NOT NULL DEFAULT '',
	ADD COLUMN IF NOT EXISTS embedded_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_content_chunks_embedding
	ON content_chunks
	USING ivfflat (embedding vector_cosine_ops)
	WITH (lists = 100)
	WHERE embedding IS NOT NULL;

-- Why this file exists:
-- Phase 5 turns completed ingestion chunks into searchable semantic vectors.
-- The vector dimension matches all-MiniLM-L6-v2 and the Python service contract.
