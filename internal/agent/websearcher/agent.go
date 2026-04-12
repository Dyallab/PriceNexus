package websearcher

import (
	"context"
	"fmt"

	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

type WebSearcherAgent struct {
	executor *agents.Executor
	llm      llms.Model
	tools    []tools.Tool
}

func NewWebSearcherAgent(llm llms.Model) (*WebSearcherAgent, error) {
	toolList := []tools.Tool{}

	agent := agents.NewOneShotAgent(llm, toolList)
	executor := agents.NewExecutor(agent)

	return &WebSearcherAgent{
		executor: executor,
		llm:      llm,
		tools:    toolList,
	}, nil
}

func (w *WebSearcherAgent) Search(ctx context.Context, query string) ([]shared.SearchResult, error) {
	input := fmt.Sprintf("Search for: %s", query)
	result, err := w.executor.Call(ctx, map[string]any{"input": input})
	if err != nil {
		return nil, fmt.Errorf("error searching: %w", err)
	}

	_ = result

	return []shared.SearchResult{}, nil
}
