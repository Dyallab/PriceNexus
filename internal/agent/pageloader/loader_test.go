package pageloader

import (
	"strings"
	"testing"
)

func TestNewPageLoader(t *testing.T) {
	loader := NewPageLoader()
	if loader == nil {
		t.Fatal("PageLoader should not be nil")
	}
	if loader.client == nil {
		t.Fatal("HTTP client should be initialized")
	}
}

func TestExtractSection(t *testing.T) {
	loader := NewPageLoader()

	html := `<div class="product"><h1>Test Product</h1><p>Price: $100</p></div>`

	section := loader.ExtractSection(html, "div")
	if !strings.Contains(section, "Test Product") {
		t.Errorf("Expected section to contain 'Test Product', got: %s", section)
	}
}
