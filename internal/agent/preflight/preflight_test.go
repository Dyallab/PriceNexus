package preflight

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestURLStatus_String(t *testing.T) {
	tests := []struct {
		status   URLStatus
		expected string
	}{
		{StatusUnchecked, "unchecked"},
		{StatusValid, "valid"},
		{StatusUnresolvableHost, "unresolvable_host"},
		{StatusNotFound, "not_found"},
		{StatusSoftMissing, "soft_missing"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.expected {
			t.Errorf("URLStatus(%d).String() = %q, want %q", tt.status, got, tt.expected)
		}
	}
}

func TestPreflightResult_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		result   PreflightResult
		expected bool
	}{
		{
			name:     "valid with 200",
			result:   PreflightResult{Status: StatusValid, StatusCode: http.StatusOK},
			expected: true,
		},
		{
			name:     "valid with soft missing",
			result:   PreflightResult{Status: StatusSoftMissing},
			expected: false,
		},
		{
			name:     "not found",
			result:   PreflightResult{Status: StatusNotFound, StatusCode: http.StatusNotFound},
			expected: false,
		},
		{
			name:     "server error",
			result:   PreflightResult{Status: StatusServerError, StatusCode: http.StatusInternalServerError},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPreflightResult_ShouldExtract(t *testing.T) {
	tests := []struct {
		name     string
		result   PreflightResult
		expected bool
	}{
		{
			name:     "valid should extract",
			result:   PreflightResult{Status: StatusValid},
			expected: true,
		},
		{
			name:     "soft missing should not extract",
			result:   PreflightResult{Status: StatusSoftMissing},
			expected: false,
		},
		{
			name:     "not found should not extract",
			result:   PreflightResult{Status: StatusNotFound},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.ShouldExtract(); got != tt.expected {
				t.Errorf("ShouldExtract() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewPreflighter(t *testing.T) {
	p := NewPreflighter()
	if p == nil {
		t.Fatal("NewPreflighter should not return nil")
	}
	if p.client == nil {
		t.Error("HTTP client should be initialized")
	}
	if p.maxRedirects != 10 {
		t.Errorf("default maxRedirects = %d, want 10", p.maxRedirects)
	}
}

func TestNewPreflighter_WithOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 5 * time.Second}
	p := NewPreflighter(
		WithClient(customClient),
		WithMaxRedirects(5),
		WithSoftDetection([]string{"custom pattern"}),
	)

	if p.client != customClient {
		t.Error("custom client not applied")
	}
	if p.maxRedirects != 5 {
		t.Errorf("maxRedirects = %d, want 5", p.maxRedirects)
	}
	if len(p.softDetectText) != 1 || p.softDetectText[0] != "custom pattern" {
		t.Errorf("softDetectText = %v, want [custom pattern]", p.softDetectText)
	}
}

func TestCheck_ValidURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body>Product Page</body></html>`))
	}))
	defer server.Close()

	p := NewPreflighter()
	result := p.Check(context.Background(), server.URL)

	if !result.IsValid() {
		t.Errorf("expected valid result, got status=%s code=%d", result.Status, result.StatusCode)
	}
	if result.FinalURL == "" {
		t.Error("FinalURL should be set")
	}
}

func TestCheck_404Page(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`<html><body>Some content</body></html>`))
	}))
	defer server.Close()

	p := NewPreflighter()
	result := p.Check(context.Background(), server.URL)

	if result.Status != StatusNotFound {
		t.Errorf("expected StatusNotFound, got %s", result.Status)
	}
	if result.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code 404, got %d", result.StatusCode)
	}
}

func TestCheck_GonePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte(`<html><body>Resource unavailable</body></html>`))
	}))
	defer server.Close()

	p := NewPreflighter()
	result := p.Check(context.Background(), server.URL)

	if result.Status != StatusGone {
		t.Errorf("expected StatusGone, got %s", result.Status)
	}
}

func TestCheck_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p := NewPreflighter()
	result := p.Check(context.Background(), server.URL)

	if result.Status != StatusServerError {
		t.Errorf("expected StatusServerError, got %s", result.Status)
	}
}

func TestCheck_InvalidURL(t *testing.T) {
	p := NewPreflighter()

	tests := []struct {
		url      string
		expected string
	}{
		{"", "missing scheme or host"},
		{"ftp://example.com", "unsupported scheme"},
	}

	for _, tt := range tests {
		result := p.Check(context.Background(), tt.url)
		if result.Status != StatusInvalidURL {
			t.Errorf("Check(%q) status = %s, want StatusInvalidURL", tt.url, result.Status)
		}
		if !strings.Contains(result.ErrorMsg, tt.expected) {
			t.Errorf("Check(%q) error = %q, want containing %q", tt.url, result.ErrorMsg, tt.expected)
		}
	}
}

func TestCheck_RedirectChain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		location := r.URL.Path
		switch location {
		case "/start":
			http.Redirect(w, r, "/middle", http.StatusMovedPermanently)
		case "/middle":
			http.Redirect(w, r, "/end", http.StatusFound)
		case "/end":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`Product`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	p := NewPreflighter()
	result := p.Check(context.Background(), server.URL+"/start")

	if result.Status != StatusValid {
		t.Errorf("expected StatusValid, got %s", result.Status)
	}
	if result.FinalURL == "" {
		t.Error("FinalURL should be set after redirect")
	}
	if result.OriginalURL != result.FinalURL {
		t.Logf("OriginalURL=%s FinalURL=%s", result.OriginalURL, result.FinalURL)
	}
}

func TestCheck_RedirectLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := r.URL.Path
		if current == "" || current == "/" {
			http.Redirect(w, r, "/1", http.StatusMovedPermanently)
			return
		}

		page := strings.TrimPrefix(current, "/")
		if page == "10" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next := strings.TrimPrefix(current, "/")
		http.Redirect(w, r, "/"+next+"1", http.StatusMovedPermanently)
	}))
	defer server.Close()

	p := NewPreflighter(WithMaxRedirects(5))
	result := p.Check(context.Background(), server.URL+"/")

	if result.Status != StatusRedirectLimitExceeded {
		t.Errorf("expected StatusRedirectLimitExceeded, got %s", result.Status)
	}
}

func TestCheck_NoRedirectFollow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/target", http.StatusMovedPermanently)
	}))
	defer server.Close()

	p := NewPreflighter()

	p.followRedirects = false
	result := p.Check(context.Background(), server.URL+"/")

	if result.Status != StatusValid {
		t.Errorf("expected StatusValid with redirects disabled, got %s", result.Status)
	}
}

func TestCheck_Timeout(t *testing.T) {
	p := NewPreflighter(WithClient(&http.Client{
		Timeout: 1 * time.Millisecond,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}))

	result := p.Check(context.Background(), "http://10.255.255.1:12345")

	if result.Status == StatusTimeout || result.Status == StatusConnectionFailed {
		t.Logf("Got expected timeout/connection error status: %s", result.Status)
	} else {
		t.Errorf("expected timeout status, got %s", result.Status)
	}
}

func TestCheck_UnresolvableHost(t *testing.T) {
	p := NewPreflighter()

	result := p.Check(context.Background(), "http://this-host-does-not-exist-12345.com/")

	if result.Status != StatusUnresolvableHost {
		t.Errorf("expected StatusUnresolvableHost, got %s", result.Status)
	}
}

func TestCheck_SoftDetection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`<html><body>Producto no encontrado</body></html>`))
	}))
	defer server.Close()

	p := NewPreflighter()
	result, err := p.CheckWithSoftDetection(context.Background(), server.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusSoftMissing {
		t.Errorf("expected StatusSoftMissing, got %s", result.Status)
	}
	if !result.SoftMissing {
		t.Error("expected SoftMissing to be true")
	}
}

func TestCheck_SoftDetection_NoMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`<html><body>Some random page content</body></html>`))
	}))
	defer server.Close()

	p := NewPreflighter()
	result, err := p.CheckWithSoftDetection(context.Background(), server.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusNotFound {
		t.Errorf("expected StatusNotFound, got %s", result.Status)
	}
}

func TestCheck_SoftDetection_ValidPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body>Producto encontrado - $100</body></html>`))
	}))
	defer server.Close()

	p := NewPreflighter()
	result, err := p.CheckWithSoftDetection(context.Background(), server.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusValid {
		t.Errorf("expected StatusValid, got %s", result.Status)
	}
}

