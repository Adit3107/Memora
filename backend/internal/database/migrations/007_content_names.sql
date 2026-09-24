ALTER TABLE content
	ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';

UPDATE content
SET name = CASE
	WHEN btrim(title) <> '' AND title !~* '^https?://' THEN btrim(title)
	WHEN type = 'video' AND COALESCE(source_url, title) ILIKE '%/shorts/%'
		THEN 'YouTube Short ' || COALESCE(NULLIF(split_part(split_part(COALESCE(source_url, title), '/shorts/', 2), '?', 1), ''), left(id, 8))
	WHEN type = 'video' AND COALESCE(source_url, title) ILIKE '%youtu.be/%'
		THEN 'YouTube video ' || COALESCE(NULLIF(split_part(split_part(COALESCE(source_url, title), 'youtu.be/', 2), '?', 1), ''), left(id, 8))
	WHEN type = 'video' AND COALESCE(source_url, title) ILIKE '%youtube.com/watch%'
		THEN 'YouTube video ' || COALESCE(NULLIF(split_part(split_part(COALESCE(source_url, title), 'v=', 2), '&', 1), ''), left(id, 8))
	WHEN type = 'video'
		THEN 'YouTube video ' || left(id, 8)
	WHEN type = 'document'
		THEN 'Document ' || left(id, 8)
	WHEN type = 'image'
		THEN 'Image ' || left(id, 8)
	ELSE 'Saved content ' || left(id, 8)
END
WHERE btrim(name) = '';

CREATE INDEX IF NOT EXISTS idx_content_name_fts
	ON content
	USING GIN (to_tsvector('english', name || ' ' || title || ' ' || description));

-- Why this file exists:
-- URLs are good source identifiers, but poor library labels. This migration
-- adds a user-facing content name and backfills existing URL-titled saves into
-- readable labels while preserving the original source_url and title fields.
