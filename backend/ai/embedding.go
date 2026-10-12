package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"backend/config"
	"backend/utils"
)

type geminiEmbeddingPart struct {
	Text string `json:"text"`
}

type geminiEmbeddingContent struct {
	Parts []geminiEmbeddingPart `json:"parts"`
}

type geminiEmbeddingRequest struct {
	Model                string                 `json:"model"`
	Content              geminiEmbeddingContent `json:"content"`
	OutputDimensionality int                    `json:"outputDimensionality,omitempty"`
}

type geminiEmbeddingResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

func GenerateEmbedding(pDebug *utils.HelperStruct, pText string) ([]float32, error) {
	pDebug.Log(utils.Statement, "GenerateEmbedding(+)")
	defer pDebug.Log(utils.Statement, "GenerateEmbedding(-)")

	lCfg := config.AppConfig.AI
	lBaseURL := lCfg.BaseURL
	if lBaseURL == "" {
		lBaseURL = "https://generativelanguage.googleapis.com"
	}
	lEmbedModel := lCfg.EmbedModel
	if lEmbedModel == "" {
		lEmbedModel = "gemini-embedding-001"
	}
	lAPIKey := lCfg.APIKey
	lDimension := lCfg.EmbedDimension
	if lDimension <= 0 {
		lDimension = 1536
	}

	lModelFullName := lEmbedModel
	if !strings.HasPrefix(lModelFullName, "models/") {
		lModelFullName = "models/" + lEmbedModel
	}

	lReqBody := geminiEmbeddingRequest{
		Model: lModelFullName,
		Content: geminiEmbeddingContent{
			Parts: []geminiEmbeddingPart{{Text: pText}},
		},
		OutputDimensionality: lDimension,
	}

	lReqBytes, lErr := json.Marshal(lReqBody)
	if lErr != nil {
		pDebug.Log(utils.Elog, "AI_EMB_01", "Error marshalling Gemini embedding request:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}

	lTimeoutSec := lCfg.TimeoutSeconds
	if lTimeoutSec <= 0 {
		lTimeoutSec = 30
	}
	lCtx, lCancel := context.WithTimeout(context.Background(), time.Duration(lTimeoutSec)*time.Second)
	defer lCancel()

	lURL := fmt.Sprintf("%s/v1beta/models/%s:embedContent?key=%s", lBaseURL, lEmbedModel, lAPIKey)

	lHTTPReq, lErr := http.NewRequestWithContext(lCtx, http.MethodPost, lURL, bytes.NewReader(lReqBytes))
	if lErr != nil {
		pDebug.Log(utils.Elog, "AI_EMB_02", "Error building Gemini embedding HTTP request:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}
	lHTTPReq.Header.Set("Content-Type", "application/json")

	lClient := &http.Client{}
	lHTTPResp, lErr := lClient.Do(lHTTPReq)
	if lErr != nil {
		pDebug.Log(utils.Elog, "AI_EMB_03", "Error calling Gemini embedding API:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}
	defer lHTTPResp.Body.Close()

	lBodyBytes, lErr := io.ReadAll(lHTTPResp.Body)
	if lErr != nil {
		pDebug.Log(utils.Elog, "AI_EMB_04", "Error reading Gemini embedding response body:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}

	if lHTTPResp.StatusCode != http.StatusOK {
		pDebug.Log(utils.Elog, "AI_EMB_05", "Gemini embedding API returned non-200 status:", lHTTPResp.StatusCode, string(lBodyBytes))
		return nil, utils.ErrReturn(fmt.Errorf("gemini embedding API status %d: %s", lHTTPResp.StatusCode, string(lBodyBytes)))
	}

	var lResData geminiEmbeddingResponse
	if lErr := json.Unmarshal(lBodyBytes, &lResData); lErr != nil {
		pDebug.Log(utils.Elog, "AI_EMB_06", "Error decoding Gemini embedding JSON:", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}

	if len(lResData.Embedding.Values) == 0 {
		pDebug.Log(utils.Elog, "AI_EMB_07", "No embedding values returned by Gemini")
		return nil, utils.ErrReturn(fmt.Errorf("no embedding returned"))
	}

	return lResData.Embedding.Values, nil
}
