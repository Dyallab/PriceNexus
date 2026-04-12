package formfinder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// FormInfo representa un formulario HTML encontrado en una página
type FormInfo struct {
	Action     string
	Method     string
	Inputs     []InputField
	IsSearch   bool
	Confidence float64 // 0.0 a 1.0, qué tan seguro es que es un buscador
}

// InputField representa un campo de entrada dentro de un formulario
type InputField struct {
	Name     string
	Type     string
	ID       string
	Class    string
	Required bool
}

// SearchResult representa una URL de búsqueda encontrada
type SearchResult struct {
	URL        string
	Confidence float64
	Source     string // "form-action", "direct-link", "query-param"
}

// FormFinderAgent descubre automáticamente formularios de búsqueda en sitios web
type FormFinderAgent struct {
	logger *logrus.Logger
	client *http.Client
}

// NewFormFinderAgent crea un nuevo agente de descubrimiento de formularios
func NewFormFinderAgent() *FormFinderAgent {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &FormFinderAgent{
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DiscoverSearchForms analiza una página web y descubre formularios de búsqueda
func (f *FormFinderAgent) DiscoverSearchForms(ctx context.Context, pageURL string) ([]FormInfo, error) {
	f.logger.Infof("Analyzing page for search forms: %s", pageURL)

	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Use realistic browser headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	var forms []FormInfo

	// Find all forms on the page
	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		form := FormInfo{
			Inputs: []InputField{},
		}

		// Get form action
		action, exists := s.Attr("action")
		if exists {
			form.Action = action
		} else {
			// If no action, use the current page URL
			form.Action = pageURL
		}

		// Get form method (default to GET)
		method, exists := s.Attr("method")
		if exists {
			form.Method = strings.ToLower(method)
		} else {
			form.Method = "get"
		}

		// Analyze input fields
		s.Find("input").Each(func(j int, input *goquery.Selection) {
			inputField := InputField{}

			if name, exists := input.Attr("name"); exists {
				inputField.Name = name
			}
			if inputType, exists := input.Attr("type"); exists {
				inputField.Type = inputType
			}
			if id, exists := input.Attr("id"); exists {
				inputField.ID = id
			}
			if class, exists := input.Attr("class"); exists {
				inputField.Class = class
			}
			if _, exists := input.Attr("required"); exists {
				inputField.Required = true
			}

			form.Inputs = append(form.Inputs, inputField)
		})

		// Check if this form is likely a search form
		form.IsSearch, form.Confidence = f.isSearchForm(form)

		if form.IsSearch {
			forms = append(forms, form)
		}
	})

	f.logger.Infof("Found %d search form(s) on %s", len(forms), pageURL)
	return forms, nil
}

// isSearchForm determines if a form is likely a search form
func (f *FormFinderAgent) isSearchForm(form FormInfo) (bool, float64) {
	searchKeywords := []string{"q", "search", "query", "keyword", "s", "busqueda", "buscar", "p"}
	confidence := 0.0

	// Check input names
	for _, input := range form.Inputs {
		for _, keyword := range searchKeywords {
			if strings.Contains(strings.ToLower(input.Name), keyword) {
				confidence += 0.3
				break
			}
		}
	}

	// Check if form method is GET (search forms usually use GET)
	if form.Method == "get" {
		confidence += 0.2
	}

	// Check if action URL contains search-related terms
	if strings.Contains(strings.ToLower(form.Action), "search") ||
		strings.Contains(strings.ToLower(form.Action), "buscar") {
		confidence += 0.3
	}

	// Also check form ID or class
	if strings.Contains(strings.ToLower(form.Action), "javascript") {
		confidence -= 0.5
	}

	return confidence > 0.3, confidence
}

// GenerateSearchURL generates a search URL from a form and query
func (f *FormFinderAgent) GenerateSearchURL(form FormInfo, query string) (string, error) {
	// Find the query input field
	var queryInput *InputField
	searchKeywords := []string{"q", "search", "query", "keyword", "s", "busqueda", "buscar", "p"}

	for _, input := range form.Inputs {
		if input.Name != "" {
			for _, keyword := range searchKeywords {
				if strings.Contains(strings.ToLower(input.Name), keyword) {
					queryInput = &input
					break
				}
			}
		}
		if queryInput != nil {
			break
		}
	}

	// If no query input found, try the first text input as fallback
	if queryInput == nil {
		for _, input := range form.Inputs {
			if input.Type == "text" || input.Type == "search" {
				queryInput = &input
				break
			}
		}
	}

	if queryInput == nil {
		return "", fmt.Errorf("no query input field found in form")
	}

	// Build the URL based on the form method
	if form.Method == "get" {
		// For GET forms, append query parameters
		parsedAction, err := url.Parse(form.Action)
		if err != nil {
			return "", fmt.Errorf("error parsing form action: %w", err)
		}

		q := parsedAction.Query()
		q.Set(queryInput.Name, query)
		parsedAction.RawQuery = q.Encode()

		return parsedAction.String(), nil
	}

	// For POST forms, we can't generate a direct URL
	// We would need to submit the form
	return "", fmt.Errorf("POST forms not supported for URL generation")
}

// DiscoverFromBaseURL tries to discover search functionality from a base URL
func (f *FormFinderAgent) DiscoverFromBaseURL(ctx context.Context, baseURL string) ([]SearchResult, error) {
	f.logger.Infof("Discovering search from base URL: %s", baseURL)

	var results []SearchResult

	// Try common search paths
	commonPaths := []string{
		"/",
		"/search",
		"/buscar",
		"/catalogsearch/result/",
	}

	for _, path := range commonPaths {
		testURL := baseURL + path
		forms, err := f.DiscoverSearchForms(ctx, testURL)
		if err != nil {
			f.logger.Debugf("Error analyzing %s: %v", testURL, err)
			continue
		}

		for _, form := range forms {
			if form.IsSearch {
				// Generate a test search URL
				testQuery := "test"
				searchURL, err := f.GenerateSearchURL(form, testQuery)
				if err == nil {
					results = append(results, SearchResult{
						URL:        searchURL,
						Confidence: form.Confidence,
						Source:     "form-action",
					})
				}
			}
		}
	}

	return results, nil
}
