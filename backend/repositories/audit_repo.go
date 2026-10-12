package repositories

import (
	"backend/models"
	"backend/utils"
)

func RecordValidationAudit(pDebug *utils.HelperStruct, pDocID string, pCheckName string, pStatus string, pMessage string) error {
	pDebug.Log(utils.Statement, "RecordValidationAudit(+)")
	defer pDebug.Log(utils.Statement, "RecordValidationAudit(-)")

	lSqlStr := `
		INSERT INTO document_validation (document_id, check_name, status, message)
		VALUES ($1, $2, $3, $4)`

	_, lErr := DB.Exec(lSqlStr, pDocID, pCheckName, pStatus, pMessage)
	if lErr != nil {
		pDebug.Log(utils.Elog, "AUDIT_VAL_01", "Failed to insert validation audit:", lErr.Error())
		return utils.ErrReturn(lErr)
	}

	return nil
}

func GetValidationAudits(pDebug *utils.HelperStruct, pDocID string) ([]models.DocumentValidationAudit, error) {
	pDebug.Log(utils.Statement, "GetValidationAudits(+)")
	defer pDebug.Log(utils.Statement, "GetValidationAudits(-)")

	lSqlStr := `
		SELECT id, document_id, check_name, status, COALESCE(message, ''), created_at
		FROM document_validation
		WHERE document_id = $1
		ORDER BY created_at ASC`

	lRows, lErr := DB.Query(lSqlStr, pDocID)
	if lErr != nil {
		pDebug.Log(utils.Elog, "AUDIT_VAL_02", "Failed to query validation audits:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}
	defer lRows.Close()

	var lAudits []models.DocumentValidationAudit
	for lRows.Next() {
		var lAudit models.DocumentValidationAudit
		if lScanErr := lRows.Scan(&lAudit.ID, &lAudit.DocumentID, &lAudit.CheckName, &lAudit.Status, &lAudit.Message, &lAudit.CreatedAt); lScanErr != nil {
			pDebug.Log(utils.Elog, "AUDIT_VAL_03", "Failed to scan audit row:", lScanErr.Error())
			return nil, utils.ErrReturn(lScanErr)
		}
		lAudits = append(lAudits, lAudit)
	}

	return lAudits, nil
}

func RecordQueryAudit(pDebug *utils.HelperStruct, pLog *models.QueryLog) error {
	pDebug.Log(utils.Statement, "RecordQueryAudit(+)")
	defer pDebug.Log(utils.Statement, "RecordQueryAudit(-)")

	lSqlStr := `
		INSERT INTO query_logs (
			user_id, document_id, original_query, normalized_query, query_type,
			retrieval_attempt, top_similarity, context_valid, llm_provider,
			llm_attempt, answer_valid, fallback_used
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, lErr := DB.Exec(
		lSqlStr,
		pLog.UserID, pLog.DocumentID, pLog.OriginalQuery, pLog.NormalizedQuery, pLog.QueryType,
		pLog.RetrievalAttempt, pLog.TopSimilarity, pLog.ContextValid, pLog.LLMProvider,
		pLog.LLMAttempt, pLog.AnswerValid, pLog.FallbackUsed,
	)
	if lErr != nil {
		pDebug.Log(utils.Elog, "AUDIT_QRY_01", "Failed to insert query log:", lErr.Error())
		return utils.ErrReturn(lErr)
	}

	return nil
}
