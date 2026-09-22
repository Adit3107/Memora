CREATE INDEX IF NOT EXISTS idx_content_chunks_text_fts
	ON content_chunks
	USING GIN (to_tsvector('english', text));

CREATE INDEX IF NOT EXISTS idx_content_title_fts
	ON content
	USING GIN (to_tsvector('english', title || ' ' || description));

CREATE INDEX IF NOT EXISTS idx_content_created_at ON content(created_at);

-- Why this file exists:
-- Phase 6 adds PostgreSQL keyword search over stored content chunks. These
-- indexes support full-text matching without adding Elasticsearch/OpenSearch.
