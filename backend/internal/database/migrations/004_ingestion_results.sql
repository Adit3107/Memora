CREATE TABLE IF NOT EXISTS ingestion_results (
	id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
	content_id TEXT NOT NULL UNIQUE REFERENCES content(id) ON DELETE CASCADE,
	status TEXT NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
	error_message TEXT NOT NULL DEFAULT '',
	raw_text TEXT NOT NULL DEFAULT '',
	clean_text TEXT NOT NULL DEFAULT '',
	metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
	transcript JSONB NOT NULL DEFAULT '[]'::jsonb,
	pages JSONB NOT NULL DEFAULT '[]'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ingestion_results_content_id ON ingestion_results(content_id);
CREATE INDEX IF NOT EXISTS idx_ingestion_results_status ON ingestion_results(status);

CREATE TABLE IF NOT EXISTS content_chunks (
	id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
	content_id TEXT NOT NULL REFERENCES content(id) ON DELETE CASCADE,
	chunk_index INT NOT NULL,
	text TEXT NOT NULL,
	source_type TEXT NOT NULL,
	page_index INT,
	start_seconds DOUBLE PRECISION,
	end_seconds DOUBLE PRECISION,
	metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (content_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS idx_content_chunks_content_id ON content_chunks(content_id);

-- Why this file exists:
-- Ingestion is a small state machine around the existing content table.
-- content stores user-facing metadata, ingestion_results stores extracted text/status,
-- and content_chunks stores the final Phase 4 output that Phase 5 embeddings will use.
