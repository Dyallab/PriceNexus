package websearcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	agentruntime "github.com/dyallo/pricenexus/internal/agent"
	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
)

type urlResult struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Snippet string `json:"snippet,omitempty"`
}

type urlSearchResponse struct {
	URLs []urlResult `json:"urls"`
}

const urlSchemaName = "url_search_result"

var excludedStoreDomains = []string{"mercadolibre.com.ar"}

var urlSearchSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"urls": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type": "string",
					},
					"title": map[string]any{
						"type": "string",
					},
					"snippet": map[string]any{
						"type": "string",
					},
				},
				"required": []any{"url"},
			},
		},
	},
	"required": []any{"urls"},
}

const schemaFirstSearchPrompt = `Search for online stores selling "%s" in Argentina.
You must use the web search tool to find relevant product pages from Argentine online stores (.com.ar or .ar domains).
Focus on finding specific product pages or store search results.
Return a JSON object with a "urls" array containing objects with "url" (required), "title" (optional), and "snippet" (optional) fields.
Only include URLs from Argentine domains (.com.ar or .ar).`

const proseFallbackSearchPrompt = `Search for online stores selling "%s" in Argentina.
You must use the web search tool to find relevant product pages from Argentine online stores.
List all URLs you found from .com.ar or .ar domains, one per line.
Only include URLs from Argentine domains.`

// WebSearcherAgent searches the web for relevant store URLs and product pages
// using OpenRouter's web_search server tool
type WebSearcherAgent struct {
	llm             llms.Model
	logger          *logrus.Logger
	allowedDomains  []string
	defaultCurrency string
}

// NewWebSearcherAgent creates a new WebSearcherAgent with OpenRouter web_search capabilities
func NewWebSearcherAgent(llm llms.Model, allowedDomains []string, defaultCurrency string) (*WebSearcherAgent, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	if strings.TrimSpace(defaultCurrency) == "" {
		defaultCurrency = "ARS"
	}
	return &WebSearcherAgent{
		llm:             llm,
		logger:          logger,
		allowedDomains:  normalizeAllowedDomains(allowedDomains),
		defaultCurrency: strings.ToUpper(strings.TrimSpace(defaultCurrency)),
	}, nil
}

// Search performs a web search for the given query using OpenRouter's web_search tool.
// It first attempts to use structured JSON schema output, then falls back to text parsing.
func (w *WebSearcherAgent) Search(ctx context.Context, query string) ([]shared.SearchResult, error) {
	var results []shared.SearchResult
	var err error

	results, err = w.searchWithSchema(ctx, query)
	if err != nil {
		w.logger.Debugf("Schema-based search failed: %v", err)
	}

	if len(results) == 0 {
		w.logger.Debug("Schema search yielded no results, falling back to text parsing")
		results, err = w.searchWithTextFallback(ctx, query)
		if err != nil {
			return nil, err
		}
	}

	w.logger.Debugf("Final result count: %d", len(results))
	return results, nil
}

func (w *WebSearcherAgent) searchWithSchema(ctx context.Context, query string) ([]shared.SearchResult, error) {
	prompt := fmt.Sprintf(schemaFirstSearchPrompt, query)

	openRouterModel, ok := w.llm.(*agentruntime.OpenRouterModel)
	if !ok {
		w.logger.Debug("LLM is not an OpenRouterModel, skipping schema-based search")
		return nil, fmt.Errorf("LLM does not support CallWithJSONSchema")
	}

	response, err := openRouterModel.CallWithJSONSchema(ctx, prompt, urlSchemaName, urlSearchSchema)
	if err != nil {
		return nil, fmt.Errorf("schema-based search failed: %w", err)
	}

	w.logger.Debugf("Schema response:\n%s", response)

	var structuredResp urlSearchResponse
	if err := json.Unmarshal([]byte(response), &structuredResp); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	results := w.urlsToSearchResults(structuredResp.URLs, query)
	if len(results) == 0 {
		return nil, fmt.Errorf("structured response contained no URLs")
	}

	return results, nil
}

func (w *WebSearcherAgent) urlsToSearchResults(urlResults []urlResult, query string) []shared.SearchResult {
	results := make([]shared.SearchResult, 0, len(urlResults))
	seen := make(map[string]bool)

	for _, ur := range urlResults {
		if seen[ur.URL] {
			continue
		}
		if !isAllowedDomain(ur.URL, w.allowedDomains) {
			w.logger.Debugf("Filtered out non-Argentine URL from schema: %s", ur.URL)
			continue
		}
		seen[ur.URL] = true

		shopName := extractShopName(ur.URL)
		results = append(results, shared.SearchResult{
			SearchTerm:  query,
			ProductName: query,
			Price:       0,
			Currency:    w.defaultCurrency,
			URL:         ur.URL,
			HasStock:    false,
			HasShipping: false,
			ShopName:    shopName,
		})
	}

	return results
}

func (w *WebSearcherAgent) searchWithTextFallback(ctx context.Context, query string) ([]shared.SearchResult, error) {
	prompt := fmt.Sprintf(proseFallbackSearchPrompt, query)

	response, err := w.llm.Call(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("error searching web: %w", err)
	}

	w.logger.Debugf("Text fallback response:\n%s", response)

	results := w.parseSearchResponse(response, query)
	if len(results) == 0 {
		return nil, fmt.Errorf("no URLs found in response")
	}

	return results, nil
}

