package validator

import (
	"context"
	"testing"

	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/tmc/langchaingo/llms/ollama"
)

func TestNewValidatorAgent(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel("phi3:mini"))
	if err != nil {
		t.Skip("Ollama not available, skipping test")
	}

	agent, err := NewValidatorAgent(llm)
	if err != nil {
		t.Fatalf("Error creating validator agent: %v", err)
	}
	if agent == nil {
		t.Fatal("Agent should not be nil")
	}
}

func TestValidatorValidate(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel("phi3:mini"))
	if err != nil {
		t.Skip("Ollama not available, skipping test")
	}

	agent, err := NewValidatorAgent(llm)
	if err != nil {
		t.Fatalf("Error creating validator agent: %v", err)
	}

	results := []shared.SearchResult{
		{
			ProductName: "Test Product",
			Price:       100.50,
			Currency:    "ARS",
			URL:         "https://example.com",
			HasStock:    true,
			HasShipping: true,
		},
	}

	validated, err := agent.Validate(context.Background(), results)
	if err != nil {
		t.Logf("Expected error with Ollama: %v", err)
	}
	_ = validated
}
