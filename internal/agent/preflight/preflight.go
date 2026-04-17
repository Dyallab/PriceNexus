package preflight

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type URLStatus int

const (
	StatusUnchecked URLStatus = iota
	StatusValid
	StatusUnresolvableHost
	StatusConnectionFailed
	StatusTimeout
	StatusTooManyRedirects
	StatusRedirectLimitExceeded
	StatusForbidden
	StatusNotFound
	StatusGone
	StatusServerError
	StatusSoftMissing
	StatusInvalidURL
)

func (s URLStatus) String() string {
	switch s {
	case StatusUnchecked:
		return "unchecked"
	case StatusValid:
		return "valid"
	case StatusUnresolvableHost:
		return "unresolvable_host"
	case StatusConnectionFailed:
		return "connection_failed"
	case StatusTimeout:
		return "timeout"
	case StatusTooManyRedirects:
		return "too_many_redirects"
	case StatusRedirectLimitExceeded:
		return "redirect_limit_exceeded"
	case StatusForbidden:
		return "forbidden"
	case StatusNotFound:
		return "not_found"
	case StatusGone:
		return "gone"
	case StatusServerError:
		return "server_error"
	case StatusSoftMissing:
		return "soft_missing"
	case StatusInvalidURL:
		return "invalid_url"
	default:
		return "unknown"
	}
}

type PreflightResult struct {
	OriginalURL  string
	FinalURL     string
	Status       URLStatus
	StatusCode   int
	ResolvedHost string
	ErrorMsg     string
	SoftMissing  bool
}

func (r PreflightResult) IsValid() bool {
	return r.Status == StatusValid && r.StatusCode == http.StatusOK
}

func (r PreflightResult) ShouldExtract() bool {
	return r.Status == StatusValid
}

func (r PreflightResult) IsGone() bool {
	return r.Status == StatusGone
}

type Preflighter struct {
	client          *http.Client
	softDetectText  []string
	followRedirects bool
	maxRedirects    int
}

type Option func(*Preflighter)

func WithSoftDetection(patterns []string) Option {
	return func(p *Preflighter) {
		p.softDetectText = patterns
	}
}

func WithMaxRedirects(n int) Option {
	return func(p *Preflighter) {
		p.maxRedirects = n
	}
}

func WithClient(client *http.Client) Option {
	return func(p *Preflighter) {
		p.client = client
	}
}

