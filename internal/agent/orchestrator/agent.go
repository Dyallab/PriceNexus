package orchestrator

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

type OrchestratorAgent struct {
	executor *agents.Executor
	llm      llms.Model
}

func NewOrchestratorAgent(llm llms.Model, toolList []tools.Tool) (*OrchestratorAgent, error) {
	agent := agents.NewOneShotAgent(llm, toolList)
	executor := agents.NewExecutor(agent)

	return &OrchestratorAgent{
		executor: executor,
		llm:      llm,
	}, nil
}

func (oa *OrchestratorAgent) Execute(ctx context.Context, input string) (string, error) {
	result, err := oa.executor.Call(ctx, map[string]any{"input": input})
	if err != nil {
		return "", fmt.Errorf("error executing agent: %w", err)
	}

	output, ok := result["output"].(string)
	if !ok {
		return "", fmt.Errorf("expected string output, got %T", result["output"])
	}

	return output, nil
}

func (oa *OrchestratorAgent) GetTools() []tools.Tool {
	return oa.executor.Agent.GetTools()
}
