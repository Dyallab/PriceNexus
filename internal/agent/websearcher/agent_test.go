package websearcher

import (
	"reflect"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestIsArgentinianDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		url    string
		wanted bool
	}{
		{
			name:   "com ar domain",
			url:    "https://electronerd.com.ar/tienda/consolas-retro/game-stick-lite-64-gb/",
			wanted: true,
		},
		{
			name:   "ar domain without trailing slash in host",
			url:    "https://tienda.ar/productos/game-stick-lite",
			wanted: true,
		},
		{
			name:   "non argentinian domain",
			url:    "https://example.com/product/game-stick-lite",
			wanted: false,
		},
		{
			name:   "excluded mercadolibre domain",
			url:    "https://www.mercadolibre.com.ar/game-stick-lite",
			wanted: false,
		},
		{
			name:   "invalid url",
			url:    "notaurl",
			wanted: false,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isArgentinianDomain(tc.url)
			if got != tc.wanted {
				t.Fatalf("isArgentinianDomain(%q) = %v, want %v", tc.url, got, tc.wanted)
			}
		})
	}
}

func TestExtractURLs(t *testing.T) {
	t.Parallel()

	response := `1. https://electronerd.com.ar/tienda/consolas-retro/game-stick-lite-64-gb/
2. https://tienda.ar/productos/game-stick-lite
3. https://example.com/not-allowed
4. https://electronerd.com.ar/tienda/consolas-retro/game-stick-lite-64-gb/`

	wanted := []string{
		"https://electronerd.com.ar/tienda/consolas-retro/game-stick-lite-64-gb/",
		"https://tienda.ar/productos/game-stick-lite",
	}

	got := extractURLs(response)
	if !reflect.DeepEqual(got, wanted) {
		t.Fatalf("extractURLs() = %v, want %v", got, wanted)
	}
}

func TestExtractURLsForDomains(t *testing.T) {
	t.Parallel()

	response := `1. https://example.tienda.ar/producto
2. https://example.com.ar/producto
3. https://www.mercadolibre.com.ar/producto
4. https://example.com/producto`

	got := extractURLsForDomains(response, []string{".tienda.ar"})
	wanted := []string{"https://example.tienda.ar/producto"}

	if !reflect.DeepEqual(got, wanted) {
		t.Fatalf("extractURLsForDomains() = %v, want %v", got, wanted)
	}
}

func TestExtractQuotedURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		allowedDomains []string
		wantURLs       []string
	}{
		{
			name:           "double quoted URLs in prose",
			input:          `I found "https://electronerd.com.ar/product1" and "https://tienda.ar/product2" in my search`,
			allowedDomains: []string{".com.ar", ".ar"},
			wantURLs:       []string{"https://electronerd.com.ar/product1", "https://tienda.ar/product2"},
		},
		{
			name:           "single quoted URLs",
			input:          `Check out 'https://example.com.ar/test' for more info`,
			allowedDomains: []string{".com.ar", ".ar"},
			wantURLs:       []string{"https://example.com.ar/test"},
		},
		{
			name:           "parenthesized URLs",
			input:          `See (https://shop.com.ar/item) for details`,
			allowedDomains: []string{".com.ar", ".ar"},
			wantURLs:       []string{"https://shop.com.ar/item"},
		},
		{
			name:           "URLs embedded in longer prose with no actual URLs",
			input:          "I searched everywhere but only found planning prose and no actual product URLs",
			allowedDomains: []string{".com.ar", ".ar"},
			wantURLs:       nil,
		},
		{
			name:           "mixed quoted and line URLs",
			input:          "Found \"https://store.com.ar/item1\"\nhttps://another.ar/product\n\"https://nomatch.com/foo\"",
			allowedDomains: []string{".com.ar", ".ar"},
			wantURLs:       []string{"https://store.com.ar/item1"},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			seen := make(map[string]bool)
			var urls []string
			rest := extractQuotedURLs(tc.input, tc.allowedDomains, &seen, &urls)
			if !reflect.DeepEqual(urls, tc.wantURLs) {
				t.Errorf("extractQuotedURLs() urls = %v, want %v", urls, tc.wantURLs)
			}
			if tc.name == "mixed quoted and line URLs" && !strings.Contains(rest, "https://another.ar/product") {
				t.Fatalf("extractQuotedURLs() rest = %q, want to preserve unquoted URL", rest)
			}
		})
	}
}

func TestSearchWithSchemaFallsBackWhenStructuredURLsFiltered(t *testing.T) {
	t.Parallel()

	agent := &WebSearcherAgent{
		llm:             nil,
		logger:          logrus.New(),
		defaultCurrency: "ARS",
		allowedDomains:  []string{".com.ar", ".ar"},
	}

	results := agent.urlsToSearchResults([]urlResult{{URL: "https://example.com/product"}}, "Game Stick Lite")
	if len(results) != 0 {
		t.Fatalf("urlsToSearchResults() len = %d, want 0", len(results))
	}
}

