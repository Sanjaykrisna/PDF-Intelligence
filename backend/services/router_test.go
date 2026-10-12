package services

import (
	"strings"
	"testing"

	"backend/models"
	"backend/utils"
)

func TestAnalyzeQuery(t *testing.T) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()

	// Test 1: Basic intent
	lRes1 := AnalyzeQuery(lDebug, "What is the project title?")
	if lRes1.Intent != "basic" {
		t.Errorf("Expected basic intent, got: %s", lRes1.Intent)
	}

	// Test 2: Reasoning intent
	lRes2 := AnalyzeQuery(lDebug, "Why does the system reduce motor speed when an obstacle is detected?")
	if lRes2.Intent != "reasoning" {
		t.Errorf("Expected reasoning intent, got: %s", lRes2.Intent)
	}

	// Test 3: Semantic candidate expansion (break -> brake in vehicle context)
	lRes3 := AnalyzeQuery(lDebug, "How does the break mechanism work?")
	hasBrake := false
	for _, lAlt := range lRes3.AlternativeQueries {
		if strings.Contains(lAlt, "brake") {
			hasBrake = true
			break
		}
	}
	if !hasBrake {
		t.Errorf("Expected 'break' to expand to 'brake', alternatives: %v", lRes3.AlternativeQueries)
	}
}

func TestTryDirectAnswer(t *testing.T) {
	lDebug := new(utils.HelperStruct)
	lDebug.Init()

	lDoc := &models.Document{
		Title: "Autonomous Obstacle Avoidance Rover",
		Summary: models.DocumentSummary{
			Objective:    "To build an obstacle avoidance rover using Arduino.",
			Technologies: []string{"Arduino", "C++", "Sensors"},
		},
	}

	// Direct Title
	lAnalysisTitle := AnalyzeQuery(lDebug, "What is the project title?")
	lAns, lOk := TryDirectAnswer(lDebug, lDoc, lAnalysisTitle, nil)
	if !lOk || !strings.Contains(lAns.Answer, "Autonomous Obstacle Avoidance Rover") {
		t.Errorf("Expected direct title answer, got: %v, ok: %v", lAns, lOk)
	}

	// Direct Technologies
	lAnalysisTech := AnalyzeQuery(lDebug, "What technologies are used?")
	lAnsTech, lOkTech := TryDirectAnswer(lDebug, lDoc, lAnalysisTech, nil)
	if !lOkTech || !strings.Contains(lAnsTech.Answer, "Arduino") {
		t.Errorf("Expected direct technology answer, got: %v, ok: %v", lAnsTech, lOkTech)
	}

	// Non-basic query should NOT direct answer
	lAnalysisReason := AnalyzeQuery(lDebug, "Explain why ultrasonic sensors fail in certain weather conditions.")
	_, lOkReason := TryDirectAnswer(lDebug, lDoc, lAnalysisReason, nil)
	if lOkReason {
		t.Errorf("Reasoning query should not return direct answer")
	}
}
