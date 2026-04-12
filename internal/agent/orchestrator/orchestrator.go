package orchestrator

import (
	"context"
	"fmt"

	"github.com/dyallo/pricenexus/internal/agent"
	"github.com/dyallo/pricenexus/internal/agent/dataextractor"
	"github.com/dyallo/pricenexus/internal/agent/pageloader"
	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/dyallo/pricenexus/internal/agent/storage"
	"github.com/dyallo/pricenexus/internal/agent/urlfinder"
	"github.com/dyallo/pricenexus/internal/agent/validator"
	"github.com/sirupsen/logrus"
)

type Orchestrator struct {
	urlFinder     *urlfinder.URLFinderAgent
	pageLoader    *pageloader.PageLoader
	dataExtractor *dataextractor.DataExtractorAgent
	validator     *validator.ValidatorAgent
	storage       *storage.StorageAgent
	logger        *logrus.Logger
}

func NewOrchestrator(dbPath string, logger *logrus.Logger) (*Orchestrator, error) {
	config := agent.LoadFromEnv()
	_, webSearcherLLM, dataExtractorLLM, err := agent.CreateLLMs(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLMs: %w", err)
	}

	if webSearcherLLM == nil {
		return nil, fmt.Errorf("web searcher LLM is nil after initialization")
	}

	if dataExtractorLLM == nil {
		return nil, fmt.Errorf("data extractor LLM is nil after initialization")
	}

	urlFinder, err := urlfinder.NewURLFinderAgent(webSearcherLLM)
	if err != nil {
		return nil, err
	}

	pageLoader := pageloader.NewPageLoader()

	dataExtractor, err := dataextractor.NewDataExtractorAgent(dataExtractorLLM)
	if err != nil {
		return nil, err
	}

	validatorAgent, err := validator.NewValidatorAgent(dataExtractorLLM)
	if err != nil {
		return nil, err
	}

	storageAgent, err := storage.NewStorageAgent(dbPath, logger)
	if err != nil {
		return nil, err
	}

	return &Orchestrator{
		urlFinder:     urlFinder,
		pageLoader:    pageLoader,
		dataExtractor: dataExtractor,
		validator:     validatorAgent,
		storage:       storageAgent,
		logger:        logger,
	}, nil
}

func (o *Orchestrator) Search(ctx context.Context, query string) ([]shared.SearchResult, error) {
	o.logger.Infof("Starting search for: %s", query)

	// Search for URLs using the URL finder agent
	// The agent will search for small/niche shops using the LLM
	urls, err := o.urlFinder.FindURLs(ctx, query, "")
	if err != nil {
		o.logger.Warnf("Error finding URLs: %v", err)
		return nil, fmt.Errorf("error finding URLs: %w", err)
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("no URLs found for query: %s", query)
	}

	var allResults []shared.SearchResult
	for _, url := range urls {
		o.logger.Infof("Loading page: %s", url)

		html, err := o.pageLoader.LoadHTML(ctx, url)
		if err != nil {
			o.logger.Warnf("Error loading page %s: %v", url, err)
			continue
		}

		results, err := o.dataExtractor.Extract(ctx, html)
		if err != nil {
			o.logger.Warnf("Error extracting data from %s: %v", url, err)
			continue
		}

		allResults = append(allResults, results...)
	}

	validatedResults, err := o.validator.Validate(ctx, allResults)
	if err != nil {
		o.logger.Warnf("Error validating results: %v", err)
		validatedResults = allResults
	}

	if len(validatedResults) > 0 {
		err = o.storage.SavePrices(ctx, validatedResults)
		if err != nil {
			o.logger.Warnf("Could not save prices: %v", err)
		}
	}

	return validatedResults, nil
}

func (o *Orchestrator) GetHistory(ctx context.Context, productName string) ([]shared.SearchResult, error) {
	return o.storage.GetHistory(ctx, productName)
}

func (o *Orchestrator) Close() error {
	return o.storage.Close()
}
