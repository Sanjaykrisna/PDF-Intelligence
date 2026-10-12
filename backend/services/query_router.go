package services

import (
	"fmt"
	"regexp"
	"strings"

	"backend/models"
	"backend/utils"
)

func AnalyzeQuery(pDebug *utils.HelperStruct, pRawQuery string) *models.QueryAnalysis {
	pDebug.Log(utils.Statement, "AnalyzeQuery(+)")
	defer pDebug.Log(utils.Statement, "AnalyzeQuery(-)")

	lTrimmed := strings.TrimSpace(pRawQuery)
	lLower := strings.ToLower(lTrimmed)

	// Clean punctuation from search tokens
	lCleanRegex := regexp.MustCompile(`[^\w\s\-]`)
	lNormalized := lCleanRegex.ReplaceAllString(lLower, " ")
	lSpaceRegex := regexp.MustCompile(`\s+`)
	lNormalized = strings.TrimSpace(lSpaceRegex.ReplaceAllString(lNormalized, " "))

	// Check intent: basic lookup vs reasoning
	lIntent := "reasoning"
	lConfidence := 0.70

	lBasicPatterns := []string{
		`\b(?:what\s+is\s+the\s+title|project\s+title)\b`,
		`\b(?:who\s+is\s+the\s+author|student\s+name|submitted\s+by)\b`,
		`\b(?:what\s+(?:is\s+the\s+)?sensor\w*|which\s+sensor\w*)\b`,
		`\b(?:what\s+(?:is\s+the\s+)?(?:micro)?controller\w*|which\s+(?:micro)?controller\w*)\b`,
		`\b(?:what\s+(?:are\s+)?(?:the\s+)?technolog\w*|which\s+technolog\w*|tech\s*stack|tools\s+used)\b`,
		`\b(?:what\s+algorithm\w*|which\s+algorithm\w*)\b`,
		`\b(?:what\s+is\s+the\s+objective\w*|aim\s+of\s+the\s+project)\b`,
		`\b(?:what\s+is\s+the\s+problem\w*|problem\s+statement)\b`,
		`\b(?:what\s+are\s+the\s+advantage\w*|benefit\w*)\b`,
		`\b(?:what\s+are\s+the\s+limitation\w*|disadvantage\w*)\b`,
		`\b(?:future\s+scope|future\s+enhancement\w*)\b`,
	}

	for _, lPattern := range lBasicPatterns {
		if matched, _ := regexp.MatchString(lPattern, lLower); matched {
			lIntent = "basic"
			lConfidence = 0.90
			break
		}
	}

	// Generate candidate semantic alternatives (Phase 15: semantic expansion)
	var lAlternatives []string

	// Check vehicle/mechanical typos like break -> brake
	if strings.Contains(lNormalized, "break") && (strings.Contains(lNormalized, "mechanism") || strings.Contains(lNormalized, "speed") || strings.Contains(lNormalized, "motor") || strings.Contains(lNormalized, "car")) {
		lAlt := strings.ReplaceAll(lNormalized, "break", "brake")
		lAlternatives = append(lAlternatives, lAlt)
	}

	// Synonym expansions
	if strings.Contains(lNormalized, "hardware") {
		lAlternatives = append(lAlternatives, strings.ReplaceAll(lNormalized, "hardware", "components"))
	}
	if strings.Contains(lNormalized, "algorithm") {
		lAlternatives = append(lAlternatives, strings.ReplaceAll(lNormalized, "algorithm", "methodology"))
	}

	return &models.QueryAnalysis{
		OriginalQuery:      lTrimmed,
		NormalizedQuery:    lNormalized,
		AlternativeQueries: lAlternatives,
		Intent:             lIntent,
		Confidence:         lConfidence,
	}
}

