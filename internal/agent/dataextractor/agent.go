package dataextractor

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	agentruntime "github.com/dyallo/pricenexus/internal/agent"
	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

// DataExtractorAgent extracts product data from HTML pages
// using a multi-strategy approach:
// 1. First tries structured data (meta tags, JSON-LD, data attributes)
// 2. If that fails, uses LLM to find content and extract
// 3. Falls back to gentle HTML cleaning with preserved data attributes
type DataExtractorAgent struct {
	executor *agents.Executor
	llm      llms.Model
}

type htmlFragmentResponse struct {
	HTMLFragment string `json:"html_fragment"`
}

type productExtractionResponse struct {
	Products []shared.SearchResult `json:"products"`
}

// NewDataExtractorAgent creates a new data extractor agent
func NewDataExtractorAgent(llm llms.Model) (*DataExtractorAgent, error) {
	toolList := []tools.Tool{}

	agent := agents.NewOneShotAgent(llm, toolList)
	executor := agents.NewExecutor(agent)

	return &DataExtractorAgent{
		executor: executor,
		llm:      llm,
	}, nil
}

func htmlFragmentSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"html_fragment": map[string]any{
				"type": "string",
			},
		},
		"required": []string{"html_fragment"},
	}
}

func productExtractionSchema() map[string]any {
	productProperties := map[string]any{
		"product_name": map[string]any{"type": "string"},
		"price":        map[string]any{"type": "number"},
		"currency":     map[string]any{"type": "string"},
		"url":          map[string]any{"type": "string"},
		"has_stock":    map[string]any{"type": "boolean"},
		"has_shipping": map[string]any{"type": "boolean"},
	}

	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"products": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties":           productProperties,
					"required": []string{
						"product_name",
						"price",
						"currency",
						"url",
						"has_stock",
						"has_shipping",
					},
				},
			},
		},
		"required": []string{"products"},
	}
}

func (d *DataExtractorAgent) callLLM(ctx context.Context, prompt string) (string, error) {
	if openRouterModel, ok := d.llm.(*agentruntime.OpenRouterModel); ok {
		return openRouterModel.Call(ctx, prompt)
	}

	result, err := d.executor.Call(ctx, map[string]any{"input": prompt})
	if err != nil {
		return "", err
	}

	output, ok := result["output"].(string)
	if !ok {
		return "", fmt.Errorf("expected string output, got %T", result["output"])
	}

	return output, nil
}

func (d *DataExtractorAgent) callLLMWithJSONSchema(
	ctx context.Context,
	prompt string,
	schemaName string,
	schema map[string]any,
) (string, error) {
	if openRouterModel, ok := d.llm.(*agentruntime.OpenRouterModel); ok {
		return openRouterModel.CallWithJSONSchema(ctx, prompt, schemaName, schema)
	}

	return d.callLLM(ctx, prompt)
}

func parseHTMLFragmentOutput(output string) (string, error) {
	jsonStr := extractJSON(output)
	if jsonStr == "" {
		jsonStr = output
	}

	var response htmlFragmentResponse
	if err := json.Unmarshal([]byte(jsonStr), &response); err == nil {
		if strings.TrimSpace(response.HTMLFragment) == "" {
			return "", fmt.Errorf(
				"structured response did not include html_fragment; output=%q",
				outputPreview(output),
			)
		}

		return normalizeHTMLFragment(response.HTMLFragment), nil
	}

	if fragment, ok := extractJSONStringFieldRelaxed(jsonStr, "html_fragment"); ok {
		if strings.TrimSpace(fragment) != "" {
			return normalizeHTMLFragment(fragment), nil
		}
	}

	return "", fmt.Errorf(
		"failed to parse html fragment response; output=%q",
		outputPreview(output),
	)
}

func parseProductsOutput(output string) ([]shared.SearchResult, error) {
	return parseProductsOutputWithRetry(output, 2)
}

