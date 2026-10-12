-- ==============================================================================
-- SQL Query Catalog - PDF Intelligence (PT_05)
-- All database queries executed by the Go backend repositories
-- ==============================================================================

-- ==============================================================================
-- 1. DATABASE SETUP & SCHEMA DDL QUERIES
-- Executed by: repositories/db.go -> InitDB(), createTables()
-- ==============================================================================

-- 1.1 Enable Required Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";

-- 1.2 Cleanup Legacy Constraints
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_user_id_fkey CASCADE;
ALTER TABLE documents ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE documents ALTER COLUMN user_id SET DEFAULT NULL;
ALTER TABLE query_logs DROP CONSTRAINT IF EXISTS query_logs_user_id_fkey CASCADE;
ALTER TABLE query_logs ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE query_logs ALTER COLUMN user_id SET DEFAULT NULL;
DROP TABLE IF EXISTS sessions CASCADE;

-- 1.3 Create Documents Table
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID DEFAULT NULL,
    title VARCHAR(255),
    filename VARCHAR(255),
    status VARCHAR(50),
    summary JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 1.4 Dynamic Document Column Additions
ALTER TABLE documents ADD COLUMN IF NOT EXISTS stored_file_name VARCHAR(255);
ALTER TABLE documents ADD COLUMN IF NOT EXISTS file_hash VARCHAR(64);
ALTER TABLE documents ADD COLUMN IF NOT EXISTS file_size BIGINT DEFAULT 0;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS mime_type VARCHAR(100) DEFAULT 'application/pdf';
ALTER TABLE documents ADD COLUMN IF NOT EXISTS page_count INT DEFAULT 0;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS text_page_count INT DEFAULT 0;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS validation_status VARCHAR(50) DEFAULT 'PENDING';
ALTER TABLE documents ADD COLUMN IF NOT EXISTS processing_status VARCHAR(50) DEFAULT 'PENDING';
ALTER TABLE documents ADD COLUMN IF NOT EXISTS pii_detected BOOLEAN DEFAULT FALSE;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- 1.5 Create Chunks Table
CREATE TABLE IF NOT EXISTS chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    page_number INT,
    section_name VARCHAR(255),
    chunk_index INT,
    text TEXT,
    embedding vector(1536)
);

-- 1.6 Dynamic Chunks Column Additions & Indexes
ALTER TABLE chunks ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64);
ALTER TABLE chunks ADD COLUMN IF NOT EXISTS embedding_provider VARCHAR(50);
ALTER TABLE chunks ADD COLUMN IF NOT EXISTS embedding_model VARCHAR(100);
ALTER TABLE chunks ADD COLUMN IF NOT EXISTS embedding_dimension INT DEFAULT 1536;
ALTER TABLE chunks ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_chunks_document_id ON chunks (document_id);
CREATE INDEX IF NOT EXISTS idx_chunks_fts ON chunks USING gin(to_tsvector('english', text));

-- 1.7 Create Document Validation Audit Table
CREATE TABLE IF NOT EXISTS document_validation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    check_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_doc_val_doc_id ON document_validation(document_id);

-- 1.8 Create Query Logs Audit Table
CREATE TABLE IF NOT EXISTS query_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID DEFAULT NULL,
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    original_query TEXT NOT NULL,
    normalized_query TEXT NOT NULL,
    query_type VARCHAR(50),
    retrieval_attempt INT DEFAULT 1,
    top_similarity FLOAT8 DEFAULT 0,
    context_valid BOOLEAN DEFAULT TRUE,
    llm_provider VARCHAR(50),
    llm_attempt INT DEFAULT 1,
    answer_valid BOOLEAN DEFAULT TRUE,
    fallback_used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_query_logs_doc_id ON query_logs(document_id);


-- ==============================================================================
-- 2. DOCUMENT MANAGEMENT QUERIES
-- Executed by: repositories/document_repo.go
-- ==============================================================================

