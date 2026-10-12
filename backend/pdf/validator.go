package pdf

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"

	"backend/utils"
)

type AuditCheckResult struct {
	CheckName string
	Status    string // PASS, FAIL, WARN
	Message   string
}

type PDFValidationResult struct {
	IsValid         bool
	TotalPages      int
	TextPages       int
	EmptyPages      int
	TotalCharacters int
	TotalWords      int
	Classification  string // FULLY_TEXT, MIXED, IMAGE_ONLY
	FileHash        string
	FileSize        int64
	MimeType        string
	AuditChecks     []AuditCheckResult
	ErrorMessage    string
}

// ValidateUploadedFile inspects file headers, magic bytes, and size before persisting
func ValidateUploadedFile(
	pDebug *utils.HelperStruct,
	pFile multipart.File,
	pHeader *multipart.FileHeader,
	pMaxBytes int64,
) (*PDFValidationResult, error) {
	pDebug.Log(utils.Statement, "ValidateUploadedFile(+)")
	defer pDebug.Log(utils.Statement, "ValidateUploadedFile(-)")

	lRes := &PDFValidationResult{
		IsValid:     false,
		AuditChecks: make([]AuditCheckResult, 0),
	}

	// 1. Extension validation
	lExt := strings.ToLower(filepath.Ext(pHeader.Filename))
	if lExt != ".pdf" {
		lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
			CheckName: "FILE_EXTENSION",
			Status:    "FAIL",
			Message:   "File extension is not .pdf",
		})
		lRes.ErrorMessage = "Only PDF files are allowed"
		pDebug.Log(utils.Elog, "VAL_EXT_01", lRes.ErrorMessage)
		return lRes, fmt.Errorf("VAL_EXT_01: %s", lRes.ErrorMessage)
	}
	lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
		CheckName: "FILE_EXTENSION",
		Status:    "PASS",
		Message:   "Valid .pdf extension verified",
	})

	// 2. File size validation
	lRes.FileSize = pHeader.Size
	if pHeader.Size > pMaxBytes {
		lErrMsg := fmt.Sprintf("File size (%d bytes) exceeds maximum limit (%d bytes)", pHeader.Size, pMaxBytes)
		lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
			CheckName: "FILE_SIZE",
			Status:    "FAIL",
			Message:   lErrMsg,
		})
		lRes.ErrorMessage = lErrMsg
		pDebug.Log(utils.Elog, "VAL_SIZE_01", lErrMsg)
		return lRes, fmt.Errorf("VAL_SIZE_01: %s", lErrMsg)
	}
	lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
		CheckName: "FILE_SIZE",
		Status:    "PASS",
		Message:   fmt.Sprintf("File size %d bytes within limit", pHeader.Size),
	})

	// 3. Read header for magic bytes and MIME type
	lHeaderBuf := make([]byte, 1024)
	lBytesRead, lErr := pFile.Read(lHeaderBuf)
	if lErr != nil && lErr != io.EOF {
		lRes.ErrorMessage = "Failed to read file header"
		pDebug.Log(utils.Elog, "VAL_HDR_01", lErr.Error())
		return lRes, utils.ErrReturn(lErr)
	}

	// Check PDF magic signature: %PDF-
	if !bytes.HasPrefix(lHeaderBuf[:lBytesRead], []byte("%PDF-")) {
		lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
			CheckName: "PDF_SIGNATURE",
			Status:    "FAIL",
			Message:   "Invalid PDF signature (missing %PDF- magic bytes)",
		})
		lRes.ErrorMessage = "The uploaded file is not a valid PDF document"
		pDebug.Log(utils.Elog, "VAL_SIG_01", lRes.ErrorMessage)
		return lRes, fmt.Errorf("VAL_SIG_01: %s", lRes.ErrorMessage)
	}
	lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
		CheckName: "PDF_SIGNATURE",
		Status:    "PASS",
		Message:   "Valid %PDF- signature verified",
	})

	// 4. MIME type check
	lMime := http.DetectContentType(lHeaderBuf[:lBytesRead])
	lRes.MimeType = lMime
	lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
		CheckName: "MIME_TYPE",
		Status:    "PASS",
		Message:   fmt.Sprintf("Detected MIME type: %s", lMime),
	})

	// 5. Calculate SHA-256 hash of entire file
	if _, lErr := pFile.Seek(0, io.SeekStart); lErr != nil {
		return lRes, utils.ErrReturn(lErr)
	}
	lHasher := sha256.New()
	if _, lErr := io.Copy(lHasher, pFile); lErr != nil {
		return lRes, utils.ErrReturn(lErr)
	}
	lRes.FileHash = hex.EncodeToString(lHasher.Sum(nil))

	// Reset file reader for subsequent saving
	if _, lErr := pFile.Seek(0, io.SeekStart); lErr != nil {
		return lRes, utils.ErrReturn(lErr)
	}

	lRes.IsValid = true
	return lRes, nil
}

