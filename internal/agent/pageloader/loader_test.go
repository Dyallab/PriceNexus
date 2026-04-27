package pageloader

import (
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewPageLoader(t *testing.T) {
	loader := NewPageLoader(nil)
	if loader == nil {
		t.Fatal("PageLoader should not be nil")
	}
	if loader.client == nil {
		t.Fatal("HTTP client should be initialized")
	}
}

func TestExtractSection(t *testing.T) {
	loader := NewPageLoader(nil)

	html := `<div class="product"><h1>Test Product</h1><p>Price: $100</p></div>`

	section := loader.ExtractSection(html, "div")
	if !strings.Contains(section, "Test Product") {
		t.Errorf("Expected section to contain 'Test Product', got: %s", section)
	}
}

func TestLoadHTMLHandlesGzipResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		_, _ = gz.Write([]byte(`<html><head><meta property="og:title" content="Compressed Product"></head><body>ok</body></html>`))
	}))
	defer server.Close()

	loader := NewPageLoader(nil)
	html, err := loader.LoadHTML(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("LoadHTML() unexpected error: %v", err)
	}

	if !strings.Contains(html, `Compressed Product`) {
		t.Fatalf("LoadHTML() = %q, want decompressed html", html)
	}
}