// SearchStores searches for online stores relevant to a product query
func (w *WebSearcherAgent) SearchStores(ctx context.Context, productQuery string) ([]shared.SearchResult, error) {
	return w.Search(ctx, productQuery)
}

// parseSearchResponse extracts URLs and product information from the model's response
func (w *WebSearcherAgent) parseSearchResponse(response string, query string) []shared.SearchResult {
	results := []shared.SearchResult{}

	// Extract URLs from the response
	urls := extractURLsForDomains(response, w.allowedDomains)
	w.logger.Debugf("Found %d URLs after extraction: %v", len(urls), urls)

	for _, url := range urls {
		shopName := extractShopName(url)
		results = append(results, shared.SearchResult{
			SearchTerm:  query,
			ProductName: query,
			Price:       0, // Will be filled by data extractor
			Currency:    w.defaultCurrency,
			URL:         url,
			HasStock:    false,
			HasShipping: false,
			ShopName:    shopName,
		})
	}

	return results
}

// extractURLs extracts all URLs from the response text
func extractURLs(text string) []string {
	return extractURLsForDomains(text, []string{".com.ar", ".ar"})
}

func extractURLsForDomains(text string, allowedDomains []string) []string {
	urls := []string{}
	seen := make(map[string]bool)

	text = extractQuotedURLs(text, allowedDomains, &seen, &urls)

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			endIdx := strings.IndexAny(line, " \t)")
			if endIdx == -1 {
				endIdx = len(line)
			}
			url := line[:endIdx]
			url = strings.TrimRight(url, ".,;:!?)\"'")

			if isAllowedDomain(url, allowedDomains) && !seen[url] {
				urls = append(urls, url)
				seen[url] = true
			}
			continue
		}

		if strings.Contains(line, "http") {
			parts := strings.Fields(line)
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://") {
					part = strings.TrimRight(part, ".,;:!?)\"'")

					if isAllowedDomain(part, allowedDomains) {
						if !seen[part] {
							urls = append(urls, part)
							seen[part] = true
						}
					}
				}
			}
		}
	}

	return urls
}

func extractQuotedURLs(text string, allowedDomains []string, seen *map[string]bool, urls *[]string) string {
	arPattern := regexp.MustCompile(`"https?://[^"]+\.ar[^"]*"|'https?://[^']+\.ar[^']*'|\(https?://[^)]+\.ar[^)]*\)`)
	comArPattern := regexp.MustCompile(`"https?://[^"]+\.com\.ar[^"]*"|'https?://[^']+\.com\.ar[^']*'|\(https?://[^)]+\.com\.ar[^)]*\)`)

	matches := comArPattern.FindAllString(text, -1)
	for _, match := range matches {
		match = strings.Trim(match, `"'()`)
		if isAllowedDomain(match, allowedDomains) && !(*seen)[match] {
			*urls = append(*urls, match)
			(*seen)[match] = true
		}
	}
	text = comArPattern.ReplaceAllString(text, " ")

	matches = arPattern.FindAllString(text, -1)
	for _, match := range matches {
		match = strings.Trim(match, `"'()`)
		if isAllowedDomain(match, allowedDomains) && !(*seen)[match] {
			*urls = append(*urls, match)
			(*seen)[match] = true
		}
	}
	text = arPattern.ReplaceAllString(text, " ")

	return text
}

// isArgentinianDomain checks if a URL belongs to an Argentine domain
func isArgentinianDomain(urlStr string) bool {
	return isAllowedDomain(urlStr, []string{".com.ar", ".ar"})
}

func isAllowedDomain(urlStr string, allowedDomains []string) bool {
	if isExcludedDomain(urlStr, excludedStoreDomains) {
		return false
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	hostname := strings.ToLower(parsedURL.Hostname())
	for _, domain := range normalizeAllowedDomains(allowedDomains) {
		if strings.HasSuffix(hostname, domain) {
			return true
		}
	}

	return false
}

func isExcludedDomain(urlStr string, excludedDomains []string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	hostname := strings.ToLower(parsedURL.Hostname())
	for _, domain := range excludedDomains {
		normalized := strings.ToLower(strings.TrimSpace(domain))
		normalized = strings.TrimPrefix(normalized, ".")
		if normalized == "" {
			continue
		}
		if hostname == normalized || strings.HasSuffix(hostname, "."+normalized) {
			return true
		}
	}

	return false
}

func normalizeAllowedDomains(domains []string) []string {
	if len(domains) == 0 {
		return []string{".com.ar", ".ar"}
	}

	result := make([]string, 0, len(domains))
	seen := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		normalized := strings.ToLower(strings.TrimSpace(domain))
		if normalized == "" {
			continue
		}
		if !strings.HasPrefix(normalized, ".") {
			normalized = "." + normalized
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	if len(result) == 0 {
		return []string{".com.ar", ".ar"}
	}

	return result
}

// extractShopName extracts a readable shop name from a URL
func extractShopName(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err == nil {
		hostname := strings.TrimPrefix(parsedURL.Hostname(), "www.")
		if hostname != "" && strings.Contains(hostname, ".") {
			hostname = strings.TrimSuffix(hostname, ".com.ar")
			hostname = strings.TrimSuffix(hostname, ".com")
			hostname = strings.TrimSuffix(hostname, ".ar")
			if hostname != "" {
				return capitalize(hostname)
			}
		}
	}

	return "Unknown"
}

// capitalize makes the first letter uppercase
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