// ValidatePDFContent inspects page count, extractable text, and scanned/mixed status
func ValidatePDFContent(
	pDebug *utils.HelperStruct,
	pFilePath string,
	pMaxPages int,
) (*PDFValidationResult, error) {
	pDebug.Log(utils.Statement, "ValidatePDFContent(+)")
	defer pDebug.Log(utils.Statement, "ValidatePDFContent(-)")

	lRes := &PDFValidationResult{
		IsValid:     false,
		AuditChecks: make([]AuditCheckResult, 0),
	}

	// 1. Open with PDF parser
	lFile, lReader, lErr := pdf.Open(pFilePath)
	if lErr != nil {
		lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
			CheckName: "PDF_PARSER",
			Status:    "FAIL",
			Message:   "PDF parsing engine failed to open file: " + lErr.Error(),
		})
		lRes.ErrorMessage = "Unable to parse PDF content; the file may be corrupt or encrypted."
		pDebug.Log(utils.Elog, "VAL_PARSE_01", lErr.Error())
		return lRes, fmt.Errorf("VAL_PARSE_01: %s", lRes.ErrorMessage)
	}
	defer lFile.Close()

	lTotalPages := lReader.NumPage()
	lRes.TotalPages = lTotalPages

	if lTotalPages == 0 {
		lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
			CheckName: "PAGE_COUNT",
			Status:    "FAIL",
			Message:   "PDF contains 0 pages",
		})
		lRes.ErrorMessage = "The PDF contains no pages."
		pDebug.Log(utils.Elog, "VAL_PAGE_01", lRes.ErrorMessage)
		return lRes, fmt.Errorf("VAL_PAGE_01: %s", lRes.ErrorMessage)
	}

	if pMaxPages > 0 && lTotalPages > pMaxPages {
		lErrMsg := fmt.Sprintf("PDF page count (%d) exceeds maximum allowed (%d)", lTotalPages, pMaxPages)
		lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
			CheckName: "PAGE_COUNT",
			Status:    "FAIL",
			Message:   lErrMsg,
		})
		lRes.ErrorMessage = lErrMsg
		pDebug.Log(utils.Elog, "VAL_PAGE_02", lErrMsg)
		return lRes, fmt.Errorf("VAL_PAGE_02: %s", lErrMsg)
	}

	lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
		CheckName: "PAGE_COUNT",
		Status:    "PASS",
		Message:   fmt.Sprintf("PDF contains %d pages (within limit of %d)", lTotalPages, pMaxPages),
	})

	// 2. Scan pages for extractable text
	lTextPages := 0
	lEmptyPages := 0
	lTotalChars := 0
	lTotalWords := 0

	for lPageNum := 1; lPageNum <= lTotalPages; lPageNum++ {
		lPage := lReader.Page(lPageNum)
		if lPage.V.IsNull() {
			lEmptyPages++
			continue
		}

		lPlainText, _ := lPage.GetPlainText(nil)
		lTrimmed := strings.TrimSpace(lPlainText)

		if len(lTrimmed) > 20 { // Consider page text-bearing if it has >20 characters
			lTextPages++
			lTotalChars += len(lTrimmed)
			lWords := strings.Fields(lTrimmed)
			lTotalWords += len(lWords)
		} else {
			lEmptyPages++
		}
	}

	lRes.TextPages = lTextPages
	lRes.EmptyPages = lEmptyPages
	lRes.TotalCharacters = lTotalChars
	lRes.TotalWords = lTotalWords

	// 3. Classify document
	if lTextPages == 0 {
		lRes.Classification = "IMAGE_ONLY"
		lErrMsg := "This system currently supports text-based PDFs only. The uploaded PDF appears to contain scanned/image-only pages."
		lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
			CheckName: "TEXT_CONTENT",
			Status:    "FAIL",
			Message:   lErrMsg,
		})
		lRes.ErrorMessage = lErrMsg
		pDebug.Log(utils.Elog, "VAL_TXT_01", lErrMsg)
		return lRes, fmt.Errorf("VAL_TXT_01: %s", lErrMsg)
	}

	if lEmptyPages > 0 {
		lRes.Classification = "MIXED"
		lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
			CheckName: "TEXT_CONTENT",
			Status:    "WARN",
			Message:   fmt.Sprintf("Mixed PDF: %d text pages will be processed; %d non-text pages were skipped", lTextPages, lEmptyPages),
		})
	} else {
		lRes.Classification = "FULLY_TEXT"
		lRes.AuditChecks = append(lRes.AuditChecks, AuditCheckResult{
			CheckName: "TEXT_CONTENT",
			Status:    "PASS",
			Message:   fmt.Sprintf("Fully text-based PDF: %d text pages verified (%d words, %d chars)", lTextPages, lTotalWords, lTotalChars),
		})
	}

	// Calculate file metadata on disk
	if lInfo, lErr := os.Stat(pFilePath); lErr == nil {
		lRes.FileSize = lInfo.Size()
	}

	lRes.IsValid = true
	return lRes, nil
}
