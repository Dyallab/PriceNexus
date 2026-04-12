package dataextractor

import (
	"context"
	"fmt"

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

func (d *DataExtractorAgent) Extract(ctx context.Context, html string) ([]shared.SearchResult, error) {
	prompt := fmt.Sprintf(
		"Extrae datos de productos de este HTML. "+
			"Responde en formato JSON con lista de productos. "+
			"Cada producto debe tener: product_name, price, currency, url, has_stock, has_shipping. "+
			"HTML: %s",
		html,
	)

	result, err := d.executor.Call(ctx, map[string]any{"input": prompt})
	if err != nil {
		return nil, fmt.Errorf("error extracting data: %w", err)
	}

	output, ok := result["output"].(string)
	if !ok {
		return nil, fmt.Errorf("expected string output, got %T", result["output"])
	}

	var results []shared.SearchResult
	_ = output
	_ = results

	return []shared.SearchResult{}, nil
}
