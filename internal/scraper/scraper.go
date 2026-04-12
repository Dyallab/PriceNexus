package scraper

import "github.com/dyallo/pricenexus/internal/models"

type Result struct {
	ProductName string
	Price       float64
	Currency    string
	URL         string
	HasStock    bool
	HasShipping bool
}

type Scraper interface {
	Name() string
	BaseURL() string
	Search(query string) ([]Result, error)
}

func GetAllScrapers() []Scraper {
	return []Scraper{
		NewMercadoLibreScraper(),
		NewGarbarinoScraper(),
		NewTecnoshopsScraper(),
	}
}

func ResultToPrice(result Result, productID, shopID int) models.Price {
	return models.Price{
		ProductID:   productID,
		ShopID:      shopID,
		Price:       result.Price,
		Currency:    result.Currency,
		HasStock:    result.HasStock,
		HasShipping: result.HasShipping,
		URL:         result.URL,
	}
}
