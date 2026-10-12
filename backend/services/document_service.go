package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"backend/ai"
	"backend/config"
	"backend/models"
	"backend/pdf"
	"backend/repositories"
	"backend/utils"
)

func ProcessUploadedDocument(
	pDebug *utils.HelperStruct,
	pUserID, pTitle, pFilename, pStoredFilename, pFilePath, pFileHash, pMimeType string,
	pFileSize int64,
) (string, error) {
	pDebug.Log(utils.Statement, "ProcessUploadedDocument(+)")
	defer pDebug.Log(utils.Statement, "ProcessUploadedDocument(-)")

	// 0. Enforce maximum 3 documents per user policy: prune older documents to leave room for the new one
	if lPruneErr := repositories.PruneOldUserDocuments(pDebug, pUserID, 2); lPruneErr != nil {
		pDebug.Log(utils.Elog, "SRV_DOC_PRUNE", "Warning during document pruning:", lPruneErr.Error())
	}

	// 1. Create document record with pending status
	lDoc := &models.Document{
		UserID:           pUserID,
		Title:            pTitle,
		Filename:         pFilename,
		StoredFilename:   pStoredFilename,
		FileHash:         pFileHash,
		FileSize:         pFileSize,
		MimeType:         pMimeType,
		Status:           "PROCESSING",
		ValidationStatus: "IN_PROGRESS",
		ProcessingStatus: "PROCESSING",
		PIIDetected:      false,
	}

	lDocID, lErr := repositories.CreateDocument(pDebug, lDoc)
	if lErr != nil {
		return "", utils.ErrReturn(lErr)
	}

	// Record initial signature check audit
	repositories.RecordValidationAudit(pDebug, lDocID, "PDF_SIGNATURE", "PASS", "Valid %PDF- signature verified")

	// 2. Process asynchronously in background
	go func(pID, pFP, pStored string) {
		lBgDebug := new(utils.HelperStruct)
		lBgDebug.Init()
		lBgDebug.Log(utils.Statement, "BackgroundPDFProcess(+)", pID)

		lErr := processPDFWorkflow(lBgDebug, pID, pFP)
		if lErr != nil {
			lBgDebug.Log(utils.Elog, "SRV_DOC_01", "Processing failed:", lErr.Error())
			repositories.UpdateDocumentStatus(lBgDebug, pID, "ERROR")
			repositories.RecordValidationAudit(lBgDebug, pID, "PROCESSING_ERROR", "FAIL", lErr.Error())
		} else {
			repositories.UpdateDocumentStatus(lBgDebug, pID, "COMPLETED")
		}

		// Phase 6: DO NOT remove the original file; keep it safely in storage
		lBgDebug.Log(utils.Statement, "BackgroundPDFProcess(-)")
	}(lDocID, pFilePath, pStoredFilename)

	return lDocID, nil
}

func processPDFWorkflow(pDebug *utils.HelperStruct, pDocumentID, pFilePath string) error {
	pDebug.Log(utils.Statement, "processPDFWorkflow(+)")
	defer pDebug.Log(utils.Statement, "processPDFWorkflow(-)")

	lMaxPages := config.AppConfig.Validation.MaxPages

	// 1. Content validation: page counts, extractable text check
	lValRes, lValErr := pdf.ValidatePDFContent(pDebug, pFilePath, lMaxPages)
	if lValErr != nil {
		repositories.RecordValidationAudit(pDebug, pDocumentID, "TEXT_CONTENT", "FAIL", lValErr.Error())
		return utils.ErrReturn(lValErr)
	}

	for _, lChk := range lValRes.AuditChecks {
		repositories.RecordValidationAudit(pDebug, pDocumentID, lChk.CheckName, lChk.Status, lChk.Message)
	}

	// 2. Full ingestion with text standardization, header/footer removal, and PII redaction
	lExtractRes, lExtErr := pdf.ExtractAndProcessPDF(pDebug, pFilePath, pDocumentID)
	if lExtErr != nil {
		repositories.RecordValidationAudit(pDebug, pDocumentID, "EXTRACTION", "FAIL", lExtErr.Error())
		return utils.ErrReturn(lExtErr)
	}

	// Record PII scan audit
	if lExtractRes.PIIDetected {
		repositories.RecordValidationAudit(pDebug, pDocumentID, "PII_SCAN", "WARN", fmt.Sprintf("PII detected and redacted: %s", strings.Join(lExtractRes.RedactedPII, ", ")))
	} else {
		repositories.RecordValidationAudit(pDebug, pDocumentID, "PII_SCAN", "PASS", "No sensitive PII detected")
	}
	repositories.RecordValidationAudit(pDebug, pDocumentID, "CONTENT_STANDARDIZE", "PASS", fmt.Sprintf("Sanitized %d chunks", len(lExtractRes.Chunks)))

	// 3. Generate embeddings and store chunks
	var lChunksWithVectors []repositories.ChunkWithVector
	for _, lChunk := range lExtractRes.Chunks {
		lEmbedding, lErr := ai.GenerateEmbedding(pDebug, lChunk.Text)
		if lErr != nil {
			repositories.RecordValidationAudit(pDebug, pDocumentID, "EMBEDDING", "FAIL", lErr.Error())
			return utils.ErrReturn(lErr)
		}
		lChunksWithVectors = append(lChunksWithVectors, repositories.ChunkWithVector{
			Chunk:     lChunk,
			Embedding: lEmbedding,
		})
	}

	lErr := repositories.InsertChunks(pDebug, lChunksWithVectors)
	if lErr != nil {
		return utils.ErrReturn(lErr)
	}
	repositories.RecordValidationAudit(pDebug, pDocumentID, "EMBEDDING", "PASS", fmt.Sprintf("Embedded %d chunks successfully", len(lChunksWithVectors)))

	// 4. Update document metadata
	lMetaErr := repositories.UpdateDocumentMetadata(
		pDebug,
		pDocumentID,
		lExtractRes.TotalPages,
		lExtractRes.TextPages,
		lExtractRes.PIIDetected,
		"PASSED",
		"COMPLETED",
	)
	if lMetaErr != nil {
		pDebug.Log(utils.Elog, "SRV_DOC_META", lMetaErr.Error())
	}

	// 5. Generate structured summary
	lSystemPrompt := `You are an educational project review assistant analyzing a student's project report. Generate a structured JSON summary based on the provided project text.
The JSON must strictly conform to these keys:
title, objective, problem, solution, technologies (string array), algorithms (string array), methodology, results, advantages (string array), limitations (string array), future_scope.
Return ONLY valid JSON without markdown wrapping.`

	lSampleText := lExtractRes.SanitizedDoc
	if len(lSampleText) > 10000 {
		lSampleText = lSampleText[:10000]
	}

	lSummaryJSON, lErr := ai.GenerateCompletion(pDebug, lSystemPrompt, lSampleText, 0.3)
	if lErr != nil {
		pDebug.Log(utils.Elog, "SRV_DOC_02", "Summary generation failed:", lErr.Error())
	} else {
		lCleaned := strings.TrimSpace(lSummaryJSON)
		if strings.HasPrefix(lCleaned, "```json") {
			lCleaned = strings.TrimPrefix(lCleaned, "```json")
		} else if strings.HasPrefix(lCleaned, "```") {
			lCleaned = strings.TrimPrefix(lCleaned, "```")
		}
		if strings.HasSuffix(lCleaned, "```") {
			lCleaned = strings.TrimSuffix(lCleaned, "```")
		}
		lCleaned = strings.TrimSpace(lCleaned)

		var lSummary models.DocumentSummary
		if lUnmarshalErr := json.Unmarshal([]byte(lCleaned), &lSummary); lUnmarshalErr == nil {
			repositories.UpdateDocumentSummary(pDebug, pDocumentID, &lSummary)
		} else {
			pDebug.Log(utils.Elog, "SRV_DOC_03", "Summary JSON parse failed:", lUnmarshalErr.Error(), lCleaned)
		}
	}

	return nil
}
