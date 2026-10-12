package pdf

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"

	"backend/config"
	"backend/models"
	"backend/utils"
)

type ExtractionResult struct {
	Chunks       []models.Chunk
	PIIDetected  bool
	RedactedPII  []string
	TextPages    int
	TotalPages   int
	SanitizedDoc string
}

func ExtractTextAndChunk(pDebug *utils.HelperStruct, pFilePath string, pDocumentID string) ([]models.Chunk, error) {
	pDebug.Log(utils.Statement, "ExtractTextAndChunk(+)")
	defer pDebug.Log(utils.Statement, "ExtractTextAndChunk(-)")

	lRes, lErr := ExtractAndProcessPDF(pDebug, pFilePath, pDocumentID)
	if lErr != nil {
		return nil, utils.ErrReturn(lErr)
	}

	return lRes.Chunks, nil
}

func ExtractAndProcessPDF(pDebug *utils.HelperStruct, pFilePath string, pDocumentID string) (*ExtractionResult, error) {
	pDebug.Log(utils.Statement, "ExtractAndProcessPDF(+)")
	defer pDebug.Log(utils.Statement, "ExtractAndProcessPDF(-)")

	lFile, lReader, lErr := pdf.Open(pFilePath)
	if lErr != nil {
		pDebug.Log(utils.Elog, "PDF01", lErr.Error())
		return nil, utils.ErrReturn(lErr)
	}
	defer lFile.Close()

	lTotalPages := lReader.NumPage()
	var lRawPages []PageText

	// 1. Ingest raw text per page
	for lPageNum := 1; lPageNum <= lTotalPages; lPageNum++ {
		lPage := lReader.Page(lPageNum)
		if lPage.V.IsNull() {
			continue
		}

		var lBuf bytes.Buffer
		lText, _ := lPage.GetPlainText(nil)
		lBuf.WriteString(lText)

		lRawStr := lBuf.String()
		lCleaned := StandardizeText(lRawStr)

		if strings.TrimSpace(lCleaned) != "" {
			lRawPages = append(lRawPages, PageText{
				PageNumber: lPageNum,
				RawText:    lRawStr,
				CleanText:  lCleaned,
			})
		}
	}

	if len(lRawPages) == 0 {
		return nil, fmt.Errorf("no extractable text found in PDF")
	}

	// 2. Remove recurring headers and footers across pages (Phase 4)
	lFilteredPages := RemoveHeadersAndFooters(lRawPages)

	// 3. PII Detection and Redaction (Phase 5)
	lPiiDetected := false
	lRedactedSet := make(map[string]bool)
	lPrivacyPolicy := config.AppConfig.Privacy

	var lSanitizedPages []PageText
	var lFullSanitizedText strings.Builder

	for _, lP := range lFilteredPages {
		lRedactedText, lFound, lCats := RedactPII(pDebug, lP.CleanText, lPrivacyPolicy)
		if lFound {
			lPiiDetected = true
			for _, lCat := range lCats {
				lRedactedSet[lCat] = true
			}
		}

		lSanitizedPages = append(lSanitizedPages, PageText{
			PageNumber: lP.PageNumber,
			RawText:    lP.RawText,
			CleanText:  lRedactedText,
		})

		lFullSanitizedText.WriteString(lRedactedText)
		lFullSanitizedText.WriteString("\n\n")
	}

	var lRedactedList []string
	for lCat := range lRedactedSet {
		lRedactedList = append(lRedactedList, lCat)
	}

	// 4. Hierarchical Smart Chunking (Phase 9)
	var lChunks []models.Chunk
	lChunkIndex := 0
	lCurrentSection := "General"

	lHeadingRegex := regexp.MustCompile(`^(?i)(Abstract|Introduction|Problem Statement|Existing System|Proposed System|Methodology|Architecture|System Design|Implementation|Hardware Requirements|Software Requirements|Results|Advantages|Limitations|Future Scope|Conclusion)[a-z0-9\s]*$`)

	lAIConfig := config.AppConfig.AI
	lProvider := lAIConfig.Provider
	if lProvider == "" {
		lProvider = "gemini"
	}
	lModel := lAIConfig.EmbedModel
	if lModel == "" {
		lModel = "gemini-embedding-001"
	}
	lDim := lAIConfig.EmbedDimension
	if lDim <= 0 {
		lDim = 1536
	}

	for _, lP := range lSanitizedPages {
		lParagraphs := strings.Split(lP.CleanText, "\n\n")
		var lChunkBuilder strings.Builder

		for _, lPara := range lParagraphs {
			lPara = strings.TrimSpace(lPara)
			if lPara == "" {
				continue
			}

			// Detect section heading
			if len(lPara) < 100 && lHeadingRegex.MatchString(lPara) {
				if lChunkBuilder.Len() > 0 {
					lChunkText := strings.TrimSpace(lChunkBuilder.String())
					if len(lChunkText) > 20 {
						lChunks = append(lChunks, models.Chunk{
							DocumentID:         pDocumentID,
							PageNumber:         lP.PageNumber,
							SectionName:        lCurrentSection,
							ChunkIndex:         lChunkIndex,
							Text:               lChunkText,
							ContentHash:        hashText(lChunkText),
							EmbeddingProvider:  lProvider,
							EmbeddingModel:     lModel,
							EmbeddingDimension: lDim,
						})
						lChunkIndex++
					}
					lChunkBuilder.Reset()
				}
				lCurrentSection = lPara
			}

			if lChunkBuilder.Len() > 0 {
				lChunkBuilder.WriteString("\n")
			}
			lChunkBuilder.WriteString(lPara)

			// Size limit check: ~800 characters with ~150 overlap
			if lChunkBuilder.Len() > 800 {
				lChunkText := strings.TrimSpace(lChunkBuilder.String())
				lChunks = append(lChunks, models.Chunk{
					DocumentID:         pDocumentID,
					PageNumber:         lP.PageNumber,
					SectionName:        lCurrentSection,
					ChunkIndex:         lChunkIndex,
					Text:               lChunkText,
					ContentHash:        hashText(lChunkText),
					EmbeddingProvider:  lProvider,
					EmbeddingModel:     lModel,
					EmbeddingDimension: lDim,
				})
				lChunkIndex++

				// Retain last 150 chars for semantic overlap
				lChunkBuilder.Reset()
				if len(lChunkText) > 150 {
					lChunkBuilder.WriteString(lChunkText[len(lChunkText)-150:])
				}
			}
		}

		// Flush trailing chunk per page
		if lChunkBuilder.Len() > 50 {
			lChunkText := strings.TrimSpace(lChunkBuilder.String())
			lChunks = append(lChunks, models.Chunk{
				DocumentID:         pDocumentID,
				PageNumber:         lP.PageNumber,
				SectionName:        lCurrentSection,
				ChunkIndex:         lChunkIndex,
				Text:               lChunkText,
				ContentHash:        hashText(lChunkText),
				EmbeddingProvider:  lProvider,
				EmbeddingModel:     lModel,
				EmbeddingDimension: lDim,
			})
			lChunkIndex++
		}
	}

	return &ExtractionResult{
		Chunks:       lChunks,
		PIIDetected:  lPiiDetected,
		RedactedPII:  lRedactedList,
		TextPages:    len(lSanitizedPages),
		TotalPages:   lTotalPages,
		SanitizedDoc: lFullSanitizedText.String(),
	}, nil
}

func hashText(pText string) string {
	lHash := sha256.Sum256([]byte(pText))
	return hex.EncodeToString(lHash[:])
}
