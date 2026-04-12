package formfinder

import (
	"context"
	"testing"
)

func TestFormFinderDiscoverFromBaseURL(t *testing.T) {
	agent := NewFormFinderAgent()

	// Test discovering search from a real site
	results, err := agent.DiscoverFromBaseURL(context.Background(), "https://www.mexx.com.ar")
	if err != nil {
		t.Logf("Discovery error (expected for some sites): %v", err)
	}

	t.Logf("Found %d search URLs", len(results))
	for _, r := range results {
		t.Logf("  - %s (confidence: %.2f, source: %s)", r.URL, r.Confidence, r.Source)
	}
}

func TestFormFinderDiscoverSearchForms(t *testing.T) {
	agent := NewFormFinderAgent()

	// Test discovering forms on a page
	forms, err := agent.DiscoverSearchForms(context.Background(), "https://www.mexx.com.ar")
	if err != nil {
		t.Logf("Discovery error (expected for some sites): %v", err)
	}

	t.Logf("Found %d search forms", len(forms))
	for _, f := range forms {
		t.Logf("  - Action: %s, Method: %s, Confidence: %.2f", f.Action, f.Method, f.Confidence)
		for _, input := range f.Inputs {
			t.Logf("    Input: %s (type: %s)", input.Name, input.Type)
		}
	}
}

func TestFormFinderGenerateSearchURL(t *testing.T) {
	agent := NewFormFinderAgent()

	// Test URL generation with a known form structure
	form := FormInfo{
		Action:   "https://www.mexx.com.ar/buscar/",
		Method:   "get",
		IsSearch: true,
		Inputs: []InputField{
			{Name: "p", Type: "text"},
		},
	}

	url, err := agent.GenerateSearchURL(form, "RTX 4060")
	if err != nil {
		t.Fatalf("Error generating URL: %v", err)
	}

	expectedURL := "https://www.mexx.com.ar/buscar/?p=RTX+4060"
	if url != expectedURL {
		t.Errorf("Expected URL %q, got %q", expectedURL, url)
	}
}
