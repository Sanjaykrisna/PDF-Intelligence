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

type LLMProvider interface {
	Name() string
	GenerateCompletion(pDebug *utils.HelperStruct, pSystemPrompt, pUserPrompt string, pTemperature float32) (string, error)
}

type EmbeddingProvider interface {
	Name() string
	Model() string
	Dimension() int
	GenerateEmbedding(pDebug *utils.HelperStruct, pText string) ([]float32, error)
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	Temperature float64 `json:"temperature,omitempty"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent         `json:"system_instruction,omitempty"`
	Contents          []geminiContent        `json:"contents"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig"`
}

type geminiCandidateContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role"`
}

type geminiCandidate struct {
	Content      geminiCandidateContent `json:"content"`
	FinishReason string                 `json:"finishReason"`
}

type geminiAPIResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

// ---- Gemini Provider Implementation ----

type GeminiLLMProvider struct {
	BaseURL        string
	APIKey         string
	ModelName      string
	Temperature    float64
	TimeoutSeconds int
}

func NewGeminiLLMProvider(pCfg config.AIConfig) *GeminiLLMProvider {
	lBase := pCfg.BaseURL
	if lBase == "" {
		lBase = "https://generativelanguage.googleapis.com"
	}
	lMod := pCfg.Model
	if lMod == "" {
		lMod = "gemini-3.5-flash-lite"
	}
	lTimeout := pCfg.TimeoutSeconds
	if lTimeout <= 0 {
		lTimeout = 30
	}
	return &GeminiLLMProvider{
		BaseURL:        lBase,
		APIKey:         pCfg.APIKey,
		ModelName:      lMod,
		Temperature:    pCfg.Temperature,
		TimeoutSeconds: lTimeout,
	}
}

func (g *GeminiLLMProvider) Name() string {
	return "gemini"
}

func (g *GeminiLLMProvider) GenerateCompletion(pDebug *utils.HelperStruct, pSystemPrompt, pUserPrompt string, pTemperature float32) (string, error) {
	pDebug.Log(utils.Statement, "GeminiLLMProvider.GenerateCompletion(+)")
	defer pDebug.Log(utils.Statement, "GeminiLLMProvider.GenerateCompletion(-)")

	lTemp := float64(pTemperature)
	if lTemp <= 0 {
		if g.Temperature > 0 {
			lTemp = g.Temperature
		} else {
			lTemp = 0.7
		}
	}

	lReqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: pUserPrompt}},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature: lTemp,
		},
	}

	if pSystemPrompt != "" {
		lReqBody.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: pSystemPrompt}},
		}
	}

	lReqBytes, lErr := json.Marshal(lReqBody)
	if lErr != nil {
		pDebug.Log(utils.Elog, "AI_GEM_01", lErr.Error())
		return "", utils.ErrReturn(lErr)
	}

	lCtx, lCancel := context.WithTimeout(context.Background(), time.Duration(g.TimeoutSeconds)*time.Second)
	defer lCancel()

	lURL := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", g.BaseURL, g.ModelName, g.APIKey)
	lHTTPReq, lErr := http.NewRequestWithContext(lCtx, http.MethodPost, lURL, bytes.NewReader(lReqBytes))
	if lErr != nil {
		pDebug.Log(utils.Elog, "AI_GEM_02", lErr.Error())
		return "", utils.ErrReturn(lErr)
	}
	lHTTPReq.Header.Set("Content-Type", "application/json")

	lClient := &http.Client{}
	lHTTPResp, lErr := lClient.Do(lHTTPReq)
	if lErr != nil {
		pDebug.Log(utils.Elog, "AI_GEM_03", lErr.Error())
		return "", utils.ErrReturn(lErr)
	}
	defer lHTTPResp.Body.Close()

	lBodyBytes, lErr := io.ReadAll(lHTTPResp.Body)
	if lErr != nil {
		pDebug.Log(utils.Elog, "AI_GEM_04", lErr.Error())
		return "", utils.ErrReturn(lErr)
	}

	if lHTTPResp.StatusCode != http.StatusOK {
		pDebug.Log(utils.Elog, "AI_GEM_05", fmt.Sprintf("status %d: %s", lHTTPResp.StatusCode, string(lBodyBytes)))
		return "", fmt.Errorf("gemini api status %d: %s", lHTTPResp.StatusCode, string(lBodyBytes))
	}

	var lGeminiResp geminiAPIResponse
	if lErr := json.Unmarshal(lBodyBytes, &lGeminiResp); lErr != nil {
		pDebug.Log(utils.Elog, "AI_GEM_06", lErr.Error())
		return "", utils.ErrReturn(lErr)
	}

	if len(lGeminiResp.Candidates) == 0 || len(lGeminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from gemini")
	}

	return lGeminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// ---- OpenAI Provider Implementation ----

