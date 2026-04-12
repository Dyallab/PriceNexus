package storage

import (
	"context"
	"fmt"
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
	for _, price := range prices {
		product, err := s.repo.GetProduct(price.ProductName)
		if err != nil {
			s.logger.Infof("Product %s not found, creating new entry", price.ProductName)
		}

		shop, err := s.repo.GetProduct(price.ShopName)
		if err != nil {
			s.logger.Infof("Shop %s not found, creating new entry", price.ShopName)
		}

		priceModel := models.Price{
			ProductID:   product.ID,
			ShopID:      shop.ID,
			Price:       price.Price,
			Currency:    price.Currency,
			HasStock:    price.HasStock,
			HasShipping: price.HasShipping,
			URL:         price.URL,
			ScrapedAt:   time.Now(),
		}

		_, err = s.repo.AddPrice(priceModel)
		if err != nil {
			return fmt.Errorf("error saving price: %w", err)
		}
	}

	return nil
}

func (s *StorageAgent) GetHistory(ctx context.Context, productName string) ([]shared.SearchResult, error) {
	product, err := s.repo.GetProduct(productName)
	if err != nil {
		return nil, err
	}

	prices, err := s.repo.GetPricesByProduct(product.ID)
	if err != nil {
		return nil, err
	}

	var results []shared.SearchResult
	for _, price := range prices {
		results = append(results, shared.SearchResult{
			ProductName: product.Name,
			Price:       price.Price,
			Currency:    price.Currency,
			URL:         price.URL,
			HasStock:    price.HasStock,
			HasShipping: price.HasShipping,
		})
	}

	return results, nil
}

func (s *StorageAgent) Close() error {
	return s.repo.Close()
}
