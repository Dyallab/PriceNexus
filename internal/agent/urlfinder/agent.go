package urlfinder

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
)

// Popular sites to exclude from search results
var popularSites = []string{
	"mercadolibre.com",
	"mercadolibre.com.ar",
	"garbarino.com",
	"garbarino.com.ar",
	"tecnoshops.com.ar",
	"amazon.com",
	"amazon.com.ar",
	"ebay.com",
	"facebook.com",
	"instagram.com",
	"tienda.nucleo.com.ar",
	"mercadolibre",
}

type StoreConfig struct {
	Name        string
	BaseURL     string
	SearchPath  string
	Enabled     bool
	IsSmallShop bool
}

type URLFinderAgent struct {
	llm    llms.Model
	logger *logrus.Logger
	stores map[string]StoreConfig
}

func DefaultStoreConfigs() map[string]StoreConfig {
	return map[string]StoreConfig{
		"CompuGamer": {
			Name:        "CompuGamer",
			BaseURL:     "https://www.compagamer.com",
			SearchPath:  "search?q=%s",
			Enabled:     true,
			IsSmallShop: true,
		},
		"CompuOro": {
			Name:        "CompuOro",
			BaseURL:     "https://www.compuoro.com.ar",
			SearchPath:  "catalogsearch/result/?q=%s",
			Enabled:     true,
			IsSmallShop: true,
		},
		"Mexx": {
			Name:        "Mexx",
			BaseURL:     "https://www.mexx.com.ar",
			SearchPath:  "buscar/?p=%s",
			Enabled:     true,
			IsSmallShop: true,
		},
		"Venex": {
			Name:        "Venex",
			BaseURL:     "https://www.venex.com.ar",
			SearchPath:  "busqueda?q=%s",
			Enabled:     true,
			IsSmallShop: true,
		},
		"FullHard": {
			Name:        "FullHard",
			BaseURL:     "https://www.fullh4rd.com.ar",
			SearchPath:  "busqueda?q=%s",
			Enabled:     true,
			IsSmallShop: true,
		},
	}
}

func NewURLFinderAgent(llm llms.Model) (*URLFinderAgent, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &URLFinderAgent{
		llm:    llm,
		logger: logger,
		stores: DefaultStoreConfigs(),
	}, nil
}

// FindURLs generates search URLs for configured shops
func (u *URLFinderAgent) FindURLs(ctx context.Context, query string, shopName string) ([]string, error) {
	u.logger.Infof("Finding URLs for query '%s'", query)

	var urls []string

	// If a specific shop is requested, use only that shop
	if shopName != "" {
		store, exists := u.stores[shopName]
		if !exists {
			u.logger.Warnf("Store not configured: %s", shopName)
			return []string{}, nil
		}

		if !store.Enabled {
			u.logger.Warnf("Store disabled: %s", shopName)
			return []string{}, nil
		}

		searchURL := u.generateSearchURL(store, query)
		urls = append(urls, searchURL)
		u.logger.Infof("Generated URL for %s: %s", shopName, searchURL)
	} else {
		// Search in all configured shops
		for name, store := range u.stores {
			if !store.Enabled {
				continue
			}

			searchURL := u.generateSearchURL(store, query)
			urls = append(urls, searchURL)
			u.logger.Infof("Generated URL for %s: %s", name, searchURL)
		}
	}

	return urls, nil
}

func (u *URLFinderAgent) generateSearchURL(store StoreConfig, query string) string {
	if store.Name == "MercadoLibre" {
		// Replace spaces with dashes and convert to lowercase for MercadoLibre
		formattedQuery := strings.ReplaceAll(strings.ToLower(query), " ", "-")
		return fmt.Sprintf("%s/%s", store.BaseURL, formattedQuery)
	}

	encodedQuery := url.QueryEscape(query)

	if strings.Contains(store.SearchPath, "%s") {
		return fmt.Sprintf("%s/%s", store.BaseURL, fmt.Sprintf(store.SearchPath, encodedQuery))
	}

	return fmt.Sprintf("%s/%s", store.BaseURL, store.SearchPath)
}

// AddStore allows adding new shops dynamically
func (u *URLFinderAgent) AddStore(name, baseURL, searchPath string, isSmallShop bool) {
	u.stores[name] = StoreConfig{
		Name:        name,
		BaseURL:     baseURL,
		SearchPath:  searchPath,
		Enabled:     true,
		IsSmallShop: isSmallShop,
	}
	u.logger.Infof("Added store: %s (%s/%s)", name, baseURL, searchPath)
}

// ListStores returns all configured stores
func (u *URLFinderAgent) ListStores() []StoreConfig {
	var stores []StoreConfig
	for _, store := range u.stores {
		stores = append(stores, store)
	}
	return stores
}
