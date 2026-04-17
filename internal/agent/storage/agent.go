package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/dyallo/pricenexus/internal/db"
	"github.com/dyallo/pricenexus/internal/models"
	"github.com/sirupsen/logrus"
)

type StorageAgent struct {
	repo   db.Repository
	logger *logrus.Logger
}

func NewStorageAgent(dbPath string, logger *logrus.Logger) (*StorageAgent, error) {
	repo, err := db.NewRepository(dbPath, logger)
	if err != nil {
		return nil, err
	}

	return &StorageAgent{
		repo:   repo,
		logger: logger,
	}, nil
}

func (s *StorageAgent) SavePrices(ctx context.Context, prices []shared.SearchResult) error {
	_ = ctx

	for _, price := range prices {
		productSearchTerm := strings.TrimSpace(price.SearchTerm)
		if productSearchTerm == "" {
			productSearchTerm = strings.TrimSpace(price.ProductName)
		}

		product, err := s.repo.GetProduct(productSearchTerm)
		if err != nil {
			s.logger.Infof("Product %s not found, creating new entry", productSearchTerm)
			productID, addErr := s.repo.AddProduct(models.Product{
				Name:       strings.TrimSpace(price.ProductName),
				SearchTerm: productSearchTerm,
			})
			if addErr != nil {
				return fmt.Errorf("error creating product %q: %w", productSearchTerm, addErr)
			}
			product = models.Product{
				ID:         int(productID),
				Name:       strings.TrimSpace(price.ProductName),
				SearchTerm: productSearchTerm,
			}
		}

		shopName := strings.TrimSpace(price.ShopName)
		if shopName == "" {
			shopName = deriveShopName(price.URL)
		}

		shop, err := s.repo.GetShopByName(shopName)
		if err != nil {
			s.logger.Infof("Shop %s not found, creating new entry", shopName)
			shopURL := deriveShopURL(price.URL)
			shopID, addErr := s.repo.AddShop(models.Shop{
				Name:   shopName,
				URL:    shopURL,
				Active: true,
			})
			if addErr != nil {
				return fmt.Errorf("error creating shop %q: %w", shopName, addErr)
			}
			shop = models.Shop{
				ID:     int(shopID),
				Name:   shopName,
				URL:    shopURL,
				Active: true,
			}
		}

		priceModel := models.Price{
			ProductID:   product.ID,
			ShopID:      shop.ID,
			Price:       price.Price,
			Currency:    price.Currency,
			HasStock:    price.HasStock,
			HasShipping: price.HasShipping,
			URL:         price.URL,
			ScrapedAt:   time.Now().Format(time.RFC3339),
		}

		_, err = s.repo.AddPrice(priceModel)
		if err != nil {
			return fmt.Errorf("error saving price: %w", err)
		}
	}

	return nil
}

func (s *StorageAgent) GetHistory(ctx context.Context, productName string) ([]shared.SearchResult, error) {
	_ = ctx

	product, err := s.repo.GetProduct(productName)
	if err != nil {
		return nil, err
	}

	prices, err := s.repo.GetPriceHistoryByProduct(product.ID)
	if err != nil {
		return nil, err
	}

	var results []shared.SearchResult
	for _, price := range prices {
		shopName := fmt.Sprintf("Tienda %d", price.ShopID)
		shop, err := s.repo.GetShopByID(price.ShopID)
		if err == nil && strings.TrimSpace(shop.Name) != "" {
			shopName = shop.Name
		}

		results = append(results, shared.SearchResult{
			SearchTerm:  product.SearchTerm,
			ProductName: product.Name,
			Price:       price.Price,
			Currency:    price.Currency,
			URL:         price.URL,
			HasStock:    price.HasStock,
			HasShipping: price.HasShipping,
			ShopName:    shopName,
		})
	}

	return results, nil
}

func (s *StorageAgent) Close() error {
	return s.repo.Close()
}

func deriveShopURL(rawURL string) string {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return strings.TrimSpace(rawURL)
	}

	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
}

func deriveShopName(rawURL string) string {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "Unknown"
	}

	hostname := strings.TrimPrefix(strings.ToLower(parsedURL.Hostname()), "www.")
	hostname = strings.TrimSuffix(hostname, ".com.ar")
	hostname = strings.TrimSuffix(hostname, ".ar")
	hostname = strings.TrimSuffix(hostname, ".com")
	hostname = strings.Trim(hostname, ".")
	if hostname == "" {
		return "Unknown"
	}

	return strings.ToUpper(hostname[:1]) + hostname[1:]
}
