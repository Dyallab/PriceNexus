package urlfinder

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms/ollama"
)

func TestNewURLFinderAgent(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel("gemma4:e4b"))
	if err != nil {
		t.Skip("Ollama not available, skipping test")
	}

	agent, err := NewURLFinderAgent(llm)
	if err != nil {
		t.Fatalf("Error creating URL finder agent: %v", err)
	}
	if agent == nil {
		t.Fatal("Agent should not be nil")
	}
}

func TestURLFinderFindURLs(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel("gemma4:e4b"))
	if err != nil {
		t.Skip("Ollama not available, skipping test")
	}

	agent, err := NewURLFinderAgent(llm)
	if err != nil {
		t.Fatalf("Error creating URL finder agent: %v", err)
	}

	// Test finding URLs for a product in all shops
	urls, err := agent.FindURLs(context.Background(), "Game Stick Lite", "")
	if err != nil {
		t.Fatalf("FindURLs failed: %v", err)
	}

	// Should generate URLs for configured shops
	if len(urls) == 0 {
		t.Fatal("Expected at least one URL, got empty slice")
	}

	t.Logf("Generated %d URLs:", len(urls))
	for _, u := range urls {
		t.Logf("  - %s", u)
	}
}

func TestURLFinderSpecificShop(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel("gemma4:e4b"))
	if err != nil {
		t.Skip("Ollama not available, skipping test")
	}

	agent, err := NewURLFinderAgent(llm)
	if err != nil {
		t.Fatalf("Error creating URL finder agent: %v", err)
	}

	// Test finding URLs for a specific shop
	urls, err := agent.FindURLs(context.Background(), "iPhone 15", "CompuGamer")
	if err != nil {
		t.Fatalf("FindURLs failed: %v", err)
	}

	if len(urls) != 1 {
		t.Fatalf("Expected 1 URL for specific shop, got %d", len(urls))
	}

	expectedURL := "https://www.compagamer.com/search?q=iPhone+15"
	if urls[0] != expectedURL {
		t.Errorf("Expected URL %q, got %q", expectedURL, urls[0])
	}
}

func TestURLFinderUnknownShop(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel("gemma4:e4b"))
	if err != nil {
		t.Skip("Ollama not available, skipping test")
	}

	agent, err := NewURLFinderAgent(llm)
	if err != nil {
		t.Fatalf("Error creating URL finder agent: %v", err)
	}

	// Test finding URLs for an unknown shop
	urls, err := agent.FindURLs(context.Background(), "test product", "unknown-shop")
	if err != nil {
		t.Fatalf("FindURLs failed: %v", err)
	}

	if len(urls) != 0 {
		t.Errorf("Expected empty slice for unknown shop, got %v", urls)
	}
}

func TestURLFinderAddStore(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel("gemma4:e4b"))
	if err != nil {
		t.Skip("Ollama not available, skipping test")
	}

	agent, err := NewURLFinderAgent(llm)
	if err != nil {
		t.Fatalf("Error creating URL finder agent: %v", err)
	}

	// Add a new store
	agent.AddStore("mitienda", "https://mitienda.com.ar", "search?q=%s", true)

	// Verify the store was added
	urls, err := agent.FindURLs(context.Background(), "test product", "mitienda")
	if err != nil {
		t.Fatalf("FindURLs failed: %v", err)
	}

	if len(urls) != 1 {
		t.Fatalf("Expected 1 URL for added shop, got %d", len(urls))
	}

	expectedURL := "https://mitienda.com.ar/search?q=test+product"
	if urls[0] != expectedURL {
		t.Errorf("Expected URL %q, got %q", expectedURL, urls[0])
	}
}

func TestURLFinderListStores(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel("gemma4:e4b"))
	if err != nil {
		t.Skip("Ollama not available, skipping test")
	}

	agent, err := NewURLFinderAgent(llm)
	if err != nil {
		t.Fatalf("Error creating URL finder agent: %v", err)
	}

	stores := agent.ListStores()
	if len(stores) == 0 {
		t.Fatal("Expected at least one store, got empty slice")
	}

	foundCompuGamer := false
	for _, store := range stores {
		if store.Name == "CompuGamer" {
			foundCompuGamer = true
			if !store.IsSmallShop {
				t.Error("Expected CompuGamer to be marked as small shop")
			}
		}
	}

	if !foundCompuGamer {
		t.Error("Expected CompuGamer to be in the store list")
	}
}

func TestURLFinderComputerStoreURLs(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel("gemma4:e4b"))
	if err != nil {
		t.Skip("Ollama not available, skipping test")
	}

	agent, err := NewURLFinderAgent(llm)
	if err != nil {
		t.Fatalf("Error creating URL finder agent: %v", err)
	}

	// Test computer store URLs
	tests := []struct {
		shopName string
		query    string
		expected string
	}{
		{"CompuGamer", "Game Stick Lite", "https://www.compagamer.com/search?q=Game+Stick+Lite"},
		{"CompuOro", "iPhone 15", "https://www.compuoro.com.ar/catalogsearch/result/?q=iPhone+15"},
		{"Mexx", "Laptop Dell", "https://www.mexx.com.ar/buscar/?p=Laptop+Dell"},
		{"Venex", "Teclado Mecanico", "https://www.venex.com.ar/busqueda?q=Teclado+Mecanico"},
		{"FullHard", "RTX 4060", "https://www.fullh4rd.com.ar/busqueda?q=RTX+4060"},
	}

	for _, tt := range tests {
		t.Run(tt.shopName, func(t *testing.T) {
			urls, err := agent.FindURLs(context.Background(), tt.query, tt.shopName)
			if err != nil {
				t.Fatalf("FindURLs failed: %v", err)
			}

			if len(urls) != 1 {
				t.Fatalf("Expected 1 URL for %s, got %d", tt.shopName, len(urls))
			}

			if urls[0] != tt.expected {
				t.Errorf("Expected URL %q, got %q", tt.expected, urls[0])
			}
		})
	}
}
