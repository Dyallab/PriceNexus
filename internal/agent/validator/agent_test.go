package validator

import (
	"context"
	"testing"

	"github.com/dyallo/pricenexus/internal/agent/shared"
)

func TestNewValidatorAgent(t *testing.T) {
	t.Parallel()

	agent, err := NewValidatorAgent(nil)
	if err != nil {
		t.Fatalf("Error creating validator agent: %v", err)
	}
	if agent == nil {
		t.Fatal("Agent should not be nil")
	}
}

func TestValidatorValidate(t *testing.T) {
	t.Parallel()

	agent, err := NewValidatorAgent(nil)
	if err != nil {
		t.Fatalf("Error creating validator agent: %v", err)
	}

	results := []shared.SearchResult{
		{
			SearchTerm:  "ps5",
			ProductName: "Test Product",
			Price:       100.50,
			Currency:    "ARS",
			URL:         "https://example.com",
			ShopName:    "Test Shop",
			HasStock:    true,
			HasShipping: true,
		},
		{
			SearchTerm:  "ps5",
			ProductName: "",
			Price:       -10,
			URL:         "notaurl",
		},
	}

	validated, err := agent.Validate(context.Background(), results)
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	if len(validated) != 1 {
		t.Fatalf("expected 1 validated result, got %d", len(validated))
	}
}
