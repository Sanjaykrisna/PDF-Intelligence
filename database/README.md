# Database Architecture & Query Catalog (PT_05 - PDF Intelligence)

This folder contains the complete database configuration, DDL schemas, and SQL query definitions for **PDF Intelligence**, structured alongside `backend` and `frontend`.

---

## 📁 Directory Files

| File | Purpose |
| :--- | :--- |
| `db_config.toml` | PostgreSQL connection parameters, credentials, pool limits, and retrieval configuration. |
| `schema.sql` | Full DDL schema initializing `pgcrypto`, `vector`, `documents`, `chunks`, `document_validation`, `query_logs`, and GIN/B-tree indexes. |
| `queries.sql` | Complete catalog of all DDL & DML queries executed across backend repository packages. |
| `docker-compose.yml` | One-click Docker container definition using `pgvector/pgvector:pg16` with auto-mounting of `schema.sql`. |

---

## 🛠️ PostgreSQL & pgvector Setup

### Option 1: Docker (Recommended)
Run the following inside this directory:
```bash
docker compose up -d
```
This spins up PostgreSQL 16 on port `5432` with `pgvector` and automatically runs `schema.sql`.

### Option 2: Native PostgreSQL
1. Ensure PostgreSQL 15+ or 16+ is installed with `pgvector`.
2. Connect to PostgreSQL and create the database:
   ```sql
   CREATE DATABASE project_review_ai;
   ```
3. Execute `schema.sql`:
   ```bash
   psql -U postgres -d project_review_ai -f schema.sql
   ```

---

## 🗄️ Database Tables Overview

1. **`documents`**:
   - Stores PDF document records, titles, filenames, physical disk paths, file hashes, sizes, validation and processing statuses.
   - Contains `summary` as `JSONB` holding structured educational analysis (objective, methodology, algorithms, results, future scope).
2. **`chunks`**:
   - Stores text chunks extracted from PDFs with page numbers, section titles, and content hashes.
   - Includes `embedding vector(1536)` indexed for fast cosine distance retrieval and a GIN index on `to_tsvector('english', text)` for full-text search.
3. **`document_validation`**:
   - Audit trail tracking integrity checks: PDF magic signature verification, extractable text check, and PII detection.
4. **`query_logs`**:
   - Complete RAG telemetry: logs questions, cosine similarity score, retrieval attempts, LLM provider, answer validity, and fallback status.
