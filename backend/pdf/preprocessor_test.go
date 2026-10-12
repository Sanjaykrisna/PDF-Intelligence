package pdf

import (
	"strings"
	"testing"

	"backend/config"
	"backend/utils"
)

func TestStandardizeText(t *testing.T) {
	// Test 1: Hyphen repair
	lInput := "The system uses an ultra-\nsonic sensor."
	lCleaned := StandardizeText(lInput)
	if !strings.Contains(lCleaned, "ultrasonic") {
		t.Errorf("Expected 'ultrasonic', got: %s", lCleaned)
	}

	// Test 2: Invisible characters removal
	lInvisible := "Hello\u200B World\uFEFF!"
	lCleaned2 := StandardizeText(lInvisible)
	if strings.Contains(lCleaned2, "\u200B") || strings.Contains(lCleaned2, "\uFEFF") {
		t.Errorf("Invisible characters not removed: %s", lCleaned2)
	}

	// Test 3: Multiple spaces and newlines
	lMulti := "Paragraph 1\n\n\n\nParagraph 2    with   extra   spaces."
	lCleaned3 := StandardizeText(lMulti)
	if strings.Contains(lCleaned3, "\n\n\n") {
		t.Errorf("Excess newlines not collapsed: %s", lCleaned3)
	}
	if strings.Contains(lCleaned3, "   ") {
		t.Errorf("Excess spaces not collapsed: %s", lCleaned3)
	}
}

func TestRemoveHeadersAndFooters(t *testing.T) {
	lPages := []PageText{
		{
			PageNumber: 1,
			CleanText:  "PROJECT REPORT 2026\nIntroduction section content.\nPage 1",
		},
		{
			PageNumber: 2,
			CleanText:  "PROJECT REPORT 2026\nMethodology section content.\nPage 2",
		},
		{
			PageNumber: 3,
			CleanText:  "PROJECT REPORT 2026\nResults section content.\nPage 3",
		},
	}

	lProcessed := RemoveHeadersAndFooters(lPages)
	for _, lP := range lProcessed {
		if strings.Contains(lP.CleanText, "PROJECT REPORT 2026") {
			t.Errorf("Recurring header 'PROJECT REPORT 2026' was not removed from page %d", lP.PageNumber)
		}
	}
}

func TestRedactPII(t *testing.T) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()

	lPolicy := config.PrivacyConfig{
		RemoveEmail:      true,
		RemovePhone:      true,
		RemoveRollNumber: true,
		RemoveStudentID:  true,
		RemoveAadhaar:    true,
	}

	lText := `
Student Details:
Name: Test Student
Email: student.project@college.edu
Contact: +91 9876543210
Alternate: 09876543211
Roll No: 21CS1084
Student ID: STU-2026-99
Aadhaar: 1234 5678 9012
Microcontroller: Arduino Uno
`

	lRedacted, lFound, lCats := RedactPII(lDebug, lText, lPolicy)

	if !lFound {
		t.Fatalf("Expected PII to be detected")
	}

	if strings.Contains(lRedacted, "student.project@college.edu") {
		t.Errorf("Email was not redacted: %s", lRedacted)
	}
	if strings.Contains(lRedacted, "9876543210") {
		t.Errorf("Phone number was not redacted: %s", lRedacted)
	}
	if strings.Contains(lRedacted, "21CS1084") {
		t.Errorf("Roll number was not redacted: %s", lRedacted)
	}
	if strings.Contains(lRedacted, "STU-2026-99") {
		t.Errorf("Student ID was not redacted: %s", lRedacted)
	}
	if strings.Contains(lRedacted, "1234 5678 9012") {
		t.Errorf("Aadhaar was not redacted: %s", lRedacted)
	}

	// Crucial check: Meaningful technical details must NOT be redacted!
	if !strings.Contains(lRedacted, "Arduino Uno") {
		t.Errorf("Technical term 'Arduino Uno' was accidentally modified: %s", lRedacted)
	}

	t.Logf("Detected PII Categories: %v", lCats)
}