type OpenAILLMProvider struct {
	BaseURL        string
	APIKey         string
	ModelName      string
	Temperature    float64
	TimeoutSeconds int
}

func NewOpenAILLMProvider(pBase, pKey, pModel string, pTemp float64, pTimeout int) *OpenAILLMProvider {
	if pBase == "" {
		pBase = "https://api.openai.com/v1"
	}
	if pModel == "" {
		pModel = "gpt-4o-mini"
	}
	if pTimeout <= 0 {
		pTimeout = 30
	}
	return &OpenAILLMProvider{
		BaseURL:        pBase,
		APIKey:         pKey,
		ModelName:      pModel,
		Temperature:    pTemp,
		TimeoutSeconds: pTimeout,
	}
}

func (o *OpenAILLMProvider) Name() string {
	return "openai"
}

func (o *OpenAILLMProvider) GenerateCompletion(pDebug *utils.HelperStruct, pSystemPrompt, pUserPrompt string, pTemperature float32) (string, error) {
	pDebug.Log(utils.Statement, "OpenAILLMProvider.GenerateCompletion(+)")
	defer pDebug.Log(utils.Statement, "OpenAILLMProvider.GenerateCompletion(-)")

	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type req struct {
		Model       string  `json:"model"`
		Messages    []msg   `json:"messages"`
		Temperature float32 `json:"temperature"`
	}

	lReqBody := req{
		Model: o.ModelName,
		Messages: []msg{
			{Role: "system", Content: pSystemPrompt},
			{Role: "user", Content: pUserPrompt},
		},
		Temperature: pTemperature,
	}

	lBytes, _ := json.Marshal(lReqBody)
	lURL := strings.TrimRight(o.BaseURL, "/") + "/chat/completions"

	lCtx, lCancel := context.WithTimeout(context.Background(), time.Duration(o.TimeoutSeconds)*time.Second)
	defer lCancel()

	lHTTPReq, lErr := http.NewRequestWithContext(lCtx, http.MethodPost, lURL, bytes.NewReader(lBytes))
	if lErr != nil {
		return "", utils.ErrReturn(lErr)
	}
	lHTTPReq.Header.Set("Authorization", "Bearer "+o.APIKey)
	lHTTPReq.Header.Set("Content-Type", "application/json")

	lClient := &http.Client{}
	lResp, lErr := lClient.Do(lHTTPReq)
	if lErr != nil {
		return "", utils.ErrReturn(lErr)
	}
	defer lResp.Body.Close()

	lBody, _ := io.ReadAll(lResp.Body)
	if lResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai error status %d: %s", lResp.StatusCode, string(lBody))
	}

	var lParsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if lErr := json.Unmarshal(lBody, &lParsed); lErr != nil {
		return "", utils.ErrReturn(lErr)
	}
	if len(lParsed.Choices) == 0 {
		return "", fmt.Errorf("no choice returned by openai")
	}

	return lParsed.Choices[0].Message.Content, nil
}

// ---- Ollama Provider Implementation ----

type OllamaLLMProvider struct {
	BaseURL        string
	ModelName      string
	Temperature    float64
	TimeoutSeconds int
}

func NewOllamaLLMProvider(pBase, pModel string, pTemp float64, pTimeout int) *OllamaLLMProvider {
	if pBase == "" {
		pBase = "http://localhost:11434"
	}
	if pModel == "" {
		pModel = "llama3.2"
	}
	if pTimeout <= 0 {
		pTimeout = 45
	}
	return &OllamaLLMProvider{
		BaseURL:        pBase,
		ModelName:      pModel,
		Temperature:    pTemp,
		TimeoutSeconds: pTimeout,
	}
}

func (ol *OllamaLLMProvider) Name() string {
	return "ollama"
}