// TryDirectAnswer handles Phase 14: Answer factual basic lookups without calling LLM when confidence is high
func TryDirectAnswer(
	pDebug *utils.HelperStruct,
	pDoc *models.Document,
	pAnalysis *models.QueryAnalysis,
	pChunks []models.ChunkSearchResult,
) (*models.AnswerResponse, bool) {
	pDebug.Log(utils.Statement, "TryDirectAnswer(+)")
	defer pDebug.Log(utils.Statement, "TryDirectAnswer(-)")

	if pAnalysis.Intent != "basic" {
		return nil, false
	}

	lQ := pAnalysis.NormalizedQuery
	var lSummary *models.DocumentSummary
	if pDoc.Summary != nil {
		if lSum, ok := pDoc.Summary.(models.DocumentSummary); ok {
			lSummary = &lSum
		}
	}

	// 1. Title lookup
	if strings.Contains(lQ, "title") {
		lTitle := pDoc.Title
		if lTitle != "" {
			return &models.AnswerResponse{
				Answer:     fmt.Sprintf("The title of this project is: %s", lTitle),
				Level:      1,
				QueryType:  "direct",
				Confidence: 1.0,
				Sources:    []models.SourceReference{{Page: 1, Section: "Title Page"}},
			}, true
		}
	}

	// 2. Structured summary lookups
	if lSummary != nil {
		if (strings.Contains(lQ, "objective") || strings.Contains(lQ, "aim")) && lSummary.Objective != "" {
			return &models.AnswerResponse{
				Answer:     lSummary.Objective,
				Level:      1,
				QueryType:  "direct",
				Confidence: 0.95,
				Sources:    []models.SourceReference{{Page: 1, Section: "Abstract / Objective"}},
			}, true
		}
		if (strings.Contains(lQ, "problem") || strings.Contains(lQ, "problem statement")) && lSummary.Problem != "" {
			return &models.AnswerResponse{
				Answer:     lSummary.Problem,
				Level:      1,
				QueryType:  "direct",
				Confidence: 0.95,
				Sources:    []models.SourceReference{{Page: 1, Section: "Problem Statement"}},
			}, true
		}
		if strings.Contains(lQ, "technolog") && len(lSummary.Technologies) > 0 {
			return &models.AnswerResponse{
				Answer:     fmt.Sprintf("The technologies used in this project are: %s", strings.Join(lSummary.Technologies, ", ")),
				Level:      1,
				QueryType:  "direct",
				Confidence: 0.95,
				Sources:    []models.SourceReference{{Page: 1, Section: "Technologies"}},
			}, true
		}
		if strings.Contains(lQ, "algorithm") && len(lSummary.Algorithms) > 0 {
			return &models.AnswerResponse{
				Answer:     fmt.Sprintf("The algorithms implemented are: %s", strings.Join(lSummary.Algorithms, ", ")),
				Level:      1,
				QueryType:  "direct",
				Confidence: 0.95,
				Sources:    []models.SourceReference{{Page: 1, Section: "Algorithms"}},
			}, true
		}
		if strings.Contains(lQ, "advantage") && len(lSummary.Advantages) > 0 {
			return &models.AnswerResponse{
				Answer:     fmt.Sprintf("Key advantages: %s", strings.Join(lSummary.Advantages, "; ")),
				Level:      1,
				QueryType:  "direct",
				Confidence: 0.95,
				Sources:    []models.SourceReference{{Page: 1, Section: "Advantages"}},
			}, true
		}
		if strings.Contains(lQ, "limitation") && len(lSummary.Limitations) > 0 {
			return &models.AnswerResponse{
				Answer:     fmt.Sprintf("Key limitations: %s", strings.Join(lSummary.Limitations, "; ")),
				Level:      1,
				QueryType:  "direct",
				Confidence: 0.95,
				Sources:    []models.SourceReference{{Page: 1, Section: "Limitations"}},
			}, true
		}
		if strings.Contains(lQ, "future scope") && lSummary.FutureScope != "" {
			return &models.AnswerResponse{
				Answer:     lSummary.FutureScope,
				Level:      1,
				QueryType:  "direct",
				Confidence: 0.95,
				Sources:    []models.SourceReference{{Page: 1, Section: "Future Scope"}},
			}, true
		}
	}

	// 3. Exact factual extraction from retrieved chunks
	if len(pChunks) > 0 && pChunks[0].CombinedScore > 0.80 {
		lTopText := pChunks[0].Text

		// Check sensor pattern: e.g. "sensor: Ultrasonic HC-SR04" or "used sensor is ..."
		if strings.Contains(lQ, "sensor") {
			lSensorRegex := regexp.MustCompile(`(?i)(?:sensor(?:\s+used)?\s*[:=-]\s*|\bused\s+)([\w\s-]{3,30}(?:sensor|ultrasonic|infrared|pir|camera|gyroscope|lidar))`)
			if lMatch := lSensorRegex.FindStringSubmatch(lTopText); len(lMatch) > 1 {
				return &models.AnswerResponse{
					Answer:     fmt.Sprintf("The sensor specified in the project is: %s", strings.TrimSpace(lMatch[1])),
					Level:      1,
					QueryType:  "direct",
					Confidence: 0.88,
					Sources:    []models.SourceReference{{Page: pChunks[0].PageNumber, Section: pChunks[0].SectionName}},
				}, true
			}
		}

		// Check controller pattern: e.g. "microcontroller: Arduino Uno" or "ESP32" or "Raspberry Pi"
		if strings.Contains(lQ, "controller") || strings.Contains(lQ, "microcontroller") {
			lCtrlRegex := regexp.MustCompile(`(?i)(?:(?:micro)?controller\s*[:=-]\s*|\busing\s+)(Arduino\s*(?:Uno|Nano|Mega)?|ESP32|ESP8266|Raspberry\s*Pi|STM32|PIC\w*|8051)`)
			if lMatch := lCtrlRegex.FindStringSubmatch(lTopText); len(lMatch) > 1 {
				return &models.AnswerResponse{
					Answer:     fmt.Sprintf("The controller used is: %s", strings.TrimSpace(lMatch[1])),
					Level:      1,
					QueryType:  "direct",
					Confidence: 0.90,
					Sources:    []models.SourceReference{{Page: pChunks[0].PageNumber, Section: pChunks[0].SectionName}},
				}, true
			}
		}
	}

	return nil, false
}
