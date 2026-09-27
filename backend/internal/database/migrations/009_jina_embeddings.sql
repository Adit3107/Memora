ALTER TABLE content_chunks
	ADD COLUMN IF NOT EXISTS jina_embedding vector(256);

CREATE INDEX IF NOT EXISTS idx_content_chunks_jina_embedding
	ON content_chunks
	USING ivfflat (jina_embedding vector_cosine_ops)
	WITH (lists = 100)
	WHERE jina_embedding IS NOT NULL;

-- Keep the original 384-dimensional embedding column during rollout so existing
-- vectors remain intact while the Jina vectors are generated in batches.