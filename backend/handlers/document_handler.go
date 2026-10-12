package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"backend/config"
	"backend/models"
	"backend/pdf"
	"backend/repositories"
	"backend/services"
	"backend/utils"
)

const DefaultGuestUserID = "00000000-0000-0000-0000-000000000001"

func resolveUserID(pReq *http.Request) string {
	if lSessID := pReq.Header.Get("X-Session-ID"); lSessID != "" {
		if _, lErr := uuid.Parse(lSessID); lErr == nil {
			return lSessID
		}
	}
	if lUserID, ok := pReq.Context().Value("user_id").(string); ok && lUserID != "" {
		return lUserID
	}
	return DefaultGuestUserID
}

func extractDocID(pReq *http.Request) string {
	if pReq != nil {
		if lPathVal := pReq.PathValue("id"); lPathVal != "" {
			return lPathVal
		}
	}
	lDocID := pReq.URL.Query().Get("id")
	if lDocID == "" {
		lDocID = pReq.URL.Query().Get("document_id")
	}
	if lDocID == "" {
		lParts := strings.Split(strings.Trim(pReq.URL.Path, "/"), "/")
		for i, part := range lParts {
			if part == "documents" && i+1 < len(lParts) {
				candidate := lParts[i+1]
				if candidate != "summary" && candidate != "insights" && candidate != "ask" && candidate != "latest" && candidate != "delete" && candidate != "upload" {
					lDocID = candidate
					break
				}
			}
		}
		if lDocID == "" && len(lParts) >= 4 {
			lDocID = lParts[3]
		}
	}
	return lDocID
}

func UploadDocument(Resp http.ResponseWriter, Req *http.Request) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()
	lDebug.Log(utils.Statement, "UploadDocument(+)", Req.Method)
	defer lDebug.Log(utils.Statement, "UploadDocument(-)")

	Resp.Header().Set("Access-Control-Allow-Origin", "*")
	Resp.Header().Set("Access-Control-Allow-Credentials", "true")
	Resp.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	Resp.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-ID")

	if Req.Method == http.MethodOptions {
		return
	}

	if Req.Method != http.MethodPost {
		fmt.Fprint(Resp, utils.GetErrorString("DOC01", "Invalid Http Method"))
		return
	}

	lUserID := resolveUserID(Req)

	// Max limit from config (e.g. 20MB)
	lMaxMB := config.AppConfig.Validation.MaxFileSizeMB
	if lMaxMB <= 0 {
		lMaxMB = 20
	}
	lMaxBytes := int64(lMaxMB) * 1024 * 1024

	// Limit request reader to avoid memory exhaustion
	Req.Body = http.MaxBytesReader(Resp, Req.Body, lMaxBytes+1024)
	if lErr := Req.ParseMultipartForm(lMaxBytes); lErr != nil {
		lDebug.Log(utils.Elog, "DOC03", "Request parsing error:", lErr.Error())
		fmt.Fprint(Resp, utils.GetErrorString("DOC03", "File exceeds maximum size limit"))
		return
	}

	lFile, lHandler, lErr := Req.FormFile("file")
	if lErr != nil {
		lDebug.Log(utils.Elog, "DOC04", lErr.Error())
		fmt.Fprint(Resp, utils.GetErrorString("DOC04", "Error retrieving file from form"))
		return
	}
	defer lFile.Close()

	// Phase 1 & 2: Secure PDF File Validation (signature, size, MIME, hash)
	lValRes, lValErr := pdf.ValidateUploadedFile(lDebug, lFile, lHandler, lMaxBytes)
	if lValErr != nil {
		lDebug.Log(utils.Elog, "DOC05", "Validation failed:", lValRes.ErrorMessage)
		fmt.Fprint(Resp, utils.GetErrorString("DOC05", lValRes.ErrorMessage))
		return
	}

	// Save file locally to storage
	lUploadDir := filepath.FromSlash(config.AppConfig.Storage.UploadDir)
	if lErr := os.MkdirAll(lUploadDir, 0777); lErr != nil {
		lDebug.Log(utils.Elog, "DOC06", lErr.Error())
		fmt.Fprint(Resp, utils.GetErrorString("DOC06", "Storage service unavailable"))
		return
	}

	lUniqueFilename := uuid.New().String() + ".pdf"
	lFilePath := filepath.Join(lUploadDir, lUniqueFilename)

	lDst, lErr := os.OpenFile(lFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if lErr != nil {
		lDebug.Log(utils.Elog, "DOC07", fmt.Sprintf("Path: %s, Err: %s", lFilePath, lErr.Error()))
		fmt.Fprint(Resp, utils.GetErrorString("DOC07", "Failed to store uploaded file: "+lErr.Error()))
		return
	}
	defer lDst.Close()

	if _, lErr := io.Copy(lDst, lFile); lErr != nil {
		lDebug.Log(utils.Elog, "DOC08", lErr.Error())
		fmt.Fprint(Resp, utils.GetErrorString("DOC08", "Failed to write uploaded file"))
		return
	}

	lTitle := Req.FormValue("title")
	if lTitle == "" {
		lTitle = lHandler.Filename
	}

	// Trigger processing pipeline
	lDocID, lErr := services.ProcessUploadedDocument(
		lDebug, lUserID, lTitle, lHandler.Filename, lUniqueFilename,
		lFilePath, lValRes.FileHash, lValRes.MimeType, lValRes.FileSize,
	)
	if lErr != nil {
		lDebug.Log(utils.Elog, "DOC09", lErr.Error())
		fmt.Fprint(Resp, utils.GetErrorString("DOC09", "Document processing initialization failed"))
		return
	}

	lRespData := map[string]interface{}{
		"status":      "S",
		"document_id": lDocID,
		"message":     "Upload successful, validation passed, processing started",
		"file_hash":   lValRes.FileHash,
	}
	json.NewEncoder(Resp).Encode(lRespData)
}

func GetSummary(Resp http.ResponseWriter, Req *http.Request) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()
	lDebug.Log(utils.Statement, "GetSummary(+)", Req.Method)
	defer lDebug.Log(utils.Statement, "GetSummary(-)")

	Resp.Header().Set("Access-Control-Allow-Origin", "*")
	Resp.Header().Set("Access-Control-Allow-Credentials", "true")
	Resp.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	Resp.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-ID")

	if Req.Method == http.MethodOptions {
		return
	}

	if Req.Method != http.MethodGet {
		fmt.Fprint(Resp, utils.GetErrorString("DOC10", "Invalid Http Method"))
		return
	}

	lUserID := resolveUserID(Req)

	lDocID := extractDocID(Req)
	if lDocID == "" {
		lDebug.Log(utils.Elog, "DOC12", "Missing document ID")
		fmt.Fprint(Resp, utils.GetErrorString("DOC12", "Missing document ID"))
		return
	}

	lDoc, lErr := repositories.GetDocumentSummary(lDebug, lDocID, lUserID)
	if lErr != nil {
		lDebug.Log(utils.Elog, "DOC13", lErr.Error())
		fmt.Fprint(Resp, utils.GetErrorString("DOC13", "Failed to retrieve summary"))
		return
	}

	if lDoc == nil {
		fmt.Fprint(Resp, utils.GetErrorString("DOC14", "Document not found"))
		return
	}

	lRespData := map[string]interface{}{
		"status":   "S",
		"document": lDoc,
	}
	json.NewEncoder(Resp).Encode(lRespData)
}

func GetInsights(Resp http.ResponseWriter, Req *http.Request) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()
	lDebug.Log(utils.Statement, "GetInsights(+)", Req.Method)
	defer lDebug.Log(utils.Statement, "GetInsights(-)")

	Resp.Header().Set("Access-Control-Allow-Origin", "*")
	Resp.Header().Set("Access-Control-Allow-Credentials", "true")
	Resp.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	Resp.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-ID")

	if Req.Method == http.MethodOptions {
		return
	}

	if Req.Method != http.MethodGet {
		fmt.Fprint(Resp, utils.GetErrorString("DOC15", "Invalid Http Method"))
		return
	}

	lUserID := resolveUserID(Req)

	lDocID := extractDocID(Req)
	if lDocID == "" {
		lDebug.Log(utils.Elog, "DOC17", "Missing document ID")
		fmt.Fprint(Resp, utils.GetErrorString("DOC17", "Missing document ID"))
		return
	}

	lInsights, lErr := repositories.GetDocumentInsights(lDebug, lDocID, lUserID)
	if lErr != nil {
		lDebug.Log(utils.Elog, "DOC18", lErr.Error())
		fmt.Fprint(Resp, utils.GetErrorString("DOC18", "Failed to retrieve document insights"))
		return
	}

	if lInsights == nil {
		fmt.Fprint(Resp, utils.GetErrorString("DOC19", "Document not found"))
		return
	}

	lRespData := map[string]interface{}{
		"status":   "S",
		"insights": lInsights,
	}
	json.NewEncoder(Resp).Encode(lRespData)
}

type AskReq struct {
	DocumentID string `json:"document_id,omitempty"`
	Question   string `json:"question"`
	Level      int    `json:"level"`
}

func Ask(Resp http.ResponseWriter, Req *http.Request) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()
	lDebug.Log(utils.Statement, "Ask(+)", Req.Method)
	defer lDebug.Log(utils.Statement, "Ask(-)")

	Resp.Header().Set("Access-Control-Allow-Origin", "*")
	Resp.Header().Set("Access-Control-Allow-Credentials", "true")
	Resp.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	Resp.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-ID")

	if Req.Method == http.MethodOptions {
		return
	}

	if Req.Method != http.MethodPost {
		fmt.Fprint(Resp, utils.GetErrorString("DOC20", "Invalid Http Method"))
		return
	}

	lUserID := resolveUserID(Req)

	var lReqData AskReq
	if lErr := json.NewDecoder(Req.Body).Decode(&lReqData); lErr != nil {
		fmt.Fprint(Resp, utils.GetErrorString("DOC23", "Invalid request body"))
		return
	}

	lDocID := lReqData.DocumentID
	if lDocID == "" {
		lDocID = extractDocID(Req)
	}
	if lDocID == "" {
		lDebug.Log(utils.Elog, "DOC22", "Missing document ID")
		fmt.Fprint(Resp, utils.GetErrorString("DOC22", "Missing document ID"))
		return
	}

	if strings.TrimSpace(lReqData.Question) == "" {
		fmt.Fprint(Resp, utils.GetErrorString("DOC24", "Question cannot be empty"))
		return
	}

	if lReqData.Level != 1 && lReqData.Level != 2 {
		lReqData.Level = 1
	}

	lAnswer, lErr := services.AskQuestion(lDebug, lDocID, lUserID, lReqData.Question, lReqData.Level)
	if lErr != nil {
		lDebug.Log(utils.Elog, "DOC25", lErr.Error())
		fmt.Fprint(Resp, utils.GetErrorString("DOC25", "Failed to process question"))
		return
	}

	json.NewEncoder(Resp).Encode(lAnswer)
}

func GetLatestDocument(Resp http.ResponseWriter, Req *http.Request) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()
	lDebug.Log(utils.Statement, "GetLatestDocument(+)", Req.Method)
	defer lDebug.Log(utils.Statement, "GetLatestDocument(-)")

	Resp.Header().Set("Access-Control-Allow-Origin", "*")
	Resp.Header().Set("Access-Control-Allow-Credentials", "true")
	Resp.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	Resp.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-ID")

	if Req.Method == http.MethodOptions {
		return
	}

	if Req.Method != http.MethodGet {
		fmt.Fprint(Resp, utils.GetErrorString("DOC26", "Invalid Http Method"))
		return
	}

	lUserID := resolveUserID(Req)

	lDoc, lErr := repositories.GetLatestDocumentByUser(lDebug, lUserID)
	if lErr != nil {
		lDebug.Log(utils.Elog, "DOC28", lErr.Error())
		fmt.Fprint(Resp, utils.GetErrorString("DOC28", "Failed to retrieve latest document"))
		return
	}

	lRespData := map[string]interface{}{
		"status":   "S",
		"document": lDoc,
	}
	json.NewEncoder(Resp).Encode(lRespData)
}

func GetSystemInfo(Resp http.ResponseWriter, Req *http.Request) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()
	lDebug.Log(utils.Statement, "GetSystemInfo(+)", Req.Method)
	defer lDebug.Log(utils.Statement, "GetSystemInfo(-)")

	Resp.Header().Set("Access-Control-Allow-Origin", "*")
	Resp.Header().Set("Access-Control-Allow-Credentials", "true")
	Resp.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	Resp.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-ID")

	if Req.Method == http.MethodOptions {
		return
	}

	lRespData := map[string]interface{}{
		"status": "S",
		"system": map[string]interface{}{
			"version":  "v2.4",
			"cluster":  "us-east-prod-4",
			"pipeline": "RAG Grounded Neural Engine",
		},
		"ai": map[string]interface{}{
			"primary": map[string]interface{}{
				"provider":    "Google Gemini",
				"model":       config.AppConfig.AI.Model,
				"status":      "Available",
				"latency_avg": "340ms avg",
				"base_url":    config.AppConfig.AI.BaseURL,
			},
			"fallback": map[string]interface{}{
				"enabled":  config.AppConfig.Fallback.Enabled,
				"provider": "Ollama",
				"model":    config.AppConfig.Fallback.Model,
				"status":   "Ready",
				"mode":     "Local AI Fallback",
			},
			"embeddings": map[string]interface{}{
				"provider":  "Google Gemini",
				"model":     config.AppConfig.AI.EmbedModel,
				"dimension": config.AppConfig.AI.EmbedDimension,
				"status":    "Ready",
			},
			"retrieval": map[string]interface{}{
				"hybrid_search":        config.AppConfig.Retrieval.HybridSearch,
				"vector_weight":        config.AppConfig.Retrieval.VectorWeight,
				"keyword_weight":       config.AppConfig.Retrieval.KeywordWeight,
				"top_k":                config.AppConfig.Retrieval.TopK,
				"similarity_threshold": config.AppConfig.Retrieval.SimilarityThreshold,
			},
			"validation": map[string]interface{}{
				"max_file_size_mb": config.AppConfig.Validation.MaxFileSizeMB,
				"max_pages":        config.AppConfig.Validation.MaxPages,
			},
			"privacy": map[string]interface{}{
				"pii_filtering":  true,
				"remove_email":   config.AppConfig.Privacy.RemoveEmail,
				"remove_phone":   config.AppConfig.Privacy.RemovePhone,
				"remove_roll":    config.AppConfig.Privacy.RemoveRollNumber,
				"remove_aadhaar": config.AppConfig.Privacy.RemoveAadhaar,
			},
		},
	}
	json.NewEncoder(Resp).Encode(lRespData)
}

func GetDocuments(Resp http.ResponseWriter, Req *http.Request) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()
	lDebug.Log(utils.Statement, "GetDocuments(+)", Req.Method)
	defer lDebug.Log(utils.Statement, "GetDocuments(-)")

	Resp.Header().Set("Access-Control-Allow-Origin", "*")
	Resp.Header().Set("Access-Control-Allow-Credentials", "true")
	Resp.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	Resp.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-ID")

	if Req.Method == http.MethodOptions {
		return
	}

	if Req.Method != http.MethodGet {
		fmt.Fprint(Resp, utils.GetErrorString("DOC35", "Invalid Http Method"))
		return
	}

	lUserID := resolveUserID(Req)

	lDocs, lErr := repositories.GetDocumentsByUser(lDebug, lUserID, 3)
	if lErr != nil {
		lDebug.Log(utils.Elog, "DOC37", lErr.Error())
		fmt.Fprint(Resp, utils.GetErrorString("DOC37", "Failed to retrieve documents"))
		return
	}

	if lDocs == nil {
		lDocs = []*models.Document{}
	}

	lRespData := map[string]interface{}{
		"status":    "S",
		"documents": lDocs,
		"count":     len(lDocs),
		"max_limit": 3,
	}
	json.NewEncoder(Resp).Encode(lRespData)
}

func DeleteDocument(Resp http.ResponseWriter, Req *http.Request) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()
	lDebug.Log(utils.Statement, "DeleteDocument(+)", Req.Method)
	defer lDebug.Log(utils.Statement, "DeleteDocument(-)")

	Resp.Header().Set("Access-Control-Allow-Origin", "*")
	Resp.Header().Set("Access-Control-Allow-Credentials", "true")
	Resp.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	Resp.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-ID")

	if Req.Method == http.MethodOptions {
		return
	}

	if Req.Method != http.MethodDelete && Req.Method != http.MethodPost {
		lDebug.Log(utils.Elog, "DOC38", "Invalid Http Method:", Req.Method)
		fmt.Fprint(Resp, utils.GetErrorString("DOC38", "Invalid Http Method"))
		return
	}

	lUserID := resolveUserID(Req)

	// 1. Try extracting document ID from Path or URL query: ?id=... or ?document_id=...
	lDocID := Req.PathValue("id")
	if lDocID == "" {
		lDocID = Req.URL.Query().Get("id")
	}
	if lDocID == "" {
		lDocID = Req.URL.Query().Get("document_id")
	}

	// 2. Try extracting from JSON body if present
	if lDocID == "" && Req.Body != nil {
		var lBody struct {
			ID         string `json:"id"`
			DocumentID string `json:"document_id"`
		}
		if json.NewDecoder(Req.Body).Decode(&lBody) == nil {
			if lBody.ID != "" {
				lDocID = lBody.ID
			} else if lBody.DocumentID != "" {
				lDocID = lBody.DocumentID
			}
		}
	}

	// 3. Try extracting from URL path e.g. /api/documents/{id} or /api/documents/{id}/delete
	if lDocID == "" {
		lTrimmed := strings.Trim(Req.URL.Path, "/")
		lParts := strings.Split(lTrimmed, "/")
		if len(lParts) >= 3 && lParts[2] != "delete" {
			lDocID = lParts[2]
		}
	}

	if lDocID == "" {
		lDebug.Log(utils.Elog, "DOC40", "Missing document ID in path, query, or body")
		fmt.Fprint(Resp, utils.GetErrorString("DOC40", "Missing or invalid document ID"))
		return
	}

	lErr := repositories.DeleteDocument(lDebug, lDocID, lUserID)
	if lErr != nil {
		lDebug.Log(utils.Elog, "DOC41", lErr.Error())
		fmt.Fprint(Resp, utils.GetErrorString("DOC41", "Failed to delete document: "+lErr.Error()))
		return
	}

	lRespData := map[string]interface{}{
		"status":      "S",
		"document_id": lDocID,
		"message":     "Document deleted successfully",
	}
	json.NewEncoder(Resp).Encode(lRespData)
}

func DocumentsRouter(Resp http.ResponseWriter, Req *http.Request) {
	Resp.Header().Set("Access-Control-Allow-Origin", "*")
	Resp.Header().Set("Access-Control-Allow-Credentials", "true")
	Resp.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	Resp.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-ID")

	if Req.Method == http.MethodOptions {
		return
	}

	lPath := Req.URL.Path
	if lPath == "/api/documents" || lPath == "/api/documents/" {
		if Req.Method == http.MethodGet {
			GetDocuments(Resp, Req)
			return
		}
		if Req.Method == http.MethodDelete || Req.Method == http.MethodPost {
			DeleteDocument(Resp, Req)
			return
		}
	}
	if strings.HasSuffix(lPath, "/summary") {
		GetSummary(Resp, Req)
		return
	}
	if strings.HasSuffix(lPath, "/insights") {
		GetInsights(Resp, Req)
		return
	}
	if strings.HasSuffix(lPath, "/ask") {
		Ask(Resp, Req)
		return
	}
	if strings.HasSuffix(lPath, "/latest") {
		GetLatestDocument(Resp, Req)
		return
	}
	if strings.HasSuffix(lPath, "/delete") || Req.Method == http.MethodDelete {
		DeleteDocument(Resp, Req)
		return
	}
	http.NotFound(Resp, Req)
}
