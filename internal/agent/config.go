package agent

import (
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type LLMConfig struct {
	Orchestrator  string
	WebSearcher   string
	DataExtractor string
}

func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Orchestrator:  "openrouter:nvidia/nemotron-3-super-120b-a12b:free",
		WebSearcher:   "openrouter:nvidia/nemotron-3-super-120b-a12b:free",
		DataExtractor: "openrouter:nvidia/nemotron-3-super-120b-a12b:free",
	}
}

func CreateLLMs(config LLMConfig) (orchestratorLLM, webSearcherLLM, dataExtractorLLM llms.Model, err error) {
	openRouterAPIKey := os.Getenv("OPENROUTER_API_KEY")

	if config.Orchestrator != "" {
		if config.Orchestrator == "openai" || config.Orchestrator == "openrouter" {
			orchestratorLLM = NewOpenRouterModel(openRouterAPIKey, "openai/gpt-4o-mini")
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
			modelName := "gpt-4o-mini"
			if strings.HasPrefix(config.WebSearcher, "openrouter:") {
				modelName = strings.TrimPrefix(config.WebSearcher, "openrouter:")
			}
			webSearcherModel := NewOpenRouterModel(openRouterAPIKey, modelName)
			webSearcherModel.AddWebSearchTool()
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
			modelName := "gpt-4o-mini"
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
