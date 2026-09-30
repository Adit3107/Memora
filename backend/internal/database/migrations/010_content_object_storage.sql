ALTER TABLE content
	ADD COLUMN IF NOT EXISTS storage_bucket TEXT,
	ADD COLUMN IF NOT EXISTS storage_key TEXT,
	ADD COLUMN IF NOT EXISTS original_filename TEXT,
	ADD COLUMN IF NOT EXISTS file_content_type TEXT,
	ADD COLUMN IF NOT EXISTS file_size BIGINT;

CREATE INDEX IF NOT EXISTS idx_content_storage_key ON content(storage_key);

-- Why this file exists:
-- Uploaded files need a durable pointer to object storage.
-- Postgres keeps searchable metadata/chunks; S3 keeps the original bytes.