func TestURLsToSearchResults(t *testing.T) {
	t.Parallel()

	agent := &WebSearcherAgent{
		llm:             nil,
		logger:          logrus.New(),
		defaultCurrency: "ARS",
		allowedDomains:  []string{".com.ar", ".ar"},
	}

	input := []urlResult{
		{URL: "https://electronerd.com.ar/product1", Title: "Product 1"},
		{URL: "https://tienda.ar/product2", Title: "Product 2"},
		{URL: "https://www.mercadolibre.com.ar/product3", Title: "Filtered"},
		{URL: "https://example.com/not-allowed", Title: "Not Allowed"},
	}

	results := agent.urlsToSearchResults(input, "test query")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].URL != "https://electronerd.com.ar/product1" {
		t.Errorf("first URL = %q, want %q", results[0].URL, "https://electronerd.com.ar/product1")
	}
	if results[0].ShopName != "Electronerd" {
		t.Errorf("first ShopName = %q, want %q", results[0].ShopName, "Electronerd")
	}
	if results[1].URL != "https://tienda.ar/product2" {
		t.Errorf("second URL = %q, want %q", results[1].URL, "https://tienda.ar/product2")
	}
	if results[1].ShopName != "Tienda" {
		t.Errorf("second ShopName = %q, want %q", results[1].ShopName, "Tienda")
	}
}

func TestParseSearchResponse(t *testing.T) {
	t.Parallel()

	agent := &WebSearcherAgent{
		llm:             nil,
		logger:          logrus.New(),
		defaultCurrency: "ARS",
		allowedDomains:  []string{".com.ar", ".ar"},
	}

	response := `1. https://electronerd.com.ar/tienda/consolas-retro/game-stick-lite-64-gb/
2. https://tienda.ar/productos/game-stick-lite
3. https://www.mercadolibre.com.ar/game-stick-lite
4. https://example.com/not-allowed
5. https://electronerd.com.ar/tienda/consolas-retro/game-stick-lite-64-gb/`

	results := agent.parseSearchResponse(response, "Game Stick Lite")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].URL != "https://electronerd.com.ar/tienda/consolas-retro/game-stick-lite-64-gb/" {
		t.Errorf("first URL = %q, want %q", results[0].URL, "https://electronerd.com.ar/tienda/consolas-retro/game-stick-lite-64-gb/")
	}
	if results[0].SearchTerm != "Game Stick Lite" {
		t.Errorf("SearchTerm = %q, want %q", results[0].SearchTerm, "Game Stick Lite")
	}
	if results[0].Currency != "ARS" {
		t.Errorf("Currency = %q, want %q", results[0].Currency, "ARS")
	}
}

func TestParseSearchResponseWithQuotedURLs(t *testing.T) {
	t.Parallel()

	agent := &WebSearcherAgent{
		llm:             nil,
		logger:          logrus.New(),
		defaultCurrency: "ARS",
		allowedDomains:  []string{".com.ar", ".ar"},
	}

	response := `I searched for Game Stick Lite and found these stores:
"https://electronerd.com.ar/product1"
The second result is 'https://tienda.ar/product2'
(Also check https://store.com.ar/item for pricing)

No other useful results were found.`

	results := agent.parseSearchResponse(response, "Game Stick Lite")

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	expectedURLs := map[string]bool{
		"https://electronerd.com.ar/product1": true,
		"https://tienda.ar/product2":          true,
		"https://store.com.ar/item":           true,
	}

	for _, r := range results {
		if !expectedURLs[r.URL] {
			t.Errorf("unexpected URL: %s", r.URL)
		}
	}
}

func TestParseSearchResponseNoURLs(t *testing.T) {
	t.Parallel()

	agent := &WebSearcherAgent{
		llm:             nil,
		logger:          logrus.New(),
		defaultCurrency: "ARS",
		allowedDomains:  []string{".com.ar", ".ar"},
	}

	response := `I searched extensively but couldn't find any Argentine stores with this product in stock.
The search was difficult and I had no success.`

	results := agent.parseSearchResponse(response, "Nonexistent Product")

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestExtractShopName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url    string
		expect string
	}{
		{"https://electronerd.com.ar/product", "Electronerd"},
		{"https://www.mercadolibre.com.ar/item", "Mercadolibre"},
		{"https://tienda.ar/product", "Tienda"},
		{"https://www.fravega.com/p/producto", "Fravega"},
		{"https://example.com.ar/something", "Example"},
		{"https://not-a-url", "Unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			got := extractShopName(tc.url)
			if got != tc.expect {
				t.Errorf("extractShopName(%q) = %q, want %q", tc.url, got, tc.expect)
			}
		})
	}
}

func TestURLExtractionEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		allowedDomains []string
		wantLen        int
	}{
		{
			name:           "URL with query params",
			input:          "https://store.com.ar/product?id=123&ref=test",
			allowedDomains: []string{".com.ar", ".ar"},
			wantLen:        1,
		},
		{
			name:           "URL with port number",
			input:          "https://shop.com.ar:8080/product",
			allowedDomains: []string{".com.ar", ".ar"},
			wantLen:        1,
		},
		{
			name:           "duplicate URLs deduplicated",
			input:          "https://store.com.ar/product\nhttps://store.com.ar/product",
			allowedDomains: []string{".com.ar", ".ar"},
			wantLen:        1,
		},
		{
			name:           "non-Argentine filtered",
			input:          "https://store.com/product\nhttps://store.com.ar/product",
			allowedDomains: []string{".com.ar", ".ar"},
			wantLen:        1,
		},
		{
			name:           "mercadolibre filtered",
			input:          "https://www.mercadolibre.com.ar/product\nhttps://store.com.ar/product",
			allowedDomains: []string{".com.ar", ".ar"},
			wantLen:        1,
		},
		{
			name:           "trailing punctuation stripped",
			input:          "https://store.com.ar/product!",
			allowedDomains: []string{".com.ar", ".ar"},
			wantLen:        1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractURLsForDomains(tc.input, tc.allowedDomains)
			if len(got) != tc.wantLen {
				t.Errorf("extractURLsForDomains() len = %d, want %d", len(got), tc.wantLen)
			}
		})
	}
}
