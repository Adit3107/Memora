-- Migration 008: Enforce cascade updates on foreign keys and optimize user queries

-- 1. Ensure foreign keys support ON UPDATE CASCADE so Clerk user sync can update identifiers cleanly
ALTER TABLE spaces DROP CONSTRAINT IF EXISTS spaces_user_id_fkey;
ALTER TABLE spaces ADD CONSTRAINT spaces_user_id_fkey 
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE content DROP CONSTRAINT IF EXISTS content_user_id_fkey;
ALTER TABLE content ADD CONSTRAINT content_user_id_fkey 
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE tags DROP CONSTRAINT IF EXISTS tags_user_id_fkey;
ALTER TABLE tags ADD CONSTRAINT tags_user_id_fkey 
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE;

-- 2. Performance indexes for user-scoped queries
CREATE INDEX IF NOT EXISTS idx_content_user_created ON content(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_spaces_user_created ON spaces(user_id, created_at DESC);
