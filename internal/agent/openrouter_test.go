package agent

import (
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestExtractOpenRouterContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content any
		wanted  string
	}{
		{
			name:    "plain string",
			content: "hello",
			wanted:  "hello",
		},
		{
			name: "array text parts",
			content: []interface{}{
				map[string]interface{}{"text": "hello "},
				map[string]interface{}{"text": "world"},
			},
			wanted: "hello world",
		},
		{
			name: "array mixed parts",
			content: []interface{}{
				map[string]interface{}{"content": "first"},
				map[string]interface{}{"text": " second"},
			},
			wanted: "first second",
		},
		{
			name:    "unsupported type",
			content: map[string]interface{}{"text": "ignored"},
			wanted:  "",
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractOpenRouterContent(tc.content)
			if got != tc.wanted {
				t.Fatalf("extractOpenRouterContent() = %q, want %q", got, tc.wanted)
			}
		})
	}
}

func TestMapOpenRouterRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		role   llms.ChatMessageType
		wanted string
	}{
		{name: "human to user", role: llms.ChatMessageTypeHuman, wanted: "user"},
		{name: "ai to assistant", role: llms.ChatMessageTypeAI, wanted: "assistant"},
		{name: "system unchanged", role: llms.ChatMessageTypeSystem, wanted: "system"},
		{name: "tool unchanged", role: llms.ChatMessageTypeTool, wanted: "tool"},
		{name: "unknown passthrough", role: llms.ChatMessageType("developer"), wanted: "developer"},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mapOpenRouterRole(tc.role)
			if got != tc.wanted {
				t.Fatalf("mapOpenRouterRole(%q) = %q, want %q", tc.role, got, tc.wanted)
			}
		})
	}
}

func TestAddWebSearchToolUsesRuntimeConfig(t *testing.T) {
	t.Parallel()

	model := NewOpenRouterModel("key", "model")
	model.AddWebSearchTool(SearchConfig{
		AllowedDomains:  []string{".com.ar", ".tienda.ar"},
		ExcludedDomains: []string{"mercadolibre.com.ar"},
		SearchEngine:    "exa",
		MaxResults:      25,
	})

	if len(model.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(model.Tools))
	}

	parameters := model.Tools[0].Parameters
	gotDomains, ok := parameters["allowed_domains"].([]string)
	if !ok {
		t.Fatalf("expected []string allowed_domains, got %T", parameters["allowed_domains"])
	}

	if len(gotDomains) != 2 || gotDomains[0] != ".com.ar" || gotDomains[1] != ".tienda.ar" {
		t.Fatalf("unexpected allowed domains: %#v", gotDomains)
	}

	gotExcludedDomains, ok := parameters["excluded_domains"].([]string)
	if !ok {
		t.Fatalf("expected []string excluded_domains, got %T", parameters["excluded_domains"])
	}

	if len(gotExcludedDomains) != 1 || gotExcludedDomains[0] != "mercadolibre.com.ar" {
		t.Fatalf("unexpected excluded domains: %#v", gotExcludedDomains)
	}

	gotEngine, ok := parameters["engine"].(string)
	if !ok {
		t.Fatalf("expected string engine, got %T", parameters["engine"])
	}

	if gotEngine != "exa" {
		t.Fatalf("expected engine exa, got %q", gotEngine)
	}

	gotMaxResults, ok := parameters["max_results"].(int)
	if !ok {
		t.Fatalf("expected int max_results, got %T", parameters["max_results"])
	}

	if gotMaxResults != 25 {
		t.Fatalf("expected max_results 25, got %d", gotMaxResults)
	}
}
