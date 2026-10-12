package repositories

import (
	"database/sql"

	"github.com/pgvector/pgvector-go"

	"backend/models"
	"backend/utils"
)

type ChunkWithVector struct {
	models.Chunk
	Embedding []float32
}

func InsertChunks(pDebug *utils.HelperStruct, pChunks []ChunkWithVector) error {
	pDebug.Log(utils.Statement, "InsertChunks(+)")
	defer pDebug.Log(utils.Statement, "InsertChunks(-)")

	lTx, lErr := DB.Begin()
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_CHK_01", "Error beginning transaction:", lErr.Error())
		return utils.ErrReturn(lErr)
	}
	defer lTx.Rollback()

	lSqlStr := `
		INSERT INTO chunks (
			document_id, page_number, section_name, chunk_index, text,
			content_hash, embedding_provider, embedding_model, embedding_dimension, embedding
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	lStmt, lErr := lTx.Prepare(lSqlStr)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_CHK_02", "Error preparing insert statement:", lErr.Error())
		return utils.ErrReturn(lErr)
	}
	defer lStmt.Close()

	for _, lChunk := range pChunks {
		lVec := pgvector.NewVector(lChunk.Embedding)
		_, lErr := lStmt.Exec(
			lChunk.DocumentID, lChunk.PageNumber, lChunk.SectionName, lChunk.ChunkIndex, lChunk.Text,
			lChunk.ContentHash, lChunk.EmbeddingProvider, lChunk.EmbeddingModel, lChunk.EmbeddingDimension, lVec,
		)
		if lErr != nil {
			pDebug.Log(utils.Elog, "REPO_CHK_03", "Error executing chunk insert:", lErr.Error())
			return utils.ErrReturn(lErr)
		}
	}

	if lErr := lTx.Commit(); lErr != nil {
		pDebug.Log(utils.Elog, "REPO_CHK_04", "Error committing transaction:", lErr.Error())
		return utils.ErrReturn(lErr)
	}

	return nil
}

func SearchSimilarChunks(pDebug *utils.HelperStruct, pDocumentID string, pQueryEmbedding []float32, pLimit int) ([]models.Chunk, error) {
	pDebug.Log(utils.Statement, "SearchSimilarChunks(+)")
	defer pDebug.Log(utils.Statement, "SearchSimilarChunks(-)")

	lVec := pgvector.NewVector(pQueryEmbedding)

	lSqlStr := `
		SELECT id, document_id, page_number, section_name, chunk_index, text
		FROM chunks
		WHERE document_id = $1
		ORDER BY embedding <=> $2
		LIMIT $3`

	lRows, lErr := DB.Query(lSqlStr, pDocumentID, lVec, pLimit)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_CHK_05", "Error querying similar chunks:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}
	defer lRows.Close()

	var lChunks []models.Chunk
	for lRows.Next() {
		var lChunk models.Chunk
		if lScanErr := lRows.Scan(&lChunk.ID, &lChunk.DocumentID, &lChunk.PageNumber, &lChunk.SectionName, &lChunk.ChunkIndex, &lChunk.Text); lScanErr != nil {
			pDebug.Log(utils.Elog, "REPO_CHK_06", "Error scanning chunk row:", lScanErr.Error())
			return nil, utils.ErrReturn(lScanErr)
		}
		lChunks = append(lChunks, lChunk)
	}

	return lChunks, nil
}

func SearchHybridChunks(
	pDebug *utils.HelperStruct,
	pDocumentID string,
	pQueryText string,
	pQueryEmbedding []float32,
	pLimit int,
	pThreshold float64,
	pVecWeight float64,
	pKeyWeight float64,
) ([]models.ChunkSearchResult, error) {
	pDebug.Log(utils.Statement, "SearchHybridChunks(+)")
	defer pDebug.Log(utils.Statement, "SearchHybridChunks(-)")

	lVec := pgvector.NewVector(pQueryEmbedding)

	if pVecWeight <= 0 && pKeyWeight <= 0 {
		pVecWeight = 0.70
		pKeyWeight = 0.30
	}

	lSqlStr := `
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
		LIMIT $6`

	lRows, lErr := DB.Query(lSqlStr, pDocumentID, lVec, pQueryText, pVecWeight, pKeyWeight, pLimit)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_CHK_07", "Error executing hybrid search:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}
	defer lRows.Close()

	var lResults []models.ChunkSearchResult
	for lRows.Next() {
		var lRes models.ChunkSearchResult
		var lHash, lProvider, lModel sql.NullString
		var lDim sql.NullInt32

		if lScanErr := lRows.Scan(
			&lRes.ID, &lRes.DocumentID, &lRes.PageNumber, &lRes.SectionName, &lRes.ChunkIndex, &lRes.Text,
			&lHash, &lProvider, &lModel, &lDim,
			&lRes.VectorSimilarity, &lRes.KeywordRank, &lRes.CombinedScore,
		); lScanErr != nil {
			pDebug.Log(utils.Elog, "REPO_CHK_08", "Error scanning hybrid search row:", lScanErr.Error())
			return nil, utils.ErrReturn(lScanErr)
		}

		lRes.ContentHash = lHash.String
		lRes.EmbeddingProvider = lProvider.String
		lRes.EmbeddingModel = lModel.String
		lRes.EmbeddingDimension = int(lDim.Int32)

		// Filter candidates by threshold unless no results pass
		if lRes.CombinedScore >= pThreshold || len(lResults) == 0 {
			lResults = append(lResults, lRes)
		}
	}

	return lResults, nil
}
