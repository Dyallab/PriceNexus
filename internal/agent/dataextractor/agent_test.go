package dataextractor

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dyallo/pricenexus/internal/agent/shared"
)

func TestParseHTMLFragmentOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		output    string
		wanted    string
		wantError bool
	}{
		{
			name:   "json object",
			output: `{"html_fragment":"<div>product</div>"}`,
			wanted: "<div>product</div>",
		},
		{
			name:      "missing field",
			output:    `{"products":[]}`,
			wantError: true,
		},
		{
			name:   "relaxed malformed escape in fragment",
			output: `{"html_fragment": "<div data-product=\\\"{\\&quot;price\\&quot;:123}\\\">Producto</div>"}`,
			wanted: `<div data-product="{&quot;price&quot;:123}">Producto</div>`,
		},
		{
			name:   "unterminated fragment string fallback",
			output: `{"html_fragment": "<div class=\\\"product\\\">Producto`,
			wanted: `<div class="product">Producto`,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseHTMLFragmentOutput(tc.output)
			if tc.wantError {
				if err == nil {
					t.Fatal("parseHTMLFragmentOutput() error = nil, want non-nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseHTMLFragmentOutput() unexpected error: %v", err)
			}

			if got != tc.wanted {
				t.Fatalf("parseHTMLFragmentOutput() = %q, want %q", got, tc.wanted)
			}
		})
	}
}

func TestParseProductsOutput(t *testing.T) {
	t.Parallel()

	wanted := []shared.SearchResult{
		{
			ProductName: "Game Stick Lite",
			Price:       39000,
			Currency:    "ARS",
			URL:         "",
			HasStock:    false,
			HasShipping: true,
		},
	}

	tests := []struct {
		name      string
		output    string
		wanted    []shared.SearchResult
		wantError bool
	}{
		{
			name:   "products wrapper",
			output: `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"","has_stock":false,"has_shipping":true}]}`,
			wanted: wanted,
		},
		{
			name:      "empty products array",
			output:    `{"products":[]}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseProductsOutput(tc.output)
			if tc.wantError {
				if err == nil {
					t.Fatal("parseProductsOutput() error = nil, want non-nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseProductsOutput() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tc.wanted) {
				t.Fatalf("parseProductsOutput() = %#v, want %#v", got, tc.wanted)
			}
		})
	}
}

func TestParseProductsOutputTruncated(t *testing.T) {
	t.Parallel()

	wanted := []shared.SearchResult{
		{
			ProductName: "Game Stick Lite",
			Price:       39000,
			Currency:    "ARS",
			URL:         "",
			HasStock:    false,
			HasShipping: true,
		},
	}

	tests := []struct {
		name      string
		output    string
		wanted    []shared.SearchResult
		wantError bool
	}{
		{
			name:   "truncated json - missing closing brace",
			output: `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"","has_stock":false,"has_shipping":true`,
			wanted: wanted,
		},
		{
			name:   "truncated json - missing array close",
			output: `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"","has_stock":false,"has_shipping":true}]`,
			wanted: wanted,
		},
		{
			name:   "truncated json - trailing garbage after valid json",
			output: `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"","has_stock":false,"has_shipping":true}]}  extra garbage here`,
			wanted: wanted,
		},
		{
			name:   "truncated json - invalid trailing content",
			output: `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"","has_stock":false,"has_shipping":true}]}  undefined is not defined`,
			wanted: wanted,
		},
		{
			name:   "truncated json - with JSON prefix noise",
			output: `Here is the result: {"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"","has_stock":false,"has_shipping":true}]}`,
			wanted: wanted,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseProductsOutput(tc.output)
			if tc.wantError {
				if err == nil {
					t.Fatal("parseProductsOutput() error = nil, want non-nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseProductsOutput() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tc.wanted) {
				t.Fatalf("parseProductsOutput() = %#v, want %#v", got, tc.wanted)
			}
		})
	}
}

func TestIsTruncationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "unexpected end of JSON input",
			err:  errors.New("unexpected end of JSON input"),
			want: true,
		},
		{
			name: "unexpected EOF",
			err:  errors.New("unexpected EOF while scanning string"),
			want: true,
		},
		{
			name: "unclosed string",
			err:  errors.New("unclosed string"),
			want: true,
		},
		{
			name: "invalid character after object",
			err:  errors.New("invalid character 'x' after object"),
			want: true,
		},
		{
			name: "invalid character 'a'",
			err:  errors.New("invalid character 'a'"),
			want: false,
		},
		{
			name: "other error",
			err:  errors.New("something else went wrong"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isTruncationError(tc.err)
			if got != tc.want {
				t.Fatalf("isTruncationError() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTrimTrailingGarbage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "valid json unchanged",
			input:  `{"products":[]}`,
			output: `{"products":[]}`,
		},
		{
			name:   "truncated - missing closing brace",
			input:  `{"products":[{"product_name":"Test","price":100,"currency":"ARS","url":"","has_stock":true,"has_shipping":false}`,
			output: `{"products":[{"product_name":"Test","price":100,"currency":"ARS","url":"","has_stock":true,"has_shipping":false}]}`,
		},
		{
			name:   "starts with text then json",
			input:  `Here is the result: {"products":[]}`,
			output: `Here is the result: {"products":[]}`,
		},
		{
			name:   "empty string",
			input:  ``,
			output: ``,
		},
		{
			name:   "only whitespace",
			input:  `   `,
			output: `   `,
		},
		{
			name:   "array json",
			input:  `[{"product_name":"Test","price":100,"currency":"ARS","url":"","has_stock":true,"has_shipping":false}]`,
			output: `[{"product_name":"Test","price":100,"currency":"ARS","url":"","has_stock":true,"has_shipping":false}]`,
		},
		{
			name:   "truncated array",
			input:  `[{"product_name":"Test","price":100,"currency":"ARS","url":"","has_stock":true,"has_shipping":false}`,
			output: `[{"product_name":"Test","price":100,"currency":"ARS","url":"","has_stock":true,"has_shipping":false}]`,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := trimTrailingGarbage(tc.input)
			if got != tc.output {
				t.Fatalf("trimTrailingGarbage(%q) = %q, want %q", tc.input, got, tc.output)
			}
		})
	}
}

