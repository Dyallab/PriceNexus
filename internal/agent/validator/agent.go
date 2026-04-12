package validator

import (
	"context"
	"fmt"

	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

type ValidatorAgent struct {
	executor *agents.Executor
	llm      llms.Model
}

func NewValidatorAgent(llm llms.Model) (*ValidatorAgent, error) {
	toolList := []tools.Tool{}

	agent := agents.NewOneShotAgent(llm, toolList)
	executor := agents.NewExecutor(agent)

	return &ValidatorAgent{
		executor: executor,
		llm:      llm,
	}, nil
}

func (v *ValidatorAgent) Validate(ctx context.Context, results []shared.SearchResult) ([]shared.SearchResult, error) {
	prompt := fmt.Sprintf(
		"Valida estos resultados de productos. "+
			"Filtra resultados inválidos (precios negativos, URLs inválidas, etc.). "+
			"Responde solo con los resultados válidos en formato JSON. "+
			"Resultados: %+v",
		results,
	)

	result, err := v.executor.Call(ctx, map[string]any{"input": prompt})
	if err != nil {
		return nil, fmt.Errorf("error validating data: %w", err)
	}

	output, ok := result["output"].(string)
	if !ok {
		return nil, fmt.Errorf("expected string output, got %T", result["output"])
	}

	_ = output

	return results, nil
}
