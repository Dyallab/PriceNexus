package orchestrator

import (
	"context"
	"fmt"

	"github.com/dyallo/pricenexus/internal/agent"
	"github.com/dyallo/pricenexus/internal/agent/dataextractor"
	"github.com/dyallo/pricenexus/internal/agent/pageloader"
	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/dyallo/pricenexus/internal/agent/storage"
	"github.com/dyallo/pricenexus/internal/agent/validator"
	"github.com/dyallo/pricenexus/internal/agent/websearcher"
	"github.com/sirupsen/logrus"
)

type Orchestrator struct {
	webSearcher   *websearcher.WebSearcherAgent
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

	webSearcherAgent, err := websearcher.NewWebSearcherAgent(webSearcherLLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create web searcher agent: %w", err)
	}

	pageLoader := pageloader.NewPageLoader()

	dataExtractor, err := dataextractor.NewDataExtractorAgent(dataExtractorLLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create data extractor agent: %w", err)
	}

	validatorAgent, err := validator.NewValidatorAgent(dataExtractorLLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator agent: %w", err)
	}

	storageAgent, err := storage.NewStorageAgent(dbPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage agent: %w", err)
	}

	return &Orchestrator{
		webSearcher:   webSearcherAgent,
		pageLoader:    pageLoader,
		dataExtractor: dataExtractor,
		validator:     validatorAgent,
		storage:       storageAgent,
		logger:        logger,
	}, nil
}

func (o *Orchestrator) Search(ctx context.Context, query string) ([]shared.SearchResult, error) {
	o.logger.Infof("Starting search for: %s", query)
	o.logger.Infof("Step 1/5: Searching the web for URLs...")

	// First, use the WebSearcherAgent to find relevant URLs on the web
	searchResults, err := o.webSearcher.Search(ctx, query)
	if err != nil {
		o.logger.Warnf("Error searching the web: %v", err)
		return nil, fmt.Errorf("error searching the web: %w", err)
	}

	if len(searchResults) == 0 {
		return nil, fmt.Errorf("no URLs found for query: %s", query)
	}

	o.logger.Infof("✓ Found %d URLs from web search", len(searchResults))

	// Extract unique URLs
	uniqueURLs := make(map[string]bool)
	var urls []string
	for _, result := range searchResults {
		if !uniqueURLs[result.URL] {
			uniqueURLs[result.URL] = true
			urls = append(urls, result.URL)
		}
	}

	o.logger.Infof("Step 2/5: Processing %d unique URLs for data extraction", len(urls))

	var allResults []shared.SearchResult
	for i, url := range urls {
		o.logger.Infof("  [%d/%d] Loading page: %s", i+1, len(urls), url)

		html, err := o.pageLoader.LoadHTML(ctx, url)
		if err != nil {
			o.logger.Warnf("    ✗ Error loading page: %v", err)
			continue
		}

		o.logger.Infof("    ✓ Page loaded (%d bytes)", len(html))
		o.logger.Infof("    Extracting product data...")

		results, err := o.dataExtractor.Extract(ctx, html)
		if err != nil {
			o.logger.Warnf("    ✗ Error extracting data: %v", err)
			continue
		}

		if len(results) > 0 {
			o.logger.Infof("    ✓ Extracted %d products", len(results))
		} else {
			o.logger.Infof("    ℹ No products found on this page")
		}

		allResults = append(allResults, results...)
	}

	if len(allResults) == 0 {
		o.logger.Warnf("Step 2/5: ✗ No products extracted from any page")
		return nil, fmt.Errorf("no products found")
	}

	o.logger.Infof("Step 2/5: ✓ Total products extracted: %d", len(allResults))

	o.logger.Infof("Step 3/5: Validating extracted data...")
	validatedResults, err := o.validator.Validate(ctx, allResults)
	if err != nil {
		o.logger.Warnf("Step 3/5: ✗ Error validating results: %v", err)
		validatedResults = allResults
	}

	o.logger.Infof("Step 3/5: ✓ Validation complete (%d valid products)", len(validatedResults))

	if len(validatedResults) > 0 {
		o.logger.Infof("Step 4/5: Saving prices to database...")
		err = o.storage.SavePrices(ctx, validatedResults)
		if err != nil {
			o.logger.Warnf("Step 4/5: ✗ Error saving prices: %v", err)
			return validatedResults, nil
		}
		o.logger.Infof("Step 4/5: ✓ Prices saved successfully")
	}

	o.logger.Infof("Step 5/5: ✓ Search complete!")
	return validatedResults, nil
}

func (o *Orchestrator) GetHistory(ctx context.Context, productName string) ([]shared.SearchResult, error) {
	return o.storage.GetHistory(ctx, productName)
}

func (o *Orchestrator) Close() error {
	return o.storage.Close()
}
