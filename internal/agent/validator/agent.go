package validator

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/tmc/langchaingo/llms"
)

type ValidatorAgent struct{}

func NewValidatorAgent(llm llms.Model) (*ValidatorAgent, error) {
	_ = llm
	return &ValidatorAgent{}, nil
}

func (v *ValidatorAgent) Validate(ctx context.Context, results []shared.SearchResult) ([]shared.SearchResult, error) {
	_ = ctx

	validated := make([]shared.SearchResult, 0, len(results))
	for _, result := range results {
		item := result
		item.SearchTerm = strings.TrimSpace(item.SearchTerm)
		item.ProductName = strings.TrimSpace(item.ProductName)
		item.ShopName = strings.TrimSpace(item.ShopName)
		item.URL = strings.TrimSpace(item.URL)
		item.Currency = strings.ToUpper(strings.TrimSpace(item.Currency))

		if item.SearchTerm == "" {
			item.SearchTerm = item.ProductName
		}
		if item.ProductName == "" || item.SearchTerm == "" {
			continue
		}
		if item.Price <= 0 {
			continue
		}
		if item.URL == "" {
			continue
		}
		if _, err := url.ParseRequestURI(item.URL); err != nil {
			continue
		}
		if item.Currency == "" {
			item.Currency = "ARS"
		}

		validated = append(validated, item)
	}

	if len(validated) == 0 {
		return nil, fmt.Errorf("no valid results after validation")
	}

	return validated, nil
}
