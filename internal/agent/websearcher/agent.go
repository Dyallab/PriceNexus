package websearcher

import (
	"context"
	"fmt"
	"strings"

	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
)

// WebSearcherAgent searches the web for relevant store URLs and product pages
// using OpenRouter's web_search server tool
type WebSearcherAgent struct {
	llm    llms.Model
	logger *logrus.Logger
}

// NewWebSearcherAgent creates a new WebSearcherAgent with OpenRouter web_search capabilities
func NewWebSearcherAgent(llm llms.Model) (*WebSearcherAgent, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return &WebSearcherAgent{
		llm:    llm,
		logger: logger,
	}, nil
}

// Search performs a web search for the given query using OpenRouter's web_search tool
func (w *WebSearcherAgent) Search(ctx context.Context, query string) ([]shared.SearchResult, error) {
	// Create a prompt that instructs the LLM to search for products in Argentina
	prompt := fmt.Sprintf(`Search for online stores selling "%s" in Argentina.
You must use the web search tool to find relevant product pages from Argentine online stores.
Focus on finding specific product pages or store search results from Argentine stores (.com.ar or .ar domains).
For each result found, list the URL clearly.
Return the URLs you found.`, query)

	// Call the model which will use the web_search tool
	response, err := w.llm.Call(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("error searching web: %w", err)
	}

	w.logger.Debugf("Raw response from LLM:\n%s", response)

	// Parse the response to extract URLs
	results := w.parseSearchResponse(response, query)
	w.logger.Debugf("Parsed %d results from response", len(results))
	return results, nil
}

// SearchStores searches for online stores relevant to a product query
func (w *WebSearcherAgent) SearchStores(ctx context.Context, productQuery string) ([]shared.SearchResult, error) {
	return w.Search(ctx, productQuery)
}

// parseSearchResponse extracts URLs and product information from the model's response
func (w *WebSearcherAgent) parseSearchResponse(response string, query string) []shared.SearchResult {
	var results []shared.SearchResult

	// Extract URLs from the response
	urls := extractURLs(response)
	w.logger.Debugf("Found %d URLs after extraction: %v", len(urls), urls)

	for _, url := range urls {
		shopName := extractShopName(url)
		results = append(results, shared.SearchResult{
			ProductName: query,
			Price:       0, // Will be filled by data extractor
			Currency:    "ARS",
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
	var urls []string
	seen := make(map[string]bool)

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for URLs starting with http
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			// Extract just the URL part (stop at whitespace or closing bracket)
			endIdx := strings.IndexAny(line, " \t)")
			if endIdx == -1 {
				endIdx = len(line)
			}
			url := line[:endIdx]
			url = strings.TrimRight(url, ".,;:!?)\"'")

			// Filter for Argentine domains only
			if isArgentinianDomain(url) && !seen[url] {
				urls = append(urls, url)
				seen[url] = true
			}
			continue
		}

		// Also check for URLs within text
		if strings.Contains(line, "http") {
			parts := strings.Fields(line)
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://") {
					// Clean up trailing punctuation
					part = strings.TrimRight(part, ".,;:!?)\"'")

					// Log all found URLs before filtering
					if isArgentinianDomain(part) {
						if !seen[part] {
							urls = append(urls, part)
							seen[part] = true
						}
					} else {
						logrus.Debugf("Filtered out non-Argentine URL: %s", part)
					}
				}
			}
		}
	}

	return urls
}

// isArgentinianDomain checks if a URL belongs to an Argentine domain
func isArgentinianDomain(urlStr string) bool {
	argentinianDomains := []string{
		".com.ar",
		".ar/",
	}

	for _, domain := range argentinianDomains {
		if strings.Contains(urlStr, domain) {
			return true
		}
	}
	return false
}

// extractShopName extracts a readable shop name from a URL
func extractShopName(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")

	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		domain := parts[0]
		// Remove common TLDs
		domain = strings.ReplaceAll(domain, ".com.ar", "")
		domain = strings.ReplaceAll(domain, ".com", "")
		domain = strings.ReplaceAll(domain, ".ar", "")
		return capitalize(domain)
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
