package repositories

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"backend/config"
	"backend/utils"
)

var DB *sql.DB

func InitDB(pDebug *utils.HelperStruct) error {
	pDebug.Log(utils.Statement, "InitDB(+)")
	defer pDebug.Log(utils.Statement, "InitDB(-)")

	dbCfg := config.AppConfig.Database
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.DBName)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		pDebug.Log(utils.Elog, "DB01", err.Error())
		return utils.ErrReturn(err)
	}

	if err = DB.Ping(); err != nil {
		pDebug.Log(utils.Elog, "DB02", fmt.Sprintf("PostgreSQL connection failed at %s:%d (%s). Please verify PostgreSQL service/container is running.", dbCfg.Host, dbCfg.Port, err.Error()))
		return utils.ErrReturn(fmt.Errorf("unable to connect to PostgreSQL at %s:%d (please ensure PostgreSQL service or Docker is running): %w", dbCfg.Host, dbCfg.Port, err))
	}

	// Register extensions
	if _, err := DB.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`); err != nil {
		pDebug.Log(utils.Elog, "DB03", "Failed to create pgcrypto extension: "+err.Error())
	}
	if _, err := DB.Exec(`CREATE EXTENSION IF NOT EXISTS "vector"`); err != nil {
		pDebug.Log(utils.Elog, "DB04", "Failed to create vector extension: "+err.Error())
	}

	err = createTables(pDebug)
	if err != nil {
		return utils.ErrReturn(err)
	}

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

func createTables(pDebug *utils.HelperStruct) error {
	pDebug.Log(utils.Statement, "createTables(+)")
	defer pDebug.Log(utils.Statement, "createTables(-)")

	// 1. Remove legacy user foreign key validations (Session-based, decoupled architecture)
	legacyCleanupStmts := []string{
		"ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_user_id_fkey CASCADE;",
		"ALTER TABLE documents ALTER COLUMN user_id DROP NOT NULL;",
		"ALTER TABLE documents ALTER COLUMN user_id SET DEFAULT NULL;",
		"ALTER TABLE query_logs DROP CONSTRAINT IF EXISTS query_logs_user_id_fkey CASCADE;",
		"ALTER TABLE query_logs ALTER COLUMN user_id DROP NOT NULL;",
		"ALTER TABLE query_logs ALTER COLUMN user_id SET DEFAULT NULL;",
		"DROP TABLE IF EXISTS sessions CASCADE;",
	}
	for _, sqlStmt := range legacyCleanupStmts {
		_, _ = DB.Exec(sqlStmt)
	}

	// 2. Documents table (Session-based, no user login/account validation)
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID DEFAULT NULL,
			title VARCHAR(255),
			filename VARCHAR(255),
			status VARCHAR(50), -- UPLOADING, PROCESSING, COMPLETED, ERROR
			summary JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return utils.ErrReturn(err)
	}

	// Safely extend documents table with metadata fields (Phase 7)
	alterDocCols := []string{
		"ALTER TABLE documents ADD COLUMN IF NOT EXISTS stored_file_name VARCHAR(255);",
		"ALTER TABLE documents ADD COLUMN IF NOT EXISTS file_hash VARCHAR(64);",
		"ALTER TABLE documents ADD COLUMN IF NOT EXISTS file_size BIGINT DEFAULT 0;",
		"ALTER TABLE documents ADD COLUMN IF NOT EXISTS mime_type VARCHAR(100) DEFAULT 'application/pdf';",
		"ALTER TABLE documents ADD COLUMN IF NOT EXISTS page_count INT DEFAULT 0;",
		"ALTER TABLE documents ADD COLUMN IF NOT EXISTS text_page_count INT DEFAULT 0;",
		"ALTER TABLE documents ADD COLUMN IF NOT EXISTS validation_status VARCHAR(50) DEFAULT 'PENDING';",
		"ALTER TABLE documents ADD COLUMN IF NOT EXISTS processing_status VARCHAR(50) DEFAULT 'PENDING';",
		"ALTER TABLE documents ADD COLUMN IF NOT EXISTS pii_detected BOOLEAN DEFAULT FALSE;",
		"ALTER TABLE documents ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;",
	}
	for _, sqlStmt := range alterDocCols {
		if _, err := DB.Exec(sqlStmt); err != nil {
			pDebug.Log(utils.Elog, "DB_ALT_DOC", err.Error())
		}
	}

	// 3. Chunks table (with pgvector)
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS chunks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
			page_number INT,
			section_name VARCHAR(255),
			chunk_index INT,
			text TEXT,
			embedding vector(1536)
		)
	`)
	if err != nil {
		return utils.ErrReturn(err)
	}

	// Safely extend chunks table with metadata (Phase 9 & 23)
	alterChunkCols := []string{
		"ALTER TABLE chunks ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64);",
		"ALTER TABLE chunks ADD COLUMN IF NOT EXISTS embedding_provider VARCHAR(50);",
		"ALTER TABLE chunks ADD COLUMN IF NOT EXISTS embedding_model VARCHAR(100);",
		"ALTER TABLE chunks ADD COLUMN IF NOT EXISTS embedding_dimension INT DEFAULT 1536;",
		"ALTER TABLE chunks ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;",
		"CREATE INDEX IF NOT EXISTS idx_chunks_document_id ON chunks (document_id);",
		"CREATE INDEX IF NOT EXISTS idx_chunks_fts ON chunks USING gin(to_tsvector('english', text));",
	}
	for _, sqlStmt := range alterChunkCols {
		if _, err := DB.Exec(sqlStmt); err != nil {
			pDebug.Log(utils.Elog, "DB_ALT_CHK", err.Error())
		}
	}

	// 4. Document Validation Audit table (Phase 8)
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS document_validation (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
			check_name VARCHAR(100) NOT NULL,
			status VARCHAR(20) NOT NULL, -- PASS, FAIL, WARN
			message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_doc_val_doc_id ON document_validation(document_id);
	`)
	if err != nil {
		return utils.ErrReturn(err)
	}

	// 5. Query Logs Audit table (Phase 25)
	_, err = DB.Exec(`
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
	`)
	if err != nil {
		return utils.ErrReturn(err)
	}

	return nil
}
