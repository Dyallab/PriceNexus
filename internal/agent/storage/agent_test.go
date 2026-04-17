package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dyallo/pricenexus/internal/agent/shared"
	"github.com/sirupsen/logrus"
)

func TestNewStorageAgent(t *testing.T) {
	t.Parallel()

	logger := logrus.New()
	dbPath := filepath.Join(t.TempDir(), "prices.db")

	agent, err := NewStorageAgent(dbPath, logger)
	if err != nil {
		t.Fatalf("NewStorageAgent() unexpected error: %v", err)
	}
	if agent == nil {
		t.Fatal("Agent should not be nil")
	}
	defer agent.Close()
}

func TestSavePricesAndGetHistory(t *testing.T) {
	t.Parallel()

	logger := logrus.New()
	dbPath := filepath.Join(t.TempDir(), "prices.db")
	agent, err := NewStorageAgent(dbPath, logger)
	if err != nil {
		t.Fatalf("NewStorageAgent() unexpected error: %v", err)
	}
	defer agent.Close()

	prices := []shared.SearchResult{
		{
			SearchTerm:  "ps5",
			ProductName: "Test Product",
			Price:       100.50,
			Currency:    "ARS",
			URL:         "https://example.com/product",
			HasStock:    true,
			HasShipping: true,
			ShopName:    "Test Shop",
		},
	}

	if err := agent.SavePrices(context.Background(), prices); err != nil {
		t.Fatalf("SavePrices() unexpected error: %v", err)
	}

	history, err := agent.GetHistory(context.Background(), "ps5")
	if err != nil {
		t.Fatalf("GetHistory() unexpected error: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 history result, got %d", len(history))
	}

	if history[0].ShopName != "Test Shop" {
		t.Fatalf("expected shop name Test Shop, got %q", history[0].ShopName)
	}
	if history[0].SearchTerm != "ps5" {
		t.Fatalf("expected search term ps5, got %q", history[0].SearchTerm)
	}
}

func TestGetHistoryReturnsPriceHistoryRows(t *testing.T) {
	t.Parallel()

	logger := logrus.New()
	dbPath := filepath.Join(t.TempDir(), "prices.db")
	agent, err := NewStorageAgent(dbPath, logger)
	if err != nil {
		t.Fatalf("NewStorageAgent() unexpected error: %v", err)
	}
	defer agent.Close()

	prices := []shared.SearchResult{
		{
			SearchTerm:  "monitor 24",
			ProductName: "Monitor 24 pulgadas",
			Price:       90000,
			Currency:    "ARS",
			URL:         "https://tienda.example.com.ar/monitor",
			HasStock:    true,
			HasShipping: true,
			ShopName:    "Example",
		},
	}

	if err := agent.SavePrices(context.Background(), prices); err != nil {
		t.Fatalf("SavePrices() unexpected error: %v", err)
	}

	history, err := agent.GetHistory(context.Background(), "monitor 24")
	if err != nil {
		t.Fatalf("GetHistory() unexpected error: %v", err)
	}

	if len(history) == 0 {
		t.Fatal("expected non-empty history")
	}
}
