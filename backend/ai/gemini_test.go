package ai

import (
	"testing"

	"backend/config"
	"backend/utils"
)

func TestGeminiCompletionAndEmbedding(t *testing.T) {
	err := config.LoadConfig("../config.toml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	lDebug := new(utils.HelperStruct)
	lDebug.Init()

	// Test 1: Completion
	resp, err := GenerateCompletion(lDebug, "You are a helpful assistant.", "Say 'Testing 123'", 0.5)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}
	if resp == "" {
		t.Fatalf("GenerateCompletion returned empty response")
	}
	t.Logf("Completion Response: %s", resp)

	// Test 2: Embedding
	emb, err := GenerateEmbedding(lDebug, "This is a test sentence for vector embedding.")
	if err != nil {
		t.Fatalf("GenerateEmbedding failed: %v", err)
	}
	if len(emb) != 1536 {
		t.Fatalf("Expected 1536 dimensions, got %d", len(emb))
	}
	t.Logf("Embedding Dimension: %d", len(emb))
}
