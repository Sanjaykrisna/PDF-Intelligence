package pdf

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"testing"

	"backend/utils"
)

type dummyFile struct {
	*bytes.Reader
}

func (d *dummyFile) Close() error {
	return nil
}

func TestValidateUploadedFile_ValidPDF(t *testing.T) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()

	lPdfContent := []byte("%PDF-1.5\n%Valid PDF content for testing\n")
	lHeader := &multipart.FileHeader{
		Filename: "report.pdf",
		Size:     int64(len(lPdfContent)),
		Header:   make(textproto.MIMEHeader),
	}
	lFile := &dummyFile{bytes.NewReader(lPdfContent)}

	lRes, lErr := ValidateUploadedFile(lDebug, lFile, lHeader, 10*1024*1024)
	if lErr != nil {
		t.Fatalf("Expected valid PDF, got error: %v", lErr)
	}
	if !lRes.IsValid {
		t.Fatalf("Expected IsValid to be true")
	}
	if lRes.FileHash == "" {
		t.Fatalf("Expected SHA-256 hash to be calculated")
	}
}

func TestValidateUploadedFile_InvalidExtension(t *testing.T) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()

	lContent := []byte("plain text")
	lHeader := &multipart.FileHeader{
		Filename: "report.txt",
		Size:     int64(len(lContent)),
	}
	lFile := &dummyFile{bytes.NewReader(lContent)}

	_, lErr := ValidateUploadedFile(lDebug, lFile, lHeader, 10*1024*1024)
	if lErr == nil {
		t.Fatalf("Expected error for non-pdf extension, got nil")
	}
}

func TestValidateUploadedFile_FakePDF(t *testing.T) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()

	// Spoofed .pdf filename but lacks %PDF- magic bytes
	lContent := []byte("<html><body>Fake file</body></html>")
	lHeader := &multipart.FileHeader{
		Filename: "malicious.pdf",
		Size:     int64(len(lContent)),
	}
	lFile := &dummyFile{bytes.NewReader(lContent)}

	_, lErr := ValidateUploadedFile(lDebug, lFile, lHeader, 10*1024*1024)
	if lErr == nil {
		t.Fatalf("Expected error for fake PDF without magic bytes, got nil")
	}
}

func TestValidateUploadedFile_Oversized(t *testing.T) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()

	lPdfContent := []byte("%PDF-1.5\nlarge file content\n")
	lHeader := &multipart.FileHeader{
		Filename: "huge.pdf",
		Size:     50 * 1024 * 1024, // 50MB
	}
	lFile := &dummyFile{bytes.NewReader(lPdfContent)}

	_, lErr := ValidateUploadedFile(lDebug, lFile, lHeader, 10*1024*1024) // 10MB limit
	if lErr == nil {
		t.Fatalf("Expected error for oversized file, got nil")
	}
}