func parseProductsOutputWithRetry(output string, maxRetries int) ([]shared.SearchResult, error) {
	jsonStr := extractJSON(output)
	if jsonStr == "" {
		jsonStr = output
	}

	candidates := []string{jsonStr}
	repaired := repairProductsJSON(jsonStr)
	if repaired != "" && repaired != jsonStr {
		candidates = append(candidates, repaired)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries && attempt < len(candidates); attempt++ {
		results, err := tryParseProductsOutput(candidates[attempt])
		if err == nil {
			return results, nil
		}

		if !isTruncationError(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, lastErr
}

func tryParseProductsOutput(jsonStr string) ([]shared.SearchResult, error) {
	var response productExtractionResponse
	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		directArray := []shared.SearchResult{}
		if err2 := json.Unmarshal([]byte(jsonStr), &directArray); err2 != nil {
			return nil, fmt.Errorf(
				"unable to parse structured output: %w",
				err,
			)
		}

		if len(directArray) == 0 {
			return nil, fmt.Errorf(
				"structured extraction returned zero products",
			)
		}

		return directArray, nil
	}

	if len(response.Products) == 0 {
		return nil, fmt.Errorf(
			"structured extraction returned zero products",
		)
	}

	return response.Products, nil
}

func isTruncationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	truncationIndicators := []string{
		"unexpected end",
		"unexpected EOF",
		"unclosed string",
		"after object",
		"after array",
		"after closing",
	}
	for _, indicator := range truncationIndicators {
		if strings.Contains(errStr, indicator) {
			return true
		}
	}
	return false
}

func trimTrailingGarbage(jsonStr string) string {
	trimmed := strings.TrimSpace(jsonStr)
	if trimmed == "" {
		return jsonStr
	}

	startChar := trimmed[0]
	if startChar != '{' && startChar != '[' {
		return jsonStr
	}

	depth := 0
	inString := false
	escapeNext := false
	openBrackets := make([]byte, 0, 10)

	for i := 0; i < len(trimmed); i++ {
		char := trimmed[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		if char == '{' || char == '[' {
			depth++
			openBrackets = append(openBrackets, char)
		} else if char == '}' || char == ']' {
			if depth > 0 {
				depth--
				if len(openBrackets) > 0 {
					openBrackets = openBrackets[:len(openBrackets)-1]
				}
			}
		}
	}

	if depth > 0 && len(openBrackets) > 0 {
		candidate := trimmed
		for len(openBrackets) > 0 {
			open := openBrackets[len(openBrackets)-1]
			openBrackets = openBrackets[:len(openBrackets)-1]
			if open == '{' {
				candidate += "}"
			} else {
				candidate += "]"
			}
		}
		var testResp productExtractionResponse
		if json.Unmarshal([]byte(candidate), &testResp) == nil {
			return candidate
		}
		var testArr []shared.SearchResult
		if json.Unmarshal([]byte(candidate), &testArr) == nil {
			return candidate
		}
	}

	return trimmed
}

func repairProductsJSON(jsonStr string) string {
	trimmed := strings.TrimSpace(jsonStr)
	if trimmed == "" {
		return ""
	}

	sanitized := sanitizeInvalidJSONEscapes(trimmed)
	reconstructed := reconstructProductsResponse(sanitized)
	if reconstructed != "" {
		return reconstructed
	}

	return trimTrailingGarbage(sanitized)
}

func sanitizeInvalidJSONEscapes(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}

	var builder strings.Builder
	builder.Grow(len(s))

	inString := false
	escapeNext := false

	for i := 0; i < len(s); i++ {
		char := s[i]

		if escapeNext {
			escapeNext = false
			switch char {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				builder.WriteByte('\\')
				builder.WriteByte(char)
			case 'u':
				if i+4 < len(s) && isHexSequence(s[i+1:i+5]) {
					builder.WriteString(`\u`)
					builder.WriteString(s[i+1 : i+5])
					i += 4
				} else {
					builder.WriteString(`\\u`)
				}
			default:
				builder.WriteByte(char)
			}
			continue
		}

		if char == '"' {
			inString = !inString
			builder.WriteByte(char)
			continue
		}

		if inString && char == '\\' {
			escapeNext = true
			continue
		}

		builder.WriteByte(char)
	}

	return builder.String()
}

func isHexSequence(s string) bool {
	if len(s) != 4 {
		return false
	}

	for _, char := range s {
		if (char < '0' || char > '9') &&
			(char < 'a' || char > 'f') &&
			(char < 'A' || char > 'F') {
			return false
		}
	}

	return true
}

func reconstructProductsResponse(jsonStr string) string {
	productsKeyIdx := strings.Index(jsonStr, `"products"`)
	if productsKeyIdx == -1 {
		return ""
	}

	arrayStartIdx := strings.Index(jsonStr[productsKeyIdx:], "[")
	if arrayStartIdx == -1 {
		return ""
	}
	arrayStartIdx += productsKeyIdx

	productObjects := extractCompleteJSONObjectItems(jsonStr[arrayStartIdx+1:])
	if len(productObjects) == 0 {
		return ""
	}

	return `{"products":[` + strings.Join(productObjects, ",") + `]}`
}

func extractCompleteJSONObjectItems(s string) []string {
	items := []string{}
	itemStart := -1
	depth := 0
	inString := false
	escapeNext := false

	for i := 0; i < len(s); i++ {
		char := s[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch char {
		case '{':
			if depth == 0 {
				itemStart = i
			}
			depth++
		case '}':
			if depth == 0 {
				continue
			}
			depth--
			if depth == 0 && itemStart >= 0 {
				candidate := s[itemStart : i+1]
				if isValidProductObject(candidate) {
					items = append(items, candidate)
				}
				itemStart = -1
			}
		case ']':
			if depth == 0 {
				return items
			}
		}
	}

	return items
}

func isValidProductObject(candidate string) bool {
	var result shared.SearchResult
	if err := json.Unmarshal([]byte(candidate), &result); err != nil {
		return false
	}

	return strings.TrimSpace(result.ProductName) != "" && result.Price > 0
}

func extractJSONStringFieldRelaxed(input string, field string) (string, bool) {
	keyIdx := strings.Index(input, `"`+field+`"`)
	if keyIdx == -1 {
		return "", false
	}

	colonIdx := strings.Index(input[keyIdx+len(field)+2:], ":")
	if colonIdx == -1 {
		return "", false
	}
	colonIdx += keyIdx + len(field) + 2

	valueStart := colonIdx + 1
	for valueStart < len(input) && (input[valueStart] == ' ' || input[valueStart] == '\n' || input[valueStart] == '\t' || input[valueStart] == '\r') {
		valueStart++
	}
	if valueStart >= len(input) || input[valueStart] != '"' {
		return "", false
	}

	var builder strings.Builder
	inEscape := false
	for i := valueStart + 1; i < len(input); i++ {
		char := input[i]

		if inEscape {
			inEscape = false
			switch char {
			case '"', '\\', '/':
				builder.WriteByte(char)
			case 'b':
				builder.WriteByte('\b')
			case 'f':
				builder.WriteByte('\f')
			case 'n':
				builder.WriteByte('\n')
			case 'r':
				builder.WriteByte('\r')
			case 't':
				builder.WriteByte('\t')
			case 'u':
				if i+4 < len(input) && isHexSequence(input[i+1:i+5]) {
					r, _ := decodeUnicodeEscape(input[i+1 : i+5])
					builder.WriteRune(r)
					i += 4
				} else {
					builder.WriteByte('u')
				}
			default:
				builder.WriteByte(char)
			}
			continue
		}

		if char == '\\' {
			inEscape = true
			continue
		}

		if char == '"' {
			return builder.String(), true
		}

		builder.WriteByte(char)
	}

	return strings.TrimSpace(builder.String()), builder.Len() > 0
}

func decodeUnicodeEscape(hexDigits string) (rune, bool) {
	value, err := strconv.ParseUint(hexDigits, 16, 64)
	if err != nil {
		return utf8.RuneError, false
	}

	r := rune(value)
	if utf16.IsSurrogate(r) {
		return utf8.RuneError, false
	}

	return r, true
}

func outputPreview(output string) string {
	trimmed := strings.TrimSpace(output)
	if len(trimmed) <= 200 {
		return trimmed
	}

	return trimmed[:200] + "..."
}

func normalizeHTMLFragment(fragment string) string {
	return strings.NewReplacer(
		`\\"`, `"`,
		`\"`, `"`,
		`\\'`, `'`,
		`\'`, `'`,
		`\\&`, `&`,
		`\&`, `&`,
	).Replace(fragment)
}

// Extract extracts product data from HTML using multiple strategies
func (d *DataExtractorAgent) Extract(ctx context.Context, html string) ([]shared.SearchResult, error) {
	// Strategy 1: Try to extract structured data first (meta tags, JSON-LD, data attributes)
	results := d.extractStructuredData(html)
	if len(results) > 0 {
		return results, nil
	}

	// Strategy 2: Use LLM to find content structure and extract
	results, llmErr := d.extractWithLLMContentFinder(ctx, html)
	if llmErr == nil && len(results) > 0 {
		return results, nil
	}

	// Strategy 3: Fall back to gentle HTML cleaning and basic extraction
	results, fallbackErr := d.extractWithGentleCleaning(ctx, html)
	if fallbackErr == nil && len(results) > 0 {
		return results, nil
	}

	if fallbackErr != nil {
		if llmErr != nil {
			return nil, fmt.Errorf(
				"llm extraction failed: content finder: %v; gentle cleaning: %w",
				llmErr,
				fallbackErr,
			)
		}

		return nil, fmt.Errorf("llm extraction failed after gentle cleaning: %w", fallbackErr)
	}

	if llmErr != nil {
		return nil, fmt.Errorf("no products extracted after llm fallback: %w", llmErr)
	}

	return nil, fmt.Errorf("no products extracted from html")
}

// =============================================================================
// STRATEGY 1: Structured Data Extraction
// =============================================================================

// extractStructuredData tries to extract product data from meta tags, JSON-LD, and data attributes
func (d *DataExtractorAgent) extractStructuredData(html string) []shared.SearchResult {
	// Try JSON-LD first
	if results := d.extractJSONLD(html); len(results) > 0 {
		return results
	}

	// Try meta tags
	if results := d.extractMetaTags(html); len(results) > 0 {
		return results
	}

	// Try data attributes (e.g., data-variants, data-product)
	if results := d.extractDataAttributes(html); len(results) > 0 {
		return results
	}

	return nil
}

// extractJSONLD extracts product data from JSON-LD structured data
func (d *DataExtractorAgent) extractJSONLD(html string) []shared.SearchResult {
	// Find all JSON-LD script tags
	re := regexp.MustCompile(`(?i)<script[^>]*type=["']application/ld\+json["'][^>]*>[\s\S]*?</script>`)
	matches := re.FindAllString(html, -1)

	var results []shared.SearchResult

	for _, match := range matches {
		// Extract content between script tags
		contentStart := strings.Index(match, ">")
		contentEnd := strings.LastIndex(match, "<")
		if contentStart == -1 || contentEnd == -1 || contentEnd <= contentStart {
			continue
		}
		content := match[contentStart+1 : contentEnd]

		var payload any
		if err := json.Unmarshal([]byte(content), &payload); err != nil {
			continue
		}

		results = append(results, d.extractJSONLDResults(payload)...)
	}

	return results
}

func (d *DataExtractorAgent) extractJSONLDResults(node any) []shared.SearchResult {
	switch value := node.(type) {
	case []any:
		results := []shared.SearchResult{}
		for _, item := range value {
			results = append(results, d.extractJSONLDResults(item)...)
		}
		return results
	case map[string]any:
		results := []shared.SearchResult{}

		if d.isJSONLDProduct(value) {
			if result, ok := d.jsonLDMapToResult(value); ok {
				results = append(results, result)
			}
		}

		for _, nested := range value {
			switch nested.(type) {
			case map[string]any, []any:
				results = append(results, d.extractJSONLDResults(nested)...)
			}
		}

		return results
	default:
		return nil
	}
}

func (d *DataExtractorAgent) isJSONLDProduct(node map[string]any) bool {
	typeValue, ok := node["@type"]
	if !ok {
		return false
	}

	switch value := typeValue.(type) {
	case string:
		return value == "Product" || value == "IndividualProduct"
	case []any:
		for _, item := range value {
			if itemStr, ok := item.(string); ok && (itemStr == "Product" || itemStr == "IndividualProduct") {
				return true
			}
		}
	}

	return false
}

func (d *DataExtractorAgent) jsonLDMapToResult(node map[string]any) (shared.SearchResult, bool) {
	productName, _ := node["name"].(string)
	productName = strings.TrimSpace(productName)
	if productName == "" {
		return shared.SearchResult{}, false
	}

	price, currency, hasStock := d.extractJSONLDPrice(node)
	if price <= 0 {
		return shared.SearchResult{}, false
	}
	if currency == "" {
		currency = "ARS"
	}

	return shared.SearchResult{
		ProductName: productName,
		Price:       price,
		Currency:    currency,
		URL:         "",
		HasStock:    hasStock,
		HasShipping: false,
	}, true
}

func (d *DataExtractorAgent) extractJSONLDPrice(node map[string]any) (float64, string, bool) {
	if offers, ok := node["offers"]; ok {
		switch value := offers.(type) {
		case map[string]any:
			price, currency := d.extractJSONLDPriceFields(value)
			return price, currency, d.extractJSONLDAvailability(value)
		case []any:
			for _, item := range value {
				offer, ok := item.(map[string]any)
				if !ok {
					continue
				}
				price, currency := d.extractJSONLDPriceFields(offer)
				if price > 0 {
					return price, currency, d.extractJSONLDAvailability(offer)
				}
			}
		}
	}

	price, currency := d.extractJSONLDPriceFields(node)
	return price, currency, d.extractJSONLDAvailability(node)
}

func (d *DataExtractorAgent) extractJSONLDPriceFields(node map[string]any) (float64, string) {
	price := d.parseAnyPrice(node["price"])
	currency, _ := node["priceCurrency"].(string)
	return price, strings.TrimSpace(currency)
}

func (d *DataExtractorAgent) extractJSONLDAvailability(node map[string]any) bool {
	availability, _ := node["availability"].(string)
	availability = strings.ToLower(strings.TrimSpace(availability))
	if availability == "" {
		return true
	}

	return strings.Contains(availability, "instock") || strings.Contains(availability, "available")
}

func (d *DataExtractorAgent) parseAnyPrice(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		parsed, err := v.Float64()
		if err == nil {
			return parsed
		}
	case string:
		return d.parsePriceString(v)
	}

	return 0
}

// extractMetaTags extracts product data from HTML meta tags
func (d *DataExtractorAgent) extractMetaTags(html string) []shared.SearchResult {
	var results []shared.SearchResult

	// Extract og:title
	productName := d.extractMetaContent(html, `(?i)<meta[^>]*property=["']og:title["'][^>]*content=["']([^"']+)["']`)
	if productName == "" {
		productName = d.extractMetaContent(html, `(?i)<meta[^>]*content=["']([^"']+)["'][^>]*property=["']og:title["']`)
	}

	// Extract price from various meta tags
	price := d.extractPriceFromMeta(html)

	// Extract twitter:data1 which often contains price
	if price == 0 {
		twitterPrice := d.extractMetaContent(html, `(?i)<meta[^>]*name=["']twitter:data1["'][^>]*content=["']([^"']+)["']`)
		if twitterPrice == "" {
			twitterPrice = d.extractMetaContent(html, `(?i)<meta[^>]*content=["']([^"']+)["'][^>]*name=["']twitter:data1["']`)
		}
		price = d.parsePriceString(twitterPrice)
	}

	// Extract stock from meta tags
	hasStock := d.extractStockFromMeta(html)

	// Extract from tiendanube specific tags
	if price == 0 {
		tiendanubePrice := d.extractMetaContent(html, `(?i)<meta[^>]*property=["']tiendanube:price["'][^>]*content=["']([^"']+)["']`)
		if tiendanubePrice == "" {
			tiendanubePrice = d.extractMetaContent(html, `(?i)<meta[^>]*content=["']([^"']+)["'][^>]*property=["']tiendanube:price["']`)
		}
		if tiendanubePrice != "" {
			price = d.parsePriceString(tiendanubePrice)
		}
	}

	if price > 0 && productName != "" {
		results = append(results, shared.SearchResult{
			ProductName: productName,
			Price:       price,
			Currency:    "ARS",
			URL:         "",
			HasStock:    hasStock,
			HasShipping: d.extractShippingFromMeta(html),
		})
	}

	return results
}

// extractMetaContent extracts content attribute value from a meta tag matching the pattern
func (d *DataExtractorAgent) extractMetaContent(html, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractPriceFromMeta tries to find price in various meta tag formats
func (d *DataExtractorAgent) extractPriceFromMeta(html string) float64 {
	// Try og:price:amount
	priceStr := d.extractMetaContent(html, `(?i)<meta[^>]*property=["']og:price:amount["'][^>]*content=["']([^"']+)["']`)
	if priceStr == "" {
		priceStr = d.extractMetaContent(html, `(?i)<meta[^>]*content=["']([^"']+)["'][^>]*property=["']og:price:amount["']`)
	}
	if priceStr != "" {
		return d.parsePriceString(priceStr)
	}

	// Try product:price:amount
	priceStr = d.extractMetaContent(html, `(?i)<meta[^>]*property=["']product:price:amount["'][^>]*content=["']([^"']+)["']`)
	if priceStr == "" {
		priceStr = d.extractMetaContent(html, `(?i)<meta[^>]*content=["']([^"']+)["'][^>]*property=["']product:price:amount["']`)
	}
	if priceStr != "" {
		return d.parsePriceString(priceStr)
	}

	return 0
}

// extractStockFromMeta checks meta tags for stock information
func (d *DataExtractorAgent) extractStockFromMeta(html string) bool {
	// Check tiendanube:stock
	stockStr := d.extractMetaContent(html, `(?i)<meta[^>]*property=["']tiendanube:stock["'][^>]*content=["']([^"']+)["']`)
	if stockStr == "" {
		stockStr = d.extractMetaContent(html, `(?i)<meta[^>]*content=["']([^"']+)["'][^>]*property=["']tiendanube:stock["']`)
	}
	if stockStr != "" {
		// Stock > 0 means available
		stock, err := strconv.Atoi(stockStr)
		if err == nil && stock > 0 {
			return true
		}
	}

	// Check twitter:data2 (sometimes used for stock)
	twitterStock := d.extractMetaContent(html, `(?i)<meta[^>]*name=["']twitter:data2["'][^>]*content=["']([^"']+)["']`)
	if twitterStock != "" && !strings.Contains(strings.ToLower(twitterStock), "agotado") && !strings.Contains(strings.ToLower(twitterStock), "sin stock") {
		return true
	}

	return false
}

// extractShippingFromMeta checks meta tags for shipping information
func (d *DataExtractorAgent) extractShippingFromMeta(html string) bool {
	// Check for shipping-related meta tags
	shipping := d.extractMetaContent(html, `(?i)<meta[^>]*property=["']product:shipping[@\w]+["'][^>]*content=["']([^"']+)["']`)
	if strings.ToLower(shipping) == "true" || shipping == "1" {
		return true
	}
	return false
}

// extractDataAttributes extracts product data from data-* attributes
func (d *DataExtractorAgent) extractDataAttributes(html string) []shared.SearchResult {
	var results []shared.SearchResult

	// Find data-variants which is commonly used in Argentine e-commerce (Tiendanube, etc.)
	variantsRe := regexp.MustCompile(`(?i)<[^>]+\s+data-variants=["']([^"']+)["'][^>]*>`)
	matches := variantsRe.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		// Decode HTML entities in the attribute value
		variantsJSON := d.decodeHTMLEntities(match[1])

		// Try to parse as array of variants
		var variants []struct {
			ProductID      int     `json:"product_id"`
			PriceShort     string  `json:"price_short"`
			PriceLong      string  `json:"price_long"`
			PriceNumber    float64 `json:"price_number"`
			PriceNumberRaw int64   `json:"price_number_raw"`
			Stock          int     `json:"stock"`
			Available      bool    `json:"available"`
		}

		if err := json.Unmarshal([]byte(variantsJSON), &variants); err != nil {
			// Try single object format
			var singleVariant struct {
				ProductID   int     `json:"product_id"`
				PriceShort  string  `json:"price_short"`
				PriceNumber float64 `json:"price_number"`
				Stock       int     `json:"stock"`
				Available   bool    `json:"available"`
			}
			if err2 := json.Unmarshal([]byte(variantsJSON), &singleVariant); err2 == nil {
				if singleVariant.PriceNumber > 0 {
					results = append(results, shared.SearchResult{
						ProductName: "",
						Price:       singleVariant.PriceNumber,
						Currency:    "ARS",
						URL:         "",
						HasStock:    singleVariant.Available && singleVariant.Stock > 0,
						HasShipping: true, // Default, will be enhanced by validator
					})
				}
			}
			continue
		}

		for _, variant := range variants {
			if variant.PriceNumber > 0 || variant.PriceNumberRaw > 0 {
				price := variant.PriceNumber
				if price == 0 && variant.PriceNumberRaw > 0 {
					price = float64(variant.PriceNumberRaw) / 100 // Often stored as cents
				}
				results = append(results, shared.SearchResult{
					ProductName: "",
					Price:       price,
					Currency:    "ARS",
					URL:         "",
					HasStock:    variant.Available && variant.Stock > 0,
					HasShipping: true,
				})
			}
		}
	}

	// Try data-product or data-price attributes
	productRe := regexp.MustCompile(`(?i)<[^>]+\s+data-product=["']([^"']+)["'][^>]*>`)
	productMatches := productRe.FindAllStringSubmatch(html, -1)
	for _, match := range productMatches {
		if len(match) < 2 {
			continue
		}
		productJSON := d.decodeHTMLEntities(match[1])
		var product struct {
			Name  string  `json:"name"`
			Price float64 `json:"price"`
			Stock int     `json:"stock"`
		}
		if err := json.Unmarshal([]byte(productJSON), &product); err == nil && product.Price > 0 {
			hasStock := product.Stock > 0
			results = append(results, shared.SearchResult{
				ProductName: product.Name,
				Price:       product.Price,
				Currency:    "ARS",
				URL:         "",
				HasStock:    hasStock,
				HasShipping: true,
			})
		}
	}

	return results
}

// =============================================================================
// STRATEGY 2: LLM Content Finder
// =============================================================================

// extractWithLLMContentFinder uses the LLM to analyze HTML structure and find content
func (d *DataExtractorAgent) extractWithLLMContentFinder(ctx context.Context, html string) ([]shared.SearchResult, error) {
	// First, ask the LLM to identify where the product information is in the HTML
	prompt := `Analiza el siguiente HTML y extrae SOLO la información del producto principal.

INSTRUCCIONES:
1. Busca en el HTML la estructura que contiene información del producto (nombre, precio, stock)
2. Devuelve SOLO el fragmento HTML relevante que contiene los datos del producto
3. NO devuelvas todo el HTML, solo la porción que tiene los datos del producto
4. Preservea los atributos data-* ya que pueden contener JSON con información importante
5. Si encuentras múltiples productos, concentra en el primero/más importante

RESPUESTA OBLIGATORIA (solo esto, sin texto adicional):
{"html_fragment": "el HTML relevante aquí"}`

	maxLen := 50000
	if len(html) < maxLen {
		maxLen = len(html)
	}

	output, err := d.callLLMWithJSONSchema(
		ctx,
		prompt+"\n\n"+html[:maxLen],
		"html_fragment_response",
		htmlFragmentSchema(),
	)
	if err != nil {
		return nil, fmt.Errorf("error identifying product fragment: %w", err)
	}

	htmlFragment, err := parseHTMLFragmentOutput(output)
	if err != nil {
		return nil, err
	}

	// Now extract product data from the HTML fragment using the LLM
	return d.extractFromHTMLFragment(ctx, htmlFragment)
}

// extractFromHTMLFragment uses LLM to extract structured data from HTML fragment
func (d *DataExtractorAgent) extractFromHTMLFragment(ctx context.Context, htmlFragment string) ([]shared.SearchResult, error) {
	// Clean the fragment gently - preserve data attributes
	cleaned := d.gentleCleanHTML(htmlFragment)

	prompt := `Extrae la información del producto del siguiente contenido HTML.

INSTRUCCIONES:
1. Identifica el nombre del producto
2. Identifica el precio (solo el número, sin $ ni símbolos)
3. Identifica si hay stock disponible
4. Identifica si hay información de envío

CAMPOS A EXTRAER:
- product_name: Nombre del producto
- price: Solo el número (ej: 1599.99)
- currency: Moneda (normalmente "ARS")
- has_stock: true/false
- has_shipping: true/false

RESPUESTA OBLIGATORIA (solo JSON):
{"products": [{"product_name": "...", "price": 0.0, "currency": "ARS", "has_stock": true, "has_shipping": false}]}

Si no encuentras productos claros, responde: {"products": []}

CONTENIDO:
` + cleaned

	output, err := d.callLLMWithJSONSchema(
		ctx,
		prompt,
		"product_extraction_response",
		productExtractionSchema(),
	)
	if err != nil {
		return nil, fmt.Errorf("error extracting data from html fragment: %w", err)
	}

	return parseProductsOutput(output)
}

// =============================================================================
// STRATEGY 3: Gentle HTML Cleaning
// =============================================================================

// extractWithGentleCleaning uses gentle HTML cleaning and basic extraction
func (d *DataExtractorAgent) extractWithGentleCleaning(ctx context.Context, html string) ([]shared.SearchResult, error) {
	// Clean HTML gently - preserve data attributes
	cleanedText := d.gentleCleanHTML(html)

	// Skip if content is too minimal
	if len(cleanedText) < 100 {
		return []shared.SearchResult{}, nil
	}

	// Cap at reasonable size for LLM
	if len(cleanedText) > 12000 {
		cleanedText = cleanedText[:12000]
	}

	prompt := `Extrae TODOS los productos del siguiente contenido de tienda online argentina.

INSTRUCCIONES:
1. Identifica cada producto distinto (nombre + precio visible)
2. Extrae estos campos para cada producto:
   - product_name: El nombre/título del producto
   - price: Solo el número (sin $, sin ARS). Ej: 1599.99
   - currency: Moneda (normalmente "ARS" para tiendas argentinas)
   - url: Déjalo vacío ""
   - has_stock: true si dice "en stock"/"disponible", false si dice "agotado"/"sin stock"
   - has_shipping: true si menciona envío, false si no

REGLAS IMPORTANTES:
- Responde SOLO con JSON, sin texto antes ni después
- Si encuentras varios productos, lista todos en el array
- Si no encuentras productos con precio visible, responde: {"products": []}
- Los precios deben ser números puros: 1599.99 (no "$ 1599,99" ni "1599.99 ARS")
- Si no estás seguro del stock/envío, usa false

FORMATO OBLIGATORIO:
{"products": [{"product_name": "Nombre del Producto", "price": 1599.99, "currency": "ARS", "url": "", "has_stock": true, "has_shipping": false}]}

CONTENIDO A ANALIZAR:
` + cleanedText

	output, err := d.callLLMWithJSONSchema(
		ctx,
		prompt,
		"product_extraction_response",
		productExtractionSchema(),
	)
	if err != nil {
		return nil, fmt.Errorf("error extracting data: %w", err)
	}

	return parseProductsOutput(output)
}

// gentleCleanHTML removes dangerous/malicious content but preserves data attributes and structure
func (d *DataExtractorAgent) gentleCleanHTML(html string) string {
	// Remove script tags but preserve their content in comments for data extraction
	re := regexp.MustCompile(`(?i)<script[\s\S]*?</script>`)
	html = re.ReplaceAllString(html, " ")

	// Remove style tags completely
	re = regexp.MustCompile(`(?i)<style[\s\S]*?</style>`)
	html = re.ReplaceAllString(html, " ")

	// Remove HTML comments
	re = regexp.MustCompile(`<!--[\s\S]*?-->`)
	html = re.ReplaceAllString(html, " ")

	// Remove noscript tags but preserve content
	re = regexp.MustCompile(`(?i)<noscript[\s\S]*?</noscript>`)
	html = re.ReplaceAllString(html, " ")

	// Decode common HTML entities
	html = strings.NewReplacer(
		"&nbsp;", " ",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&#39;", "'",
		"&apos;", "'",
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
		"<p>", "\n",
		"</p>", "\n",
		"<li>", "- ",
		"</li>", "\n",
	).Replace(html)

	// Replace data-xxx="..." patterns with placeholders to preserve them
	dataAttrPattern := regexp.MustCompile(`\s+data-[a-zA-Z0-9\-]+="[^"]*"`)
	dataAttrs := dataAttrPattern.FindAllString(html, -1)

	// Remove remaining HTML tags
	re = regexp.MustCompile(`<[^>]+>`)
	html = re.ReplaceAllString(html, " ")

	// Collapse multiple spaces/newlines
	re = regexp.MustCompile(`\s+`)
	html = re.ReplaceAllString(html, " ")

	// Re-add data attributes at the end if we found any (for LLM to see)
	if len(dataAttrs) > 0 {
		html += "\n\n[Data attributes found: " + strings.Join(dataAttrs, ", ") + "]"
	}

	return strings.TrimSpace(html)
}

// =============================================================================
// Helper Functions
// =============================================================================

// extractPriceAndCurrency extracts price and currency from various formats
func (d *DataExtractorAgent) extractPriceAndCurrency(priceValue any, currency string, offer *struct {
	Price         any    `json:"price"`
	PriceCurrency string `json:"priceCurrency"`
}) (float64, string) {
	// From direct price field
	if priceValue != nil {
		switch v := priceValue.(type) {
		case float64:
			if v > 0 {
				return v, currency
			}
		case int:
			if v > 0 {
				return float64(v), currency
			}
		case string:
			price := d.parsePriceString(v)
			if price > 0 {
				return price, currency
			}
		}
	}

	// From offer
	if offer != nil {
		if offer.Price != nil {
			switch v := offer.Price.(type) {
			case float64:
				if v > 0 {
					return v, d.getCurrency(offer.PriceCurrency, currency)
				}
			case int:
				if v > 0 {
					return float64(v), d.getCurrency(offer.PriceCurrency, currency)
				}
			case string:
				price := d.parsePriceString(v)
				if price > 0 {
					return price, d.getCurrency(offer.PriceCurrency, currency)
				}
			}
		}
	}

	return 0, "ARS"
}

// getCurrency returns the currency or a default
func (d *DataExtractorAgent) getCurrency(currency, fallback string) string {
	if currency != "" {
		return currency
	}
	if fallback != "" {
		return fallback
	}
	return "ARS"
}

// parsePriceString parses a price string like "$59.000,00" or "59000" into float64
func (d *DataExtractorAgent) parsePriceString(priceStr string) float64 {
	if priceStr == "" {
		return 0
	}

	// Remove currency symbols and spaces
	priceStr = strings.ReplaceAll(priceStr, "$", "")
	priceStr = strings.ReplaceAll(priceStr, "ARS", "")
	priceStr = strings.ReplaceAll(priceStr, "USD", "")
	priceStr = strings.TrimSpace(priceStr)

	// Handle Argentine format: 59.000,00 -> 59000.00
	// First, detect if it uses dot as thousands separator and comma as decimal
	hasDotThousands := strings.Count(priceStr, ".") > 0
	hasCommaDecimal := strings.Contains(priceStr, ",")

	if hasDotThousands && hasCommaDecimal {
		// Argentine format: 59.000,00
		priceStr = strings.ReplaceAll(priceStr, ".", "")
		priceStr = strings.ReplaceAll(priceStr, ",", ".")
	} else if hasCommaDecimal && !hasDotThousands {
		// European format: 59000,00
		priceStr = strings.ReplaceAll(priceStr, ",", ".")
	} else if hasDotThousands && !hasCommaDecimal {
		// US format with thousands: 59,000.00
		// Actually this is ambiguous, assume US format
		priceStr = strings.ReplaceAll(priceStr, ",", "")
	}

	// Parse the float
	var price float64
	fmt.Sscanf(priceStr, "%f", &price)

	return price
}

// decodeHTMLEntities decodes common HTML entities
func (d *DataExtractorAgent) decodeHTMLEntities(s string) string {
	return strings.NewReplacer(
		"&quot;", "\"",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&#39;", "'",
		"&apos;", "'",
		"&nbsp;", " ",
	).Replace(s)
}

// extractJSON finds and extracts valid JSON from text with extra content
func extractJSON(s string) string {
	// Find first { or [
	start := strings.IndexAny(s, "{[")
	if start == -1 {
		return ""
	}

	// Find matching closing bracket
	depth := 0
	inString := false
	escapeNext := false
	endPos := -1

	for i := start; i < len(s); i++ {
		char := s[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		if !inString {
			if char == '{' || char == '[' {
				depth++
			} else if char == '}' || char == ']' {
				depth--
				if depth == 0 {
					endPos = i + 1
					break
				}
			}
		}
	}

	if endPos > 0 {
		return s[start:endPos]
	}

	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