func TestFilterValidURLs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid":
			w.WriteHeader(http.StatusOK)
		case "/404":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	base := server.URL
	urls := []string{
		base + "/valid",
		base + "/404",
		base + "/valid2",
	}

	results := FilterValidURLs(context.Background(), urls)

	if len(results) != 2 {
		t.Errorf("expected 2 valid results, got %d", len(results))
	}
}

func TestExtractValidURLs(t *testing.T) {
	results := []PreflightResult{
		{OriginalURL: "http://a.com/1", FinalURL: "http://a.com/1", Status: StatusValid},
		{OriginalURL: "http://b.com/2", FinalURL: "http://b.com/2", Status: StatusNotFound},
		{OriginalURL: "http://c.com/3", FinalURL: "http://c.com/3", Status: StatusSoftMissing},
		{OriginalURL: "http://d.com/4", FinalURL: "http://d.com/4", Status: StatusGone},
	}

	urls := ExtractValidURLs(results)

	if len(urls) != 1 {
		t.Errorf("expected 1 URL, got %d: %v", len(urls), urls)
	}
}

func TestCheckURL_Helper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := CheckURL(context.Background(), server.URL)

	if result.Status != StatusValid {
		t.Errorf("expected StatusValid, got %s", result.Status)
	}
}

func TestCheckURLWithSoftDetection_Helper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`<html><body>no se encontró</body></html>`))
	}))
	defer server.Close()

	result, err := CheckURLWithSoftDetection(context.Background(), server.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.SoftMissing {
		t.Error("expected SoftMissing to be true")
	}
}

func TestCheck_FinalURLNormalization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusMovedPermanently)
	}))
	defer server.Close()

	p := NewPreflighter()
	result := p.Check(context.Background(), server.URL+"/")

	if result.OriginalURL != result.FinalURL {
		t.Logf("OriginalURL=%s FinalURL=%s", result.OriginalURL, result.FinalURL)
	}
	if !strings.HasSuffix(result.FinalURL, "/final") {
		t.Errorf("FinalURL should end with /final, got %s", result.FinalURL)
	}
}
