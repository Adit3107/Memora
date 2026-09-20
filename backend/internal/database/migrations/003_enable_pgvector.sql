-- pgvector adds the vector type PostgreSQL needs for future embeddings.
-- We enable the extension now but do not create embedding columns yet.
CREATE EXTENSION IF NOT EXISTS vector;

-- Why this file exists:
-- Memora will later store embedding vectors for semantic search.
-- A vector dimension should match the chosen embedding model.
-- Since that model is not finalized yet, this migration only enables pgvector.