func (ol *OllamaLLMProvider) GenerateCompletion(pDebug *utils.HelperStruct, pSystemPrompt, pUserPrompt string, pTemperature float32) (string, error) {
	pDebug.Log(utils.Statement, "OllamaLLMProvider.GenerateCompletion(+)")
	defer pDebug.Log(utils.Statement, "OllamaLLMProvider.GenerateCompletion(-)")

	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type req struct {
		Model    string `json:"model"`
		Messages []msg  `json:"messages"`
		Stream   bool   `json:"stream"`
	}

	lReqBody := req{
		Model: ol.ModelName,
		Messages: []msg{
			{Role: "system", Content: pSystemPrompt},
			{Role: "user", Content: pUserPrompt},
		},
		Stream: false,
	}

	lBytes, _ := json.Marshal(lReqBody)
	lURL := strings.TrimRight(ol.BaseURL, "/") + "/api/chat"

	lCtx, lCancel := context.WithTimeout(context.Background(), time.Duration(ol.TimeoutSeconds)*time.Second)
	defer lCancel()

	lHTTPReq, lErr := http.NewRequestWithContext(lCtx, http.MethodPost, lURL, bytes.NewReader(lBytes))
	if lErr != nil {
		return "", utils.ErrReturn(lErr)
	}
	lHTTPReq.Header.Set("Content-Type", "application/json")

	lClient := &http.Client{}
	lResp, lErr := lClient.Do(lHTTPReq)
	if lErr != nil {
		return "", utils.ErrReturn(lErr)
	}
	defer lResp.Body.Close()

	lBody, _ := io.ReadAll(lResp.Body)
	if lResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama status %d: %s", lResp.StatusCode, string(lBody))
	}

	var lParsed struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if lErr := json.Unmarshal(lBody, &lParsed); lErr != nil {
		return "", utils.ErrReturn(lErr)
	}

	return lParsed.Message.Content, nil
}

// ---- Multi-Provider Fallback Orchestrator (Phase 21 & 22) ----

func GetPrimaryLLMProvider() LLMProvider {
	lCfg := config.AppConfig.AI
	switch strings.ToLower(lCfg.Provider) {
	case "openai":
		return NewOpenAILLMProvider(lCfg.BaseURL, lCfg.APIKey, lCfg.Model, lCfg.Temperature, lCfg.TimeoutSeconds)
	case "ollama":
		return NewOllamaLLMProvider(lCfg.BaseURL, lCfg.Model, lCfg.Temperature, lCfg.TimeoutSeconds)
	default:
		return NewGeminiLLMProvider(lCfg)
	}
}

func GetFallbackLLMProvider() LLMProvider {
	lCfg := config.AppConfig.Fallback
	if !lCfg.Enabled {
		return nil
	}
	switch strings.ToLower(lCfg.Provider) {
	case "openai":
		return NewOpenAILLMProvider(lCfg.BaseURL, lCfg.APIKey, lCfg.Model, lCfg.Temperature, lCfg.TimeoutSeconds)
	case "ollama":
		return NewOllamaLLMProvider(lCfg.BaseURL, lCfg.Model, lCfg.Temperature, lCfg.TimeoutSeconds)
	case "gemini":
		return NewGeminiLLMProvider(config.AIConfig{
			BaseURL:        lCfg.BaseURL,
			APIKey:         lCfg.APIKey,
			Model:          lCfg.Model,
			Temperature:    lCfg.Temperature,
			TimeoutSeconds: lCfg.TimeoutSeconds,
		})
	default:
		return nil
	}
}

func GenerateCompletionWithFallback(
	pDebug *utils.HelperStruct,
	pSystemPrompt string,
	pUserPrompt string,
	pTemperature float32,
) (string, string, bool, error) {
	pDebug.Log(utils.Statement, "GenerateCompletionWithFallback(+)")
	defer pDebug.Log(utils.Statement, "GenerateCompletionWithFallback(-)")

	lPrimary := GetPrimaryLLMProvider()
	lAns, lErr := lPrimary.GenerateCompletion(pDebug, pSystemPrompt, pUserPrompt, pTemperature)
	if lErr == nil {
		return lAns, lPrimary.Name(), false, nil
	}

	pDebug.Log(utils.Elog, "AI_FALLBACK_TRIGGER", "Primary LLM provider failed:", lPrimary.Name(), lErr.Error())

	// Check if fallback is configured
	lFallback := GetFallbackLLMProvider()
	if lFallback != nil {
		pDebug.Log(utils.Statement, "Attempting Fallback LLM Provider:", lFallback.Name())
		lFbAns, lFbErr := lFallback.GenerateCompletion(pDebug, pSystemPrompt, pUserPrompt, pTemperature)
		if lFbErr == nil {
			return lFbAns, lFallback.Name(), true, nil
		}
		pDebug.Log(utils.Elog, "AI_FALLBACK_FAIL", "Fallback LLM provider also failed:", lFallback.Name(), lFbErr.Error())
	}

	return "", lPrimary.Name(), false, lErr
}
