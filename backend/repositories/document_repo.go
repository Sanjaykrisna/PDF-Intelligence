package repositories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"backend/config"
	"backend/models"
	"backend/utils"
)

func CreateDocument(pDebug *utils.HelperStruct, pDoc *models.Document) (string, error) {
	pDebug.Log(utils.Statement, "CreateDocument(+)")
	defer pDebug.Log(utils.Statement, "CreateDocument(-)")

	var lID string
	lSqlStr := `
		INSERT INTO documents (
			user_id, title, filename, stored_file_name, file_hash,
			file_size, mime_type, page_count, text_page_count,
			status, validation_status, processing_status, pii_detected
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	lErr := DB.QueryRow(
		lSqlStr,
		pDoc.UserID, pDoc.Title, pDoc.Filename, pDoc.StoredFilename, pDoc.FileHash,
		pDoc.FileSize, pDoc.MimeType, pDoc.PageCount, pDoc.TextPageCount,
		pDoc.Status, pDoc.ValidationStatus, pDoc.ProcessingStatus, pDoc.PIIDetected,
	).Scan(&lID)

	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_01", "Error creating document record:", lErr.Error())
		return "", utils.ErrReturn(lErr)
	}

	return lID, nil
}

func UpdateDocumentStatus(pDebug *utils.HelperStruct, pDocumentID, pStatus string) error {
	pDebug.Log(utils.Statement, "UpdateDocumentStatus(+)")
	defer pDebug.Log(utils.Statement, "UpdateDocumentStatus(-)")

	lSqlStr := "UPDATE documents SET status = $1, processing_status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	_, lErr := DB.Exec(lSqlStr, pStatus, pDocumentID)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_02", "Error updating document status:", lErr.Error())
		return utils.ErrReturn(lErr)
	}

	return nil
}

func UpdateDocumentMetadata(pDebug *utils.HelperStruct, pDocID string, pPageCount, pTextPageCount int, pPiiDetected bool, pValidStatus, pProcStatus string) error {
	pDebug.Log(utils.Statement, "UpdateDocumentMetadata(+)")
	defer pDebug.Log(utils.Statement, "UpdateDocumentMetadata(-)")

	lSqlStr := `
		UPDATE documents 
		SET page_count = $1, text_page_count = $2, pii_detected = $3, 
		    validation_status = $4, processing_status = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6`

	_, lErr := DB.Exec(lSqlStr, pPageCount, pTextPageCount, pPiiDetected, pValidStatus, pProcStatus, pDocID)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_03", "Error updating document metadata:", lErr.Error())
		return utils.ErrReturn(lErr)
	}

	return nil
}

func UpdateDocumentSummary(pDebug *utils.HelperStruct, pDocumentID string, pSummary *models.DocumentSummary) error {
	pDebug.Log(utils.Statement, "UpdateDocumentSummary(+)")
	defer pDebug.Log(utils.Statement, "UpdateDocumentSummary(-)")

	lSummaryJSON, lErr := json.Marshal(pSummary)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_04", "Error marshalling document summary:", lErr.Error())
		return utils.ErrReturn(lErr)
	}

	lSqlStr := "UPDATE documents SET summary = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	_, lErr = DB.Exec(lSqlStr, lSummaryJSON, pDocumentID)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_05", "Error updating document summary:", lErr.Error())
		return utils.ErrReturn(lErr)
	}

	return nil
}

func GetDocumentSummary(pDebug *utils.HelperStruct, pDocumentID, pUserID string) (*models.Document, error) {
	pDebug.Log(utils.Statement, "GetDocumentSummary(+)")
	defer pDebug.Log(utils.Statement, "GetDocumentSummary(-)")

	var lDoc models.Document
	var lSummaryJSON []byte
	var lStoredFilename, lFileHash, lMimeType, lValidationStatus, lProcessingStatus sql.NullString
	var lFileSize sql.NullInt64
	var lPageCount, lTextPageCount sql.NullInt32
	var lPiiDetected sql.NullBool
	var lUpdatedAt sql.NullTime

	lSqlStr := `
		SELECT id, user_id, title, filename, COALESCE(stored_file_name, ''), COALESCE(file_hash, ''),
		       COALESCE(file_size, 0), COALESCE(mime_type, 'application/pdf'), COALESCE(page_count, 0),
		       COALESCE(text_page_count, 0), status, COALESCE(validation_status, 'PENDING'),
		       COALESCE(processing_status, 'PENDING'), COALESCE(pii_detected, FALSE), summary, created_at, updated_at
		FROM documents WHERE id = $1 AND ($2 = '' OR $2 = '00000000-0000-0000-0000-000000000001' OR user_id = $2::uuid OR user_id = '00000000-0000-0000-0000-000000000001'::uuid OR user_id IS NULL)`

	lErr := DB.QueryRow(lSqlStr, pDocumentID, pUserID).Scan(
		&lDoc.ID, &lDoc.UserID, &lDoc.Title, &lDoc.Filename, &lStoredFilename, &lFileHash,
		&lFileSize, &lMimeType, &lPageCount, &lTextPageCount, &lDoc.Status, &lValidationStatus,
		&lProcessingStatus, &lPiiDetected, &lSummaryJSON, &lDoc.CreatedAt, &lUpdatedAt,
	)

	if lErr != nil {
		if errors.Is(lErr, sql.ErrNoRows) {
			return nil, nil
		}
		pDebug.Log(utils.Elog, "REPO_DOC_06", "Error fetching document summary:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}

	lDoc.StoredFilename = lStoredFilename.String
	lDoc.FileHash = lFileHash.String
	lDoc.FileSize = lFileSize.Int64
	lDoc.MimeType = lMimeType.String
	lDoc.PageCount = int(lPageCount.Int32)
	lDoc.TextPageCount = int(lTextPageCount.Int32)
	lDoc.ValidationStatus = lValidationStatus.String
	lDoc.ProcessingStatus = lProcessingStatus.String
	lDoc.PIIDetected = lPiiDetected.Bool
	if lUpdatedAt.Valid {
		lDoc.UpdatedAt = lUpdatedAt.Time
	} else {
		lDoc.UpdatedAt = lDoc.CreatedAt
	}

	if len(lSummaryJSON) > 0 {
		var lSummary models.DocumentSummary
		if lUerr := json.Unmarshal(lSummaryJSON, &lSummary); lUerr == nil {
			lDoc.Summary = lSummary
		}
	}

	return &lDoc, nil
}

func GetDocumentInsights(pDebug *utils.HelperStruct, pDocumentID, pUserID string) (*models.DocumentInsights, error) {
	pDebug.Log(utils.Statement, "GetDocumentInsights(+)")
	defer pDebug.Log(utils.Statement, "GetDocumentInsights(-)")

	lDoc, lErr := GetDocumentSummary(pDebug, pDocumentID, pUserID)
	if lErr != nil {
		return nil, utils.ErrReturn(lErr)
	}
	if lDoc == nil {
		return nil, nil
	}

	// Count chunks and aggregate character/word metrics
	lSqlChunks := `
		SELECT COUNT(*), COALESCE(SUM(LENGTH(text)), 0), COALESCE(SUM(array_length(regexp_split_to_array(text, '\s+'), 1)), 0)
		FROM chunks WHERE document_id = $1`

	var lChunkCount, lTotalChars, lTotalWords int
	lErr = DB.QueryRow(lSqlChunks, pDocumentID).Scan(&lChunkCount, &lTotalChars, &lTotalWords)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_07", "Error counting chunks and characters:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}

	// Fetch validation audits
	lAudits, lErr := GetValidationAudits(pDebug, pDocumentID)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_08", "Error fetching audits for insights:", lErr.Error())
		lAudits = []models.DocumentValidationAudit{}
	}

	lInsights := &models.DocumentInsights{
		DocumentID:       lDoc.ID,
		Title:            lDoc.Title,
		Filename:         lDoc.Filename,
		FileSize:         lDoc.FileSize,
		MimeType:         lDoc.MimeType,
		PageCount:        lDoc.PageCount,
		TextPageCount:    lDoc.TextPageCount,
		TotalCharacters:  lTotalChars,
		TotalWords:       lTotalWords,
		TotalChunks:      lChunkCount,
		Status:           lDoc.Status,
		ValidationStatus: lDoc.ValidationStatus,
		ProcessingStatus: lDoc.ProcessingStatus,
		PIIDetected:      lDoc.PIIDetected,
		ValidationChecks: lAudits,
		CreatedAt:        lDoc.CreatedAt,
		UpdatedAt:        lDoc.UpdatedAt,
	}

	return lInsights, nil
}

func GetLatestDocumentByUser(pDebug *utils.HelperStruct, pUserID string) (*models.Document, error) {
	pDebug.Log(utils.Statement, "GetLatestDocumentByUser(+)")
	defer pDebug.Log(utils.Statement, "GetLatestDocumentByUser(-)")

	var lDoc models.Document
	var lSummaryJSON []byte
	var lStoredFilename, lFileHash, lMimeType, lValidationStatus, lProcessingStatus sql.NullString
	var lFileSize sql.NullInt64
	var lPageCount, lTextPageCount sql.NullInt32
	var lPiiDetected sql.NullBool
	var lUpdatedAt sql.NullTime

	lSqlStr := `
		SELECT id, user_id, title, filename, COALESCE(stored_file_name, ''), COALESCE(file_hash, ''),
		       COALESCE(file_size, 0), COALESCE(mime_type, 'application/pdf'), COALESCE(page_count, 0),
		       COALESCE(text_page_count, 0), status, COALESCE(validation_status, 'PENDING'),
		       COALESCE(processing_status, 'PENDING'), COALESCE(pii_detected, FALSE), summary, created_at, updated_at
		FROM documents WHERE ($1 = '' OR $1 = '00000000-0000-0000-0000-000000000001' OR user_id = $1::uuid OR user_id IS NULL) ORDER BY created_at DESC LIMIT 1`

	lErr := DB.QueryRow(lSqlStr, pUserID).Scan(
		&lDoc.ID, &lDoc.UserID, &lDoc.Title, &lDoc.Filename, &lStoredFilename, &lFileHash,
		&lFileSize, &lMimeType, &lPageCount, &lTextPageCount, &lDoc.Status, &lValidationStatus,
		&lProcessingStatus, &lPiiDetected, &lSummaryJSON, &lDoc.CreatedAt, &lUpdatedAt,
	)

	if lErr != nil {
		if errors.Is(lErr, sql.ErrNoRows) {
			return nil, nil
		}
		pDebug.Log(utils.Elog, "REPO_DOC_09", "Error fetching latest document:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}

	lDoc.StoredFilename = lStoredFilename.String
	lDoc.FileHash = lFileHash.String
	lDoc.FileSize = lFileSize.Int64
	lDoc.MimeType = lMimeType.String
	lDoc.PageCount = int(lPageCount.Int32)
	lDoc.TextPageCount = int(lTextPageCount.Int32)
	lDoc.ValidationStatus = lValidationStatus.String
	lDoc.ProcessingStatus = lProcessingStatus.String
	lDoc.PIIDetected = lPiiDetected.Bool
	if lUpdatedAt.Valid {
		lDoc.UpdatedAt = lUpdatedAt.Time
	} else {
		lDoc.UpdatedAt = lDoc.CreatedAt
	}

	if len(lSummaryJSON) > 0 {
		var lSummary models.DocumentSummary
		if lUerr := json.Unmarshal(lSummaryJSON, &lSummary); lUerr == nil {
			lDoc.Summary = lSummary
		}
	}

	return &lDoc, nil
}

func GetDocumentsByUser(pDebug *utils.HelperStruct, pUserID string, pLimit int) ([]*models.Document, error) {
	pDebug.Log(utils.Statement, "GetDocumentsByUser(+)", pUserID, pLimit)
	defer pDebug.Log(utils.Statement, "GetDocumentsByUser(-)")

	if pLimit <= 0 || pLimit > 3 {
		pLimit = 3
	}

	// Strictly ensure no more than 3 documents are maintained in the database for this user
	if lPruneErr := PruneOldUserDocuments(pDebug, pUserID, 3); lPruneErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_PRUNE", "Warning: could not prune excess user documents:", lPruneErr.Error())
	}

	lSqlStr := `
		SELECT d.id, d.user_id, d.title, d.filename, COALESCE(d.stored_file_name, ''), COALESCE(d.file_hash, ''),
		       COALESCE(d.file_size, 0), COALESCE(d.mime_type, 'application/pdf'), COALESCE(d.page_count, 0),
		       COALESCE(d.text_page_count, 0), d.status, COALESCE(d.validation_status, 'PENDING'),
		       COALESCE(d.processing_status, 'PENDING'), COALESCE(d.pii_detected, FALSE), d.summary, d.created_at, d.updated_at,
		       COALESCE((SELECT COUNT(*) FROM chunks c WHERE c.document_id = d.id), 0) AS chunk_count
		FROM documents d
		WHERE ($1 = '' OR $1 = '00000000-0000-0000-0000-000000000001' OR d.user_id = $1::uuid OR d.user_id IS NULL)
		ORDER BY d.created_at DESC
		LIMIT $2`

	lRows, lErr := DB.Query(lSqlStr, pUserID, pLimit)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_10", "Error querying documents by user:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}
	defer lRows.Close()

	var lDocs []*models.Document
	for lRows.Next() {
		var lDoc models.Document
		var lSummaryJSON []byte
		var lStoredFilename, lFileHash, lMimeType, lValidationStatus, lProcessingStatus sql.NullString
		var lFileSize sql.NullInt64
		var lPageCount, lTextPageCount, lChunkCount sql.NullInt32
		var lPiiDetected sql.NullBool
		var lUpdatedAt sql.NullTime

		if lScanErr := lRows.Scan(
			&lDoc.ID, &lDoc.UserID, &lDoc.Title, &lDoc.Filename, &lStoredFilename, &lFileHash,
			&lFileSize, &lMimeType, &lPageCount, &lTextPageCount, &lDoc.Status, &lValidationStatus,
			&lProcessingStatus, &lPiiDetected, &lSummaryJSON, &lDoc.CreatedAt, &lUpdatedAt,
			&lChunkCount,
		); lScanErr != nil {
			pDebug.Log(utils.Elog, "REPO_DOC_11", "Error scanning document row:", lScanErr.Error())
			return nil, utils.ErrReturn(lScanErr)
		}

		lDoc.StoredFilename = lStoredFilename.String
		lDoc.FileHash = lFileHash.String
		lDoc.FileSize = lFileSize.Int64
		lDoc.MimeType = lMimeType.String
		lDoc.PageCount = int(lPageCount.Int32)
		lDoc.TextPageCount = int(lTextPageCount.Int32)
		lDoc.ValidationStatus = lValidationStatus.String
		lDoc.ProcessingStatus = lProcessingStatus.String
		lDoc.PIIDetected = lPiiDetected.Bool
		lDoc.TotalChunks = int(lChunkCount.Int32)
		if lUpdatedAt.Valid {
			lDoc.UpdatedAt = lUpdatedAt.Time
		} else {
			lDoc.UpdatedAt = lDoc.CreatedAt
		}

		if len(lSummaryJSON) > 0 {
			var lSummary models.DocumentSummary
			if lUerr := json.Unmarshal(lSummaryJSON, &lSummary); lUerr == nil {
				lDoc.Summary = lSummary
			}
		}

		lDocs = append(lDocs, &lDoc)
	}

	return lDocs, nil
}

func DeleteDocument(pDebug *utils.HelperStruct, pDocumentID, pUserID string) error {
	pDebug.Log(utils.Statement, "DeleteDocument(+)", pDocumentID)
	defer pDebug.Log(utils.Statement, "DeleteDocument(-)")

	// 1. Fetch stored file name to delete file on disk
	var lStoredFilename sql.NullString
	lSqlFind := "SELECT stored_file_name FROM documents WHERE id = $1 AND ($2 = '' OR $2 = '00000000-0000-0000-0000-000000000001' OR user_id = $2::uuid OR user_id IS NULL)"
	lErr := DB.QueryRow(lSqlFind, pDocumentID, pUserID).Scan(&lStoredFilename)
	if lErr != nil {
		if errors.Is(lErr, sql.ErrNoRows) {
			return errors.New("document not found or unauthorized")
		}
		pDebug.Log(utils.Elog, "REPO_DOC_12", "Error finding document for deletion:", lErr.Error())
		return utils.ErrReturn(lErr)
	}

	// 2. Remove physical file from disk
	if lStoredFilename.Valid && lStoredFilename.String != "" {
		lUploadDir := filepath.FromSlash(config.AppConfig.Storage.UploadDir)
		lFilePath := filepath.Join(lUploadDir, lStoredFilename.String)
		if lRemErr := os.Remove(lFilePath); lRemErr != nil && !os.IsNotExist(lRemErr) {
			pDebug.Log(utils.Elog, "REPO_DOC_13", "Warning: could not delete file from storage:", lRemErr.Error())
		}
	}

	// 3. Delete document row (cascades to chunks, document_validation, and query_logs)
	lSqlDel := "DELETE FROM documents WHERE id = $1 AND ($2 = '' OR $2 = '00000000-0000-0000-0000-000000000001' OR user_id = $2::uuid OR user_id IS NULL)"
	_, lErr = DB.Exec(lSqlDel, pDocumentID, pUserID)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_14", "Error deleting document from DB:", lErr.Error())
		return utils.ErrReturn(lErr)
	}

	return nil
}

func PruneOldUserDocuments(pDebug *utils.HelperStruct, pUserID string, pKeepCount int) error {
	pDebug.Log(utils.Statement, "PruneOldUserDocuments(+)", pKeepCount)
	defer pDebug.Log(utils.Statement, "PruneOldUserDocuments(-)")

	if pKeepCount < 0 {
		pKeepCount = 0
	}

	// Select document IDs that exceed pKeepCount (ordered newest first)
	lSqlStr := `
		SELECT id FROM documents 
		WHERE ($1 = '' OR $1 = '00000000-0000-0000-0000-000000000001' OR user_id = $1::uuid OR user_id IS NULL) 
		ORDER BY created_at DESC 
		OFFSET $2`

	lRows, lErr := DB.Query(lSqlStr, pUserID, pKeepCount)
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_15", "Error querying documents to prune:", lErr.Error())
		return utils.ErrReturn(lErr)
	}
	defer lRows.Close()

	var lDocIDs []string
	for lRows.Next() {
		var lID string
		if lScanErr := lRows.Scan(&lID); lScanErr == nil {
			lDocIDs = append(lDocIDs, lID)
		}
	}

	for _, lID := range lDocIDs {
		pDebug.Log(utils.Statement, "Pruning excess document:", lID)
		if lDelErr := DeleteDocument(pDebug, lID, pUserID); lDelErr != nil {
			pDebug.Log(utils.Elog, "REPO_DOC_16", "Failed pruning document:", lID, lDelErr.Error())
		}
	}

	return nil
}

func FlushAllDocumentsAndChunks(pDebug *utils.HelperStruct) (int64, error) {
	pDebug.Log(utils.Statement, "FlushAllDocumentsAndChunks(+)")
	defer pDebug.Log(utils.Statement, "FlushAllDocumentsAndChunks(-)")

	// 1. Delete all physical files in upload directory
	lUploadDir := filepath.FromSlash(config.AppConfig.Storage.UploadDir)
	if lEntries, lErr := os.ReadDir(lUploadDir); lErr == nil {
		for _, lEntry := range lEntries {
			if !lEntry.IsDir() {
				_ = os.Remove(filepath.Join(lUploadDir, lEntry.Name()))
			}
		}
	}

	// 2. Delete all records from documents (cascades to chunks, audits, and query_logs)
	lRes, lErr := DB.Exec("DELETE FROM documents")
	if lErr != nil {
		pDebug.Log(utils.Elog, "REPO_DOC_17", "Error flushing documents from DB:", lErr.Error())
		return 0, utils.ErrReturn(lErr)
	}

	lRowsAffected, _ := lRes.RowsAffected()
	pDebug.Log(utils.Statement, "FlushAllDocumentsAndChunks completed. Deleted documents count:", lRowsAffected)
	return lRowsAffected, nil
}