func NewPreflighter(opts ...Option) *Preflighter {
	p := &Preflighter{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		softDetectText:  defaultSoftPatterns(),
		followRedirects: true,
		maxRedirects:    10,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func defaultSoftPatterns() []string {
	return []string{
		"no se encontr",
		"no encontrado",
		"producto no encontrado",
		"artículo no encontrado",
		"no results",
		"sin resultados",
		"0 resultados",
		"no hay productos",
		"no products found",
		"product not found",
		"404",
		"page not found",
		"página no encontrada",
		"error 404",
		"removed",
		"eliminado",
		"discontinued",
		"out of stock",
		"sin stock",
		"agotado",
	}
}

func (p *Preflighter) Check(ctx context.Context, rawURL string) PreflightResult {
	result := PreflightResult{
		OriginalURL: rawURL,
		Status:      StatusUnchecked,
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		result.Status = StatusInvalidURL
		result.ErrorMsg = fmt.Sprintf("parse error: %v", err)
		return result
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		result.Status = StatusInvalidURL
		result.ErrorMsg = "missing scheme or host"
		return result
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		result.Status = StatusInvalidURL
		result.ErrorMsg = fmt.Sprintf("unsupported scheme: %s", parsedURL.Scheme)
		return result
	}

	result.ResolvedHost = parsedURL.Host

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		result.Status = StatusInvalidURL
		result.ErrorMsg = fmt.Sprintf("request creation: %v", err)
		return result
	}

	setBrowserHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return p.handleRequestError(err, rawURL, parsedURL.Host)
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.FinalURL = resp.Request.URL.String()
	result.ResolvedHost = resp.Request.URL.Host

	switch {
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		return p.handleRedirect(resp, rawURL, ctx)

	case resp.StatusCode == http.StatusOK:
		result.Status = StatusValid
		return result

	case resp.StatusCode == http.StatusNotFound:
		result.Status = StatusNotFound
		if p.containsSoftPattern(resp) {
			result.Status = StatusSoftMissing
			result.SoftMissing = true
		}
		return result

	case resp.StatusCode == http.StatusGone:
		result.Status = StatusGone
		if p.containsSoftPattern(resp) {
			result.Status = StatusSoftMissing
			result.SoftMissing = true
		}
		return result

	case resp.StatusCode == http.StatusForbidden:
		result.Status = StatusForbidden
		return result

	case resp.StatusCode >= 500:
		result.Status = StatusServerError
		return result

	default:
		result.Status = StatusUnchecked
		return result
	}
}

func (p *Preflighter) handleRequestError(err error, rawURL, host string) PreflightResult {
	result := PreflightResult{
		OriginalURL:  rawURL,
		ResolvedHost: host,
		Status:       StatusConnectionFailed,
	}

	if strings.Contains(err.Error(), "timeout") {
		result.Status = StatusTimeout
		result.ErrorMsg = "request timeout"
	} else if strings.Contains(err.Error(), "no such host") ||
		strings.Contains(err.Error(), "lookup") ||
		strings.Contains(err.Error(), "server can't find") {
		result.Status = StatusUnresolvableHost
		result.ErrorMsg = fmt.Sprintf("host not found: %s", host)
	} else {
		result.ErrorMsg = err.Error()
	}

	return result
}

func (p *Preflighter) containsSoftPattern(resp *http.Response) bool {
	bodyContent := make([]byte, 8192)
	n, _ := resp.Body.Read(bodyContent)
	if n == 0 {
		return false
	}
	contentLower := strings.ToLower(string(bodyContent[:n]))
	for _, pattern := range p.softDetectText {
		if strings.Contains(contentLower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func (p *Preflighter) handleRedirect(resp *http.Response, rawURL string, ctx context.Context) PreflightResult {
	result := PreflightResult{
		OriginalURL: rawURL,
		StatusCode:  resp.StatusCode,
		FinalURL:    resp.Request.URL.String(),
	}

	result.ResolvedHost = resp.Request.URL.Host

	if !p.followRedirects {
		result.Status = StatusValid
		return result
	}

	location := resp.Header.Get("Location")
	if location == "" {
		result.Status = StatusTooManyRedirects
		result.ErrorMsg = "no location header in redirect"
		return result
	}

	redirectCount := 1
	currentURL := resp.Request.URL.String()

	for redirectCount < p.maxRedirects {
		redirectURL, err := resp.Request.URL.Parse(location)
		if err != nil {
			result.Status = StatusTooManyRedirects
			result.ErrorMsg = fmt.Sprintf("invalid redirect URL: %v", err)
			return result
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, redirectURL.String(), nil)
		if err != nil {
			result.Status = StatusTooManyRedirects
			result.ErrorMsg = fmt.Sprintf("redirect request creation: %v", err)
			return result
		}

		setBrowserHeaders(req)

		resp, err := p.client.Do(req)
		if err != nil {
			return p.handleRequestError(err, currentURL, redirectURL.Host)
		}
		defer resp.Body.Close()

		result.StatusCode = resp.StatusCode
		result.FinalURL = resp.Request.URL.String()
		result.ResolvedHost = resp.Request.URL.Host

		switch {
		case resp.StatusCode >= 300 && resp.StatusCode < 400:
			location = resp.Header.Get("Location")
			if location == "" {
				result.Status = StatusTooManyRedirects
				result.ErrorMsg = "no location header in redirect chain"
				return result
			}
			redirectCount++
			currentURL = resp.Request.URL.String()

		case resp.StatusCode == http.StatusOK:
			result.Status = StatusValid
			return result

		case resp.StatusCode == http.StatusNotFound:
			result.Status = StatusNotFound
			if p.containsSoftPattern(resp) {
				result.Status = StatusSoftMissing
				result.SoftMissing = true
			}
			return result

		case resp.StatusCode == http.StatusGone:
			result.Status = StatusGone
			if p.containsSoftPattern(resp) {
				result.Status = StatusSoftMissing
				result.SoftMissing = true
			}
			return result

		case resp.StatusCode == http.StatusForbidden:
			result.Status = StatusForbidden
			return result

		case resp.StatusCode >= 500:
			result.Status = StatusServerError
			return result

		default:
			result.Status = StatusUnchecked
			return result
		}
	}

	result.Status = StatusRedirectLimitExceeded
	result.ErrorMsg = fmt.Sprintf("exceeded max redirects (%d)", p.maxRedirects)
	return result
}

func (p *Preflighter) CheckWithSoftDetection(ctx context.Context, rawURL string) (PreflightResult, error) {
	return p.Check(ctx, rawURL), nil
}

func setBrowserHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "max-age=0")
}

func CheckURL(ctx context.Context, rawURL string) PreflightResult {
	return NewPreflighter().Check(ctx, rawURL)
}

func CheckURLWithSoftDetection(ctx context.Context, rawURL string) (PreflightResult, error) {
	return NewPreflighter().CheckWithSoftDetection(ctx, rawURL)
}

func FilterValidURLs(ctx context.Context, urls []string) []PreflightResult {
	preflight := NewPreflighter()
	results := make([]PreflightResult, 0, len(urls))

	for _, u := range urls {
		result := preflight.Check(ctx, u)
		if result.ShouldExtract() {
			results = append(results, result)
		}
	}

	return results
}

func ExtractValidURLs(results []PreflightResult) []string {
	validURLs := make([]string, 0, len(results))
	for _, r := range results {
		if r.ShouldExtract() {
			validURLs = append(validURLs, r.FinalURL)
		}
	}
	return validURLs
}
