-- pgcrypto gives us gen_random_uuid(), used for text UUID primary keys.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- users are the root owner for Memora data.
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
	name TEXT NOT NULL,
	-- UNIQUE prevents two user rows from sharing the same email.
	email TEXT NOT NULL UNIQUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- spaces organize a user's saved knowledge into topic areas.
CREATE TABLE IF NOT EXISTS spaces (
	id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
	-- ON DELETE CASCADE removes a user's spaces if the user is deleted.
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index foreign keys because we will often query spaces by user.
CREATE INDEX IF NOT EXISTS idx_spaces_user_id ON spaces(user_id);

-- content stores metadata only; actual file bytes belong in object storage later.
CREATE TABLE IF NOT EXISTS content (
	id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	space_id TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	-- CHECK keeps content type values aligned with the Go ContentType constants.
	type TEXT NOT NULL CHECK (type IN ('video', 'document', 'article', 'image')),
	source_url TEXT,
	thumbnail_url TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_content_user_id ON content(user_id);
CREATE INDEX IF NOT EXISTS idx_content_space_id ON content(space_id);
CREATE INDEX IF NOT EXISTS idx_content_type ON content(type);

-- tags are normal user labels like RAG, AI, DSA, Kafka.
CREATE TABLE IF NOT EXISTS tags (
	id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	-- One user should not have duplicate tag names.
	UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);

-- content_tags is the many-to-many join table between content and tags.
CREATE TABLE IF NOT EXISTS content_tags (
	content_id TEXT NOT NULL REFERENCES content(id) ON DELETE CASCADE,
	tag_id TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	-- Composite primary key prevents attaching the same tag twice to one content item.
	PRIMARY KEY (content_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_content_tags_tag_id ON content_tags(tag_id);

-- Why this file exists:
-- This migration creates Memora's first reproducible PostgreSQL schema.
-- It answers "what tables and relationships does a fresh Neon database need?"
-- Future schema changes should be new migration files, not manual database edits.
