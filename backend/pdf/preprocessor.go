package pdf

import (
	"regexp"
	"strings"
	"unicode"

	"backend/config"
	"backend/utils"
)

type PageText struct {
	PageNumber int
	RawText    string
	CleanText  string
}

// StandardizeText handles Phase 3: Unicode cleanup, hyphen repair, broken lines, and whitespace
func StandardizeText(pText string) string {
	if pText == "" {
		return ""
	}

	// 1. Remove invisible / zero-width characters and BOM
	lText := strings.ReplaceAll(pText, "\uFEFF", "") // UTF-8 BOM
	lText = strings.ReplaceAll(lText, "\u200B", "") // Zero-width space
	lText = strings.ReplaceAll(lText, "\u200C", "") // Zero-width non-joiner
	lText = strings.ReplaceAll(lText, "\u200D", "") // Zero-width joiner
	lText = strings.ReplaceAll(lText, "\u00AD", "") // Soft hyphen
	lText = strings.ReplaceAll(lText, "\r\n", "\n")
	lText = strings.ReplaceAll(lText, "\r", "\n")

	// Filter unprintable control characters (except \n and \t)
	var lBuf strings.Builder
	for _, lR := range lText {
		if lR == '\n' || lR == '\t' || !unicode.IsControl(lR) {
			lBuf.WriteRune(lR)
		}
	}
	lText = lBuf.String()

	// 2. Repair hyphenated line breaks (e.g. "ultra-\nsonic" or "ultra- \nsonic" -> "ultrasonic")
	lHyphenRegex := regexp.MustCompile(`([a-zA-Z]{2,})-\s*\n\s*([a-zA-Z]{2,})`)
	lText = lHyphenRegex.ReplaceAllString(lText, "$1$2")

	// 3. Normalize single newlines inside sentences to spaces, but preserve paragraph breaks (\n\n)
	lLines := strings.Split(lText, "\n")
	var lProcessedLines []string
	for _, lLine := range lLines {
		lTrimmed := strings.TrimSpace(lLine)
		lProcessedLines = append(lProcessedLines, lTrimmed)
	}

	// Collapse 3+ consecutive newlines to double newlines
	lRecombined := strings.Join(lProcessedLines, "\n")
	lMultiNewline := regexp.MustCompile(`\n{3,}`)
	lRecombined = lMultiNewline.ReplaceAllString(lRecombined, "\n\n")

	// Collapse repeated horizontal whitespace
	lMultiSpace := regexp.MustCompile(`[ \t]+`)
	lRecombined = lMultiSpace.ReplaceAllString(lRecombined, " ")

	return strings.TrimSpace(lRecombined)
}

// RemoveHeadersAndFooters handles Phase 4: Detects repeated top/bottom lines across pages
func RemoveHeadersAndFooters(pPages []PageText) []PageText {
	if len(pPages) < 2 {
		return pPages
	}

	// Count occurrences of candidate headers (top 2 non-empty lines) and footers (bottom 2 non-empty lines)
	lHeaderCounts := make(map[string]int)
	lFooterCounts := make(map[string]int)

	for _, lP := range pPages {
		lLines := extractNonEmptyLines(lP.CleanText)
		if len(lLines) > 0 {
			lHeaderCounts[lLines[0]]++
			if len(lLines) > 1 {
				lHeaderCounts[lLines[1]]++
			}
			lFooterCounts[lLines[len(lLines)-1]]++
			if len(lLines) > 1 {
				lFooterCounts[lLines[len(lLines)-2]]++
			}
		}
	}

	// Threshold: lines appearing on at least 50% of pages (min 2) are treated as recurring headers/footers
	lThreshold := len(pPages) / 2
	if lThreshold < 2 {
		lThreshold = 2
	}

	lRecurring := make(map[string]bool)
	for lLine, lCount := range lHeaderCounts {
		if lCount >= lThreshold && len(lLine) < 120 {
			lRecurring[lLine] = true
		}
	}
	for lLine, lCount := range lFooterCounts {
		if lCount >= lThreshold && len(lLine) < 120 {
			lRecurring[lLine] = true
		}
	}

	// Filter recurring headers and footers from each page
	lResult := make([]PageText, len(pPages))
	for lIdx, lP := range pPages {
		lLines := strings.Split(lP.CleanText, "\n")
		var lFiltered []string
		for _, lLine := range lLines {
			lTrim := strings.TrimSpace(lLine)
			if lTrim != "" && lRecurring[lTrim] {
				continue
			}
			lFiltered = append(lFiltered, lLine)
		}
		lResult[lIdx] = PageText{
			PageNumber: lP.PageNumber,
			RawText:    lP.RawText,
			CleanText:  strings.TrimSpace(strings.Join(lFiltered, "\n")),
		}
	}

	return lResult
}

