package dto_test

import (
	"encoding/json"
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/generation/dto"
)

func TestCreateGenerationRequestUsesPromptField(t *testing.T) {
	var request dto.CreateGenerationRequest
	if err := json.Unmarshal([]byte(`{"prompt":"hero"}`), &request); err != nil {
		t.Fatalf("unmarshal generation request: %v", err)
	}

	if request.Prompt != "hero" {
		t.Fatalf("expected prompt %q, got %q", "hero", request.Prompt)
	}
}
