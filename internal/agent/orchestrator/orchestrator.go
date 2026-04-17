package orchestrator

import (
	"context"
	"fmt"

	"github.com/dyallo/pricenexus/internal/agent"
	"github.com/dyallo/pricenexus/internal/agent/dataextractor"
	"github.com/dyallo/pricenexus/internal/agent/normalizer"
	"github.com/dyallo/pricenexus/internal/agent/pageloader"
	"github.com/dyallo/pricenexus/internal/agent/preflight"
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
	normalizer    *normalizer.Agent
	validator     *validator.ValidatorAgent
	storage       *storage.StorageAgent
	logger        *logrus.Logger
}

func NewOrchestrator(dbPath string, logger *logrus.Logger) (*Orchestrator, error) {
	llmConfig := agent.LoadFromEnv()
	searchConfig := agent.LoadSearchConfigFromEnv()
	_, webSearcherLLM, dataExtractorLLM, err := agent.CreateLLMs(llmConfig, searchConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLMs: %w", err)
	}

	if webSearcherLLM == nil {
		return nil, fmt.Errorf("web searcher LLM is nil after initialization")
	}

	if dataExtractorLLM == nil {
		return nil, fmt.Errorf("data extractor LLM is nil after initialization")
	}

	webSearcherAgent, err := websearcher.NewWebSearcherAgent(
		webSearcherLLM,
		searchConfig.AllowedDomains,
		searchConfig.DefaultCurrency,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create web searcher agent: %w", err)
	}

	pageLoader := pageloader.NewPageLoader()

	dataExtractor, err := dataextractor.NewDataExtractorAgent(dataExtractorLLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create data extractor agent: %w", err)
	}

	normalizerAgent := normalizer.NewAgent(searchConfig.DefaultCurrency)

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
		normalizer:    normalizerAgent,
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
	uniqueSources := make(map[string]shared.SearchResult)
	var urls []string
	for _, result := range searchResults {
		if !uniqueURLs[result.URL] {
			uniqueURLs[result.URL] = true
			uniqueSources[result.URL] = result
			urls = append(urls, result.URL)
		}
	}

	o.logger.Infof("Step 2/5: Preflighting %d unique URLs", len(urls))

	preflighter := preflight.NewPreflighter()
	preflightedSources := make(map[string]shared.SearchResult)
	preflightedURLs := make([]string, 0, len(urls))
	preflightedSeen := make(map[string]struct{}, len(urls))
	for _, rawURL := range urls {
		result := preflighter.Check(ctx, rawURL)
		if !result.ShouldExtract() {
			o.logger.Warnf(
				"  ✗ Preflight rejected: %s (%s%s)",
				rawURL,
				result.Status,
				formatPreflightError(result),
			)
			continue
		}

		finalURL := result.FinalURL
		if finalURL == "" {
			finalURL = rawURL
		}
		if _, exists := preflightedSeen[finalURL]; exists {
			continue
		}

		preflightedSeen[finalURL] = struct{}{}
		preflightedURLs = append(preflightedURLs, finalURL)
		preflightedSources[finalURL] = uniqueSources[rawURL]
		o.logger.Infof("  ✓ Preflight accepted: %s -> %s", rawURL, finalURL)
	}

	if len(preflightedURLs) == 0 {
		o.logger.Warnf("Step 2/5: ✗ No valid URLs remained after preflight")
		return nil, fmt.Errorf("no valid URLs found after preflight")
	}

	o.logger.Infof("Step 2/5: ✓ %d URLs passed preflight", len(preflightedURLs))
	o.logger.Infof("Step 3/5: Processing %d preflighted URLs for data extraction", len(preflightedURLs))

	var allResults []shared.SearchResult
	for i, url := range preflightedURLs {
		o.logger.Infof("  [%d/%d] Loading page: %s", i+1, len(preflightedURLs), url)
		source := preflightedSources[url]

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

		results = o.normalizer.Normalize(results, normalizer.SourceContext{
			Query:     query,
			Source:    source,
			SourceURL: url,
		})

		if len(results) > 0 {
			o.logger.Infof("    ✓ Extracted %d products", len(results))
		} else {
			o.logger.Infof("    ℹ No products found on this page")
		}

		allResults = append(allResults, results...)
	}

	if len(allResults) == 0 {
		o.logger.Warnf("Step 3/5: ✗ No products extracted from any page")
		return nil, fmt.Errorf("no products found")
	}

	o.logger.Infof("Step 3/5: ✓ Total products extracted: %d", len(allResults))

	o.logger.Infof("Step 4/5: Validating extracted data...")
	validatedResults, err := o.validator.Validate(ctx, allResults)
	if err != nil {
		o.logger.Warnf("Step 4/5: ✗ Error validating results: %v", err)
		validatedResults = allResults
	}

	o.logger.Infof("Step 4/5: ✓ Validation complete (%d valid products)", len(validatedResults))

	if len(validatedResults) > 0 {
		o.logger.Infof("Step 5/5: Saving prices to database...")
		err = o.storage.SavePrices(ctx, validatedResults)
		if err != nil {
			o.logger.Warnf("Step 5/5: ✗ Error saving prices: %v", err)
			return validatedResults, nil
		}
		o.logger.Infof("Step 5/5: ✓ Prices saved successfully")
	}

	o.logger.Infof("✓ Search complete!")
	return validatedResults, nil
}

func formatPreflightError(result preflight.PreflightResult) string {
	if result.ErrorMsg == "" {
		return ""
	}

	return fmt.Sprintf(": %s", result.ErrorMsg)
}

func (o *Orchestrator) GetHistory(ctx context.Context, productName string) ([]shared.SearchResult, error) {
	return o.storage.GetHistory(ctx, productName)
}

func (o *Orchestrator) Close() error {
	return o.storage.Close()
}
