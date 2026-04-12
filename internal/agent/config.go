package agent

import (
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

type LLMConfig struct {
	Orchestrator  string
	WebSearcher   string
	DataExtractor string
}

func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Orchestrator:  "openrouter:xiaomi/mimo-v2-flash",
		WebSearcher:   "ollama:gemma4:e4b",
		DataExtractor: "ollama:gemma4:e4b",
	}
}

func CreateLLMs(config LLMConfig) (orchestratorLLM, webSearcherLLM, dataExtractorLLM llms.Model, err error) {
	openRouterAPIKey := os.Getenv("OPENROUTER_API_KEY")

	if config.Orchestrator != "" {
		if config.Orchestrator == "openai" || config.Orchestrator == "openrouter" {
			opts := []openai.Option{}
			if openRouterAPIKey != "" {
				opts = append(opts, openai.WithBaseURL("https://openrouter.ai/api/v1"))
			}
			orchestratorLLM, err = openai.New(opts...)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create OpenAI LLM: %w", err)
			}
		} else if strings.HasPrefix(config.Orchestrator, "openrouter:") {
			modelName := strings.TrimPrefix(config.Orchestrator, "openrouter:")
			opts := []openai.Option{
				openai.WithModel(modelName),
				openai.WithBaseURL("https://openrouter.ai/api/v1"),
			}
			orchestratorLLM, err = openai.New(opts...)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create OpenRouter LLM (%s): %w", modelName, err)
			}
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
			opts := []openai.Option{
				openai.WithModel(modelName),
				openai.WithBaseURL("https://openrouter.ai/api/v1"),
			}
			webSearcherLLM, err = openai.New(opts...)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create OpenRouter LLM for web searcher (%s): %w", modelName, err)
			}
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
			opts := []openai.Option{
				openai.WithModel(modelName),
				openai.WithBaseURL("https://openrouter.ai/api/v1"),
			}
			dataExtractorLLM, err = openai.New(opts...)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create OpenRouter LLM for data extractor (%s): %w", modelName, err)
			}
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
