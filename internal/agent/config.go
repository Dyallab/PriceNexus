package agent

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type LLMConfig struct {
	Orchestrator  string
	WebSearcher   string
	DataExtractor string
}

type SearchConfig struct {
	AllowedDomains  []string
	ExcludedDomains []string
	SearchEngine    string
	MaxResults      int
	DefaultCurrency string
}

func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Orchestrator:  "openrouter:xiaomi/mimo-v2-flash",
		WebSearcher:   "openrouter:nvidia/nemotron-3-super-120b-a12b:free",
		DataExtractor: "openrouter:xiaomi/mimo-v2-flash",
	}
}

func DefaultSearchConfig() SearchConfig {
	return SearchConfig{
		AllowedDomains:  []string{".com.ar", ".ar"},
		ExcludedDomains: []string{"mercadolibre.com.ar"},
		SearchEngine:    "exa",
		MaxResults:      10,
		DefaultCurrency: "ARS",
	}
}

func CreateLLMs(config LLMConfig, searchConfig SearchConfig) (orchestratorLLM, webSearcherLLM, dataExtractorLLM llms.Model, err error) {
	openRouterAPIKey := os.Getenv("OPENROUTER_API_KEY")

	if config.Orchestrator != "" {
		if config.Orchestrator == "openai" || config.Orchestrator == "openrouter" {
			orchestratorLLM = NewOpenRouterModel(openRouterAPIKey, "xiaomi/mimo-v2-flash")
		} else if strings.HasPrefix(config.Orchestrator, "openrouter:") {
			modelName := strings.TrimPrefix(config.Orchestrator, "openrouter:")
			orchestratorLLM = NewOpenRouterModel(openRouterAPIKey, modelName)
		}
	}

	if config.WebSearcher != "" {
		if config.WebSearcher == "ollama" || strings.HasPrefix(config.WebSearcher, "ollama:") {
			modelName := "phi3:mini"
			if strings.HasPrefix(config.WebSearcher, "ollama:") {
				modelName = strings.TrimPrefix(config.WebSearcher, "ollama:")
			}
			webSearcherLLM, err = ollama.New(ollama.WithModel(modelName))
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create Ollama LLM for web searcher (%s): %w", modelName, err)
			}
		} else if config.WebSearcher == "openai" || strings.HasPrefix(config.WebSearcher, "openrouter:") {
			modelName := "nvidia/nemotron-3-super-120b-a12b:free"
			if strings.HasPrefix(config.WebSearcher, "openrouter:") {
				modelName = strings.TrimPrefix(config.WebSearcher, "openrouter:")
			}
			webSearcherModel := NewOpenRouterModel(openRouterAPIKey, modelName)
			webSearcherModel.AddWebSearchTool(searchConfig)
			webSearcherLLM = webSearcherModel
		}
	}

	if config.DataExtractor != "" {
		if config.DataExtractor == "ollama" || strings.HasPrefix(config.DataExtractor, "ollama:") {
			modelName := "phi3:mini"
			if strings.HasPrefix(config.DataExtractor, "ollama:") {
				modelName = strings.TrimPrefix(config.DataExtractor, "ollama:")
			}
			dataExtractorLLM, err = ollama.New(ollama.WithModel(modelName))
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create Ollama LLM for data extractor (%s): %w", modelName, err)
			}
		} else if config.DataExtractor == "openai" || strings.HasPrefix(config.DataExtractor, "openrouter:") {
			modelName := "xiaomi/mimo-v2-flash"
			if strings.HasPrefix(config.DataExtractor, "openrouter:") {
				modelName = strings.TrimPrefix(config.DataExtractor, "openrouter:")
			}
			dataExtractorLLM = NewOpenRouterModel(openRouterAPIKey, modelName)
		}
	}

	return orchestratorLLM, webSearcherLLM, dataExtractorLLM, nil
}

func LoadFromEnv() LLMConfig {
	config := DefaultLLMConfig()

	if orch := os.Getenv("PRICE_NEXUS_ORCHESTRATOR_LLM"); orch != "" {
		config.Orchestrator = orch
	}
	if web := os.Getenv("PRICE_NEXUS_WEBSEARCHER_LLM"); web != "" {
		config.WebSearcher = web
	}
	if extract := os.Getenv("PRICE_NEXUS_DATAEXTRACTOR_LLM"); extract != "" {
		config.DataExtractor = extract
	}

	return config
}

func LoadSearchConfigFromEnv() SearchConfig {
	config := DefaultSearchConfig()

	if domains := os.Getenv("PRICE_NEXUS_WEBSEARCH_ALLOWED_DOMAINS"); domains != "" {
		parsed := parseCommaSeparated(domains)
		if len(parsed) > 0 {
			config.AllowedDomains = normalizeDomains(parsed)
		}
	}

	if rawMaxResults := os.Getenv("PRICE_NEXUS_WEBSEARCH_MAX_RESULTS"); rawMaxResults != "" {
		maxResults, err := strconv.Atoi(strings.TrimSpace(rawMaxResults))
		if err == nil && maxResults > 0 {
			config.MaxResults = maxResults
		}
	}

	if currency := strings.TrimSpace(os.Getenv("PRICE_NEXUS_DEFAULT_CURRENCY")); currency != "" {
		config.DefaultCurrency = strings.ToUpper(currency)
	}

	return config
}

func parseCommaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func normalizeDomains(domains []string) []string {
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
	return result
}
