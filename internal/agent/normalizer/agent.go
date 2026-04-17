package normalizer

import (
	"net/url"
	"strings"

	"github.com/dyallo/pricenexus/internal/agent/shared"
)

type SourceContext struct {
	Query     string
	Source    shared.SearchResult
	SourceURL string
}

type Agent struct {
	defaultCurrency string
}

func NewAgent(defaultCurrency string) *Agent {
	if strings.TrimSpace(defaultCurrency) == "" {
		defaultCurrency = "ARS"
	}

	return &Agent{defaultCurrency: strings.ToUpper(strings.TrimSpace(defaultCurrency))}
}

func (a *Agent) Normalize(results []shared.SearchResult, ctx SourceContext) []shared.SearchResult {
	normalized := make([]shared.SearchResult, 0, len(results))
	for _, result := range results {
		item := result

		item.SearchTerm = normalizeWhitespace(firstNonEmpty(item.SearchTerm, ctx.Query))

		item.ProductName = normalizeWhitespace(item.ProductName)
		if isGenericProductName(item.ProductName) {
			item.ProductName = normalizeWhitespace(firstNonEmpty(ctx.Source.ProductName, ctx.Query))
		}
		if item.ProductName == "" {
			item.ProductName = item.SearchTerm
		}

		item.URL = normalizeWhitespace(firstNonEmpty(item.URL, ctx.SourceURL, ctx.Source.URL))

		item.ShopName = normalizeWhitespace(firstNonEmpty(item.ShopName, ctx.Source.ShopName))
		if isGenericShopName(item.ShopName) {
			item.ShopName = deriveShopName(item.URL)
		}

		item.Currency = strings.ToUpper(normalizeWhitespace(item.Currency))
		if item.Currency == "" {
			item.Currency = a.defaultCurrency
		}

		normalized = append(normalized, item)
	}

	return normalized
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func isGenericProductName(value string) bool {
	switch strings.ToLower(normalizeWhitespace(value)) {
	case "", "product", "producto", "item", "unknown":
		return true
	default:
		return false
	}
}

func isGenericShopName(value string) bool {
	switch strings.ToLower(normalizeWhitespace(value)) {
	case "", "unknown":
		return true
	default:
		return false
	}
}

func deriveShopName(rawURL string) string {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}

	hostname := strings.TrimPrefix(strings.ToLower(parsedURL.Hostname()), "www.")
	hostname = strings.TrimSuffix(hostname, ".com.ar")
	hostname = strings.TrimSuffix(hostname, ".ar")
	hostname = strings.TrimSuffix(hostname, ".com")
	hostname = strings.Trim(hostname, ".")
	if hostname == "" {
		return ""
	}

	return strings.ToUpper(hostname[:1]) + hostname[1:]
}
