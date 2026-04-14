package dataextractor

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

type DataExtractorAgent struct {
	executor *agents.Executor
	llm      llms.Model
}

func NewDataExtractorAgent(llm llms.Model) (*DataExtractorAgent, error) {
	toolList := []tools.Tool{}

	agent := agents.NewOneShotAgent(llm, toolList)
	executor := agents.NewExecutor(agent)

	return &DataExtractorAgent{
		executor: executor,
		llm:      llm,
	}, nil
}

// cleanHTML removes scripts, styles, and other non-content elements
func cleanHTML(html string) string {
	// Remove script tags and content
	re := regexp.MustCompile(`(?i)<script[\s\S]*?</script>`)
	html = re.ReplaceAllString(html, " ")

	// Remove style tags and content
	re = regexp.MustCompile(`(?i)<style[\s\S]*?</style>`)
	html = re.ReplaceAllString(html, " ")

	// Remove HTML comments
	re = regexp.MustCompile(`<!--[\s\S]*?-->`)
	html = re.ReplaceAllString(html, " ")

	// Remove noscript tags
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

	// Remove all remaining HTML tags
	re = regexp.MustCompile(`<[^>]+>`)
	html = re.ReplaceAllString(html, " ")

	// Collapse multiple spaces/newlines
	re = regexp.MustCompile(`\s+`)
	html = re.ReplaceAllString(html, " ")

	return strings.TrimSpace(html)
}

// truncateContent limits content to a reasonable size while preserving important info
func truncateContent(html string) string {
	// Try to find main content areas
	contentPatterns := []string{
		`(?i)<main[^>]*>[\s\S]*?</main>`,
		`(?i)<div\s+class="[^"]*content[^"]*"[^>]*>[\s\S]*?</div>`,
		`(?i)<div\s+id="[^"]*main[^"]*"[^>]*>[\s\S]*?</div>`,
		`(?i)<div\s+class="[^"]*products?[^"]*"[^>]*>[\s\S]*?</div>`,
	}

	for _, pattern := range contentPatterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindString(html); match != "" {
			html = match
			break
		}
	}

	// Final size limit
	if len(html) > 200000 {
		html = html[:200000]
	}

	return html
}

func (d *DataExtractorAgent) Extract(ctx context.Context, html string) ([]shared.SearchResult, error) {
	// Preprocess: truncate large content, then clean HTML
	html = truncateContent(html)
	cleanedText := cleanHTML(html)

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

	result, err := d.executor.Call(ctx, map[string]any{"input": prompt})
	if err != nil {
		return nil, fmt.Errorf("error extracting data: %w", err)
	}

	output, ok := result["output"].(string)
	if !ok {
		return nil, fmt.Errorf("expected string output, got %T", result["output"])
	}

	// Extract JSON from potentially malformed output
	jsonStr := extractJSON(output)
	if jsonStr == "" {
		jsonStr = output
	}

	// Parse JSON response
	var response struct {
		Products []shared.SearchResult `json:"products"`
	}

	err = json.Unmarshal([]byte(jsonStr), &response)
	if err != nil {
		// Try as direct array
		var directArray []shared.SearchResult
		err2 := json.Unmarshal([]byte(jsonStr), &directArray)
		if err2 != nil {
			// Return empty rather than error
			return []shared.SearchResult{}, nil
		}
		return directArray, nil
	}

	return response.Products, nil
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