func extractNonEmptyLines(pText string) []string {
	lRaw := strings.Split(pText, "\n")
	var lOut []string
	for _, lLine := range lRaw {
		lTrim := strings.TrimSpace(lLine)
		if lTrim != "" {
			lOut = append(lOut, lTrim)
		}
	}
	return lOut
}

// RedactPII handles Phase 5: Configurable deterministic PII filtering and redaction
func RedactPII(pDebug *utils.HelperStruct, pText string, pPolicy config.PrivacyConfig) (string, bool, []string) {
	if pText == "" {
		return "", false, nil
	}

	lText := pText
	lDetected := false
	var lCategories []string

	// 1. Email Redaction
	if pPolicy.RemoveEmail {
		lEmailRegex := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`)
		if lEmailRegex.MatchString(lText) {
			lDetected = true
			lCategories = append(lCategories, "EMAIL")
			lText = lEmailRegex.ReplaceAllString(lText, "[EMAIL REDACTED]")
		}
	}

	// 2. Phone / Mobile Number Redaction (International & Indian formats)
	if pPolicy.RemovePhone {
		// Matches +91 9876543210, 09876543210, 9876543210, (123) 456-7890, 123-456-7890
		lPhoneRegex := regexp.MustCompile(`\b(?:\+?91[-.\s]?)?[6-9]\d{9}\b|\b(?:\+?1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b`)
		if lPhoneRegex.MatchString(lText) {
			lDetected = true
			lCategories = append(lCategories, "PHONE")
			lText = lPhoneRegex.ReplaceAllString(lText, "[PHONE REDACTED]")
		}
	}

	// 3. Roll / Registration Number Redaction
	if pPolicy.RemoveRollNumber {
		lRollRegex := regexp.MustCompile(`(?i)\b(?:roll\s*(?:no|number)?|reg(?:istration)?\s*(?:no|number)?)\s*[:.-]?\s*([A-Za-z0-9]{5,15})\b`)
		if lRollRegex.MatchString(lText) {
			lDetected = true
			lCategories = append(lCategories, "ROLL_NUMBER")
			lText = lRollRegex.ReplaceAllString(lText, "[ROLL NUMBER REDACTED]")
		}
	}

	// 4. Student ID Redaction
	if pPolicy.RemoveStudentID {
		lStudentRegex := regexp.MustCompile(`(?i)\b(?:student\s*id|matric(?:ulation)?\s*no)\s*[:.-]?\s*([A-Za-z0-9-]{4,15})\b`)
		if lStudentRegex.MatchString(lText) {
			lDetected = true
			lCategories = append(lCategories, "STUDENT_ID")
			lText = lStudentRegex.ReplaceAllString(lText, "[STUDENT ID REDACTED]")
		}
	}

	// 5. Aadhaar Redaction (XXXX XXXX XXXX)
	if pPolicy.RemoveAadhaar {
		lAadhaarRegex := regexp.MustCompile(`\b\d{4}\s\d{4}\s\d{4}\b`)
		if lAadhaarRegex.MatchString(lText) {
			lDetected = true
			lCategories = append(lCategories, "AADHAAR")
			lText = lAadhaarRegex.ReplaceAllString(lText, "[AADHAAR REDACTED]")
		}
	}

	return lText, lDetected, lCategories
}
