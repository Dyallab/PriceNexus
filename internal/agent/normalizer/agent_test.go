package normalizer

import (
	"testing"

	"github.com/dyallo/pricenexus/internal/agent/shared"
)

func TestNormalizeUsesSourceContext(t *testing.T) {
	t.Parallel()

	agent := NewAgent("ars")
	results := agent.Normalize([]shared.SearchResult{
		{
			ProductName: "Product",
			Price:       199999,
			Currency:    "",
			URL:         "",
			ShopName:    "Unknown",
		},
	}, SourceContext{
		Query:     "PlayStation 5",
		SourceURL: "https://www.mercadolibre.com.ar/ps5",
		Source: shared.SearchResult{
			ProductName: "PlayStation 5",
			ShopName:    "Mercadolibre",
		},
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	got := results[0]
	if got.SearchTerm != "PlayStation 5" {
		t.Fatalf("expected search term to use query, got %q", got.SearchTerm)
	}
	if got.ProductName != "PlayStation 5" {
		t.Fatalf("expected normalized product name, got %q", got.ProductName)
	}
	if got.ShopName != "Mercadolibre" {
		t.Fatalf("expected normalized shop name, got %q", got.ShopName)
	}
	if got.URL != "https://www.mercadolibre.com.ar/ps5" {
		t.Fatalf("expected source URL fallback, got %q", got.URL)
	}
	if got.Currency != "ARS" {
		t.Fatalf("expected default currency ARS, got %q", got.Currency)
	}
}
