-- ==============================================================================
-- Database Schema Definition - PDF Intelligence (PT_05)
-- Database Engine: PostgreSQL 15+ / 16+ with pgvector
-- Target DB: project_review_ai
-- ==============================================================================

-- 1. Enable Required Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";

-- 2. Clean Up Legacy Constraints (if migrating from earlier builds)
ALTER TABLE IF EXISTS documents DROP CONSTRAINT IF EXISTS documents_user_id_fkey CASCADE;
ALTER TABLE IF EXISTS query_logs DROP CONSTRAINT IF EXISTS query_logs_user_id_fkey CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;

-- 3. Documents Table
-- Tracks uploaded PDF metadata, validation statuses, hashes, and AI-generated summaries
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID DEFAULT NULL,
    title VARCHAR(255),
    filename VARCHAR(255),
    stored_file_name VARCHAR(255),
    file_hash VARCHAR(64),
    file_size BIGINT DEFAULT 0,
    mime_type VARCHAR(100) DEFAULT 'application/pdf',
    page_count INT DEFAULT 0,
    text_page_count INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'PROCESSING',          -- UPLOADING, PROCESSING, COMPLETED, ERROR
    validation_status VARCHAR(50) DEFAULT 'PENDING',  -- PENDING, IN_PROGRESS, PASSED, FAILED
    processing_status VARCHAR(50) DEFAULT 'PENDING',  -- PENDING, PROCESSING, COMPLETED, ERROR
    pii_detected BOOLEAN DEFAULT FALSE,
    summary JSONB,                                    -- Structured JSON summary (objectives, methodology, etc.)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_documents_user_id ON documents (user_id);
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents (created_at DESC);

-- 4. Chunks Table
-- Stores segmented PDF text chunks along with 1536-dimensional embeddings for vector search
CREATE TABLE IF NOT EXISTS chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    page_number INT,
    section_name VARCHAR(255),
    chunk_index INT,
    text TEXT,
    content_hash VARCHAR(64),
    embedding_provider VARCHAR(50) DEFAULT 'gemini',
    embedding_model VARCHAR(100) DEFAULT 'gemini-embedding-001',
    embedding_dimension INT DEFAULT 1536,
    embedding vector(1536),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexing for relational lookups
CREATE INDEX IF NOT EXISTS idx_chunks_document_id ON chunks (document_id);

-- GIN Index for PostgreSQL Full-Text Search (Keyword retrieval)
CREATE INDEX IF NOT EXISTS idx_chunks_fts ON chunks USING gin(to_tsvector('english', text));

-- 5. Document Validation Audit Table
-- Captures detailed verification steps (Magic number, MIME, text extractability, PII scans)
CREATE TABLE IF NOT EXISTS document_validation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    check_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,                      -- PASS, FAIL, WARN
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_doc_val_doc_id ON document_validation (document_id);

-- 6. Query Logs Audit Table
-- Complete observability for RAG operations, question answering, cosine similarity, and LLM fallback
CREATE TABLE IF NOT EXISTS query_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID DEFAULT NULL,
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    original_query TEXT NOT NULL,
    normalized_query TEXT NOT NULL,
    query_type VARCHAR(50),                           -- direct, basic, reasoning
    retrieval_attempt INT DEFAULT 1,
    top_similarity FLOAT8 DEFAULT 0,
    context_valid BOOLEAN DEFAULT TRUE,
    llm_provider VARCHAR(50),                         -- gemini, ollama, direct_lookup
    llm_attempt INT DEFAULT 1,
    answer_valid BOOLEAN DEFAULT TRUE,
    fallback_used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_query_logs_doc_id ON query_logs (document_id);
CREATE INDEX IF NOT EXISTS idx_query_logs_created_at ON query_logs (created_at DESC);