func TestTryParseProductsOutput(t *testing.T) {
	t.Parallel()

	wanted := []shared.SearchResult{
		{
			ProductName: "Test Product",
			Price:       15000,
			Currency:    "ARS",
			URL:         "https://example.com",
			HasStock:    true,
			HasShipping: false,
		},
	}

	tests := []struct {
		name      string
		jsonStr   string
		wanted    []shared.SearchResult
		wantError bool
	}{
		{
			name:    "valid products wrapper",
			jsonStr: `{"products":[{"product_name":"Test Product","price":15000,"currency":"ARS","url":"https://example.com","has_stock":true,"has_shipping":false}]}`,
			wanted:  wanted,
		},
		{
			name:      "invalid json - syntax error",
			jsonStr:   `{"products":[}`,
			wantError: true,
		},
		{
			name:      "empty products",
			jsonStr:   `{"products":[]}`,
			wantError: true,
		},
		{
			name:      "non-json string",
			jsonStr:   `this is not json`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tryParseProductsOutput(tc.jsonStr)
			if tc.wantError {
				if err == nil {
					t.Fatal("tryParseProductsOutput() error = nil, want non-nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("tryParseProductsOutput() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tc.wanted) {
				t.Fatalf("tryParseProductsOutput() = %#v, want %#v", got, tc.wanted)
			}
		})
	}
}

func TestExtractJSONLDFromGraphProduct(t *testing.T) {
	t.Parallel()

	agent := &DataExtractorAgent{}
	html := `<script type="application/ld+json">{
		"@context":"https://schema.org",
		"@graph":[
			{"@type":"BreadcrumbList","itemListElement":[]},
			{"@type":"Product","name":"Consola Game Stick Lite 4K","offers":{"@type":"Offer","price":"89928.06","priceCurrency":"ARS","availability":"https://schema.org/InStock"}}
		]
	}</script>`

	results := agent.extractJSONLD(html)
	if len(results) != 1 {
		t.Fatalf("extractJSONLD() len = %d, want 1", len(results))
	}

	if results[0].ProductName != "Consola Game Stick Lite 4K" {
		t.Fatalf("ProductName = %q, want %q", results[0].ProductName, "Consola Game Stick Lite 4K")
	}

	if results[0].Price != 89928.06 {
		t.Fatalf("Price = %v, want %v", results[0].Price, 89928.06)
	}

	if results[0].Currency != "ARS" {
		t.Fatalf("Currency = %q, want %q", results[0].Currency, "ARS")
	}

	if !results[0].HasStock {
		t.Fatal("HasStock = false, want true")
	}
}

func TestRepairProductsJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "reconstruct from complete first object",
			input:  `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"","has_stock":false,"has_shipping":true},{"product_name":"Partial`,
			output: `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"","has_stock":false,"has_shipping":true}]}`,
		},
		{
			name:   "sanitize invalid escape before closing",
			input:  `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"https://example.com/\&foo","has_stock":false,"has_shipping":true}]}`,
			output: `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"https://example.com/&foo","has_stock":false,"has_shipping":true}]}`,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := repairProductsJSON(tc.input)
			if got != tc.output {
				t.Fatalf("repairProductsJSON() = %q, want %q", got, tc.output)
			}
		})
	}
}

func TestParseProductsOutputWithRetry(t *testing.T) {
	t.Parallel()

	wanted := []shared.SearchResult{
		{
			ProductName: "Game Stick Lite",
			Price:       39000,
			Currency:    "ARS",
			URL:         "",
			HasStock:    false,
			HasShipping: true,
		},
	}

	tests := []struct {
		name      string
		output    string
		maxRetry  int
		wanted    []shared.SearchResult
		wantError bool
	}{
		{
			name:     "first attempt success",
			output:   `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"","has_stock":false,"has_shipping":true}]}`,
			maxRetry: 2,
			wanted:   wanted,
		},
		{
			name:     "second attempt success after trim",
			output:   `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"","has_stock":false,"has_shipping":true}`,
			maxRetry: 2,
			wanted:   wanted,
		},
		{
			name:      "zero retries - fails on truncation",
			output:    `{"products":[{"product_name":"Game Stick Lite","price":39000,"currency":"ARS","url":"","has_stock":false,"has_shipping":true}`,
			maxRetry:  0,
			wantError: true,
		},
		{
			name:      "max retries exceeded",
			output:    `{"products":[`,
			maxRetry:  2,
			wantError: true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseProductsOutputWithRetry(tc.output, tc.maxRetry)
			if tc.wantError {
				if err == nil {
					t.Fatal("parseProductsOutputWithRetry() error = nil, want non-nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseProductsOutputWithRetry() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tc.wanted) {
				t.Fatalf("parseProductsOutputWithRetry() = %#v, want %#v", got, tc.wanted)
			}
		})
	}
}