-- 2.1 Insert New Document Record
-- Function: CreateDocument(pDebug, pDoc)
-- Parameters: $1: user_id, $2: title, $3: filename, $4: stored_file_name,
--             $5: file_hash, $6: file_size, $7: mime_type, $8: page_count,
--             $9: text_page_count, $10: status, $11: validation_status,
--             $12: processing_status, $13: pii_detected
INSERT INTO documents (
    user_id, title, filename, stored_file_name, file_hash,
    file_size, mime_type, page_count, text_page_count,
    status, validation_status, processing_status, pii_detected
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id;

-- 2.2 Update Document Processing Status
-- Function: UpdateDocumentStatus(pDebug, pDocumentID, pStatus)
-- Parameters: $1: status, $2: document_id
UPDATE documents 
SET status = $1, processing_status = $1, updated_at = CURRENT_TIMESTAMP 
WHERE id = $2;

-- 2.3 Update Document Metadata After Text Extraction & PII Scan
-- Function: UpdateDocumentMetadata(pDebug, pDocID, pPageCount, pTextPageCount, pPiiDetected, pValidStatus, pProcStatus)
-- Parameters: $1: page_count, $2: text_page_count, $3: pii_detected,
--             $4: validation_status, $5: processing_status, $6: document_id
UPDATE documents 
SET page_count = $1, text_page_count = $2, pii_detected = $3, 
    validation_status = $4, processing_status = $5, updated_at = CURRENT_TIMESTAMP
WHERE id = $6;

-- 2.4 Update Document Summary JSONB
-- Function: UpdateDocumentSummary(pDebug, pDocumentID, pSummary)
-- Parameters: $1: summary (JSONB), $2: document_id
UPDATE documents 
SET summary = $1, updated_at = CURRENT_TIMESTAMP 
WHERE id = $2;

-- 2.5 Fetch Single Document Summary & Details
-- Function: GetDocumentSummary(pDebug, pDocumentID, pUserID)
-- Parameters: $1: document_id, $2: user_id
SELECT id, user_id, title, filename, COALESCE(stored_file_name, ''), COALESCE(file_hash, ''),
       COALESCE(file_size, 0), COALESCE(mime_type, 'application/pdf'), COALESCE(page_count, 0),
       COALESCE(text_page_count, 0), status, COALESCE(validation_status, 'PENDING'),
       COALESCE(processing_status, 'PENDING'), COALESCE(pii_detected, FALSE), summary, created_at, updated_at
FROM documents 
WHERE id = $1 AND ($2 = '' OR $2 = '00000000-0000-0000-0000-000000000001' OR user_id = $2::uuid OR user_id = '00000000-0000-0000-0000-000000000001'::uuid OR user_id IS NULL);

-- 2.6 Fetch Latest Document For User/Session
-- Function: GetLatestDocumentByUser(pDebug, pUserID)
-- Parameters: $1: user_id
SELECT id, user_id, title, filename, COALESCE(stored_file_name, ''), COALESCE(file_hash, ''),
       COALESCE(file_size, 0), COALESCE(mime_type, 'application/pdf'), COALESCE(page_count, 0),
       COALESCE(text_page_count, 0), status, COALESCE(validation_status, 'PENDING'),
       COALESCE(processing_status, 'PENDING'), COALESCE(pii_detected, FALSE), summary, created_at, updated_at
FROM documents 
WHERE ($1 = '' OR $1 = '00000000-0000-0000-0000-000000000001' OR user_id = $1::uuid OR user_id IS NULL) 
ORDER BY created_at DESC 
LIMIT 1;

-- 2.7 List Documents For User with Subquery Total Chunks (Max 3 Documents Policy)
-- Function: GetDocumentsByUser(pDebug, pUserID, pLimit)
-- Parameters: $1: user_id, $2: limit (typically 3)
SELECT d.id, d.user_id, d.title, d.filename, COALESCE(d.stored_file_name, ''), COALESCE(d.file_hash, ''),
       COALESCE(d.file_size, 0), COALESCE(d.mime_type, 'application/pdf'), COALESCE(page_count, 0),
       COALESCE(d.text_page_count, 0), d.status, COALESCE(d.validation_status, 'PENDING'),
       COALESCE(d.processing_status, 'PENDING'), COALESCE(d.pii_detected, FALSE), d.summary, d.created_at, d.updated_at,
       COALESCE((SELECT COUNT(*) FROM chunks c WHERE c.document_id = d.id), 0) AS chunk_count
FROM documents d
WHERE ($1 = '' OR $1 = '00000000-0000-0000-0000-000000000001' OR d.user_id = $1::uuid OR d.user_id IS NULL)
ORDER BY d.created_at DESC
LIMIT $2;

-- 2.8 Document Metrics Aggregation for Insights
-- Function: GetDocumentInsights(pDebug, pDocumentID, pUserID)
-- Parameters: $1: document_id
SELECT COUNT(*), COALESCE(SUM(LENGTH(text)), 0), COALESCE(SUM(array_length(regexp_split_to_array(text, '\s+'), 1)), 0)
FROM chunks 
WHERE document_id = $1;

-- 2.9 Delete Document Row (Cascades to chunks, audits, and query_logs)
-- Function: DeleteDocument(pDebug, pDocumentID, pUserID)
-- Step A: Fetch stored file name to remove from filesystem
SELECT stored_file_name 
FROM documents 
WHERE id = $1 AND ($2 = '' OR $2 = '00000000-0000-0000-0000-000000000001' OR user_id = $2::uuid OR user_id IS NULL);

-- Step B: Delete row from table
DELETE FROM documents 
WHERE id = $1 AND ($2 = '' OR $2 = '00000000-0000-0000-0000-000000000001' OR user_id = $2::uuid OR user_id IS NULL);

-- 2.10 Select Excess Document IDs for Retention Pruning
-- Function: PruneOldUserDocuments(pDebug, pUserID, pKeepCount)
-- Parameters: $1: user_id, $2: keepCount offset
SELECT id FROM documents 
WHERE ($1 = '' OR $1 = '00000000-0000-0000-0000-000000000001' OR user_id = $1::uuid OR user_id IS NULL) 
ORDER BY created_at DESC 
OFFSET $2;

-- 2.11 Scheduled / Manual Database Flush
-- Function: FlushAllDocumentsAndChunks(pDebug)
-- Deletes all documents and cascades all child tables
DELETE FROM documents;


-- ==============================================================================
-- 3. CHUNK MANAGEMENT & VECTOR SEARCH QUERIES
-- Executed by: repositories/chunk_repo.go
-- ==============================================================================

-- 3.1 Batch Chunk & Vector Insertion
-- Function: InsertChunks(pDebug, pChunks)
-- Parameters: $1: document_id, $2: page_number, $3: section_name, $4: chunk_index,
--             $5: text, $6: content_hash, $7: embedding_provider,
--             $8: embedding_model, $9: embedding_dimension, $10: embedding vector
INSERT INTO chunks (
    document_id, page_number, section_name, chunk_index, text,
    content_hash, embedding_provider, embedding_model, embedding_dimension, embedding
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- 3.2 Pure Vector Cosine Distance Search
-- Function: SearchSimilarChunks(pDebug, pDocumentID, pQueryEmbedding, pLimit)
-- Parameters: $1: document_id, $2: query vector, $3: limit
SELECT id, document_id, page_number, section_name, chunk_index, text
FROM chunks
WHERE document_id = $1
ORDER BY embedding <=> $2
LIMIT $3;

-- 3.3 Hybrid Search (pgvector Cosine Distance + PostgreSQL ts_rank_cd Full-Text Search)
-- Function: SearchHybridChunks(pDebug, pDocumentID, pQueryText, pQueryEmbedding, pLimit, pThreshold, pVecWeight, pKeyWeight)
-- Parameters: $1: document_id
--             $2: query vector (1536 dim)
--             $3: query text string
--             $4: vector_weight (e.g. 0.70)
--             $5: keyword_weight (e.g. 0.30)
--             $6: limit (top_k, e.g. 5)
SELECT 
    id, document_id, page_number, section_name, chunk_index, text,
    COALESCE(content_hash, ''), COALESCE(embedding_provider, ''),
    COALESCE(embedding_model, ''), COALESCE(embedding_dimension, 1536),
    (1 - (embedding <=> $2)) AS vector_similarity,
    COALESCE(ts_rank_cd(to_tsvector('english', text), plainto_tsquery('english', $3)), 0.0) AS keyword_rank,
    (($4 * (1 - (embedding <=> $2))) + ($5 * COALESCE(ts_rank_cd(to_tsvector('english', text), plainto_tsquery('english', $3)), 0.0))) AS combined_score
FROM chunks
WHERE document_id = $1
ORDER BY combined_score DESC
LIMIT $6;


-- ==============================================================================
-- 4. VALIDATION AUDITS & OBSERVABILITY QUERIES
-- Executed by: repositories/audit_repo.go
-- ==============================================================================

-- 4.1 Record Validation Check
-- Function: RecordValidationAudit(pDebug, pDocID, pCheckName, pStatus, pMessage)
-- Parameters: $1: document_id, $2: check_name, $3: status (PASS/FAIL/WARN), $4: message
INSERT INTO document_validation (document_id, check_name, status, message)
VALUES ($1, $2, $3, $4);

-- 4.2 Retrieve Validation Audits For Document
-- Function: GetValidationAudits(pDebug, pDocID)
-- Parameters: $1: document_id
SELECT id, document_id, check_name, status, COALESCE(message, ''), created_at
FROM document_validation
WHERE document_id = $1
ORDER BY created_at ASC;

-- 4.3 Record Query Audit Log
-- Function: RecordQueryAudit(pDebug, pLog)
-- Parameters: $1: user_id, $2: document_id, $3: original_query, $4: normalized_query,
--             $5: query_type, $6: retrieval_attempt, $7: top_similarity,
--             $8: context_valid, $9: llm_provider, $10: llm_attempt,
--             $11: answer_valid, $12: fallback_used
INSERT INTO query_logs (
    user_id, document_id, original_query, normalized_query, query_type,
    retrieval_attempt, top_similarity, context_valid, llm_provider,
    llm_attempt, answer_valid, fallback_used
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
