package scraper

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

type MercadoLibreScraper struct{}

func NewMercadoLibreScraper() *MercadoLibreScraper {
	return &MercadoLibreScraper{}
}

func (s *MercadoLibreScraper) Name() string {
	return "MercadoLibre"
}

func (s *MercadoLibreScraper) BaseURL() string {
	return "https://lista.mercadolibre.com.ar"
}

func (s *MercadoLibreScraper) Search(query string) ([]Result, error) {
	var results []Result

	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	searchURL := fmt.Sprintf("https://lista.mercadolibre.com.ar/%s", strings.ReplaceAll(query, " ", "-"))

	c.OnHTML(".ui-search-result", func(e *colly.HTMLElement) {
		title := e.ChildText(".ui-search-item__title")
		if title == "" {
			title = e.ChildText(".poly-component__title")
		}

		priceStr := e.ChildText(".ui-price__fraction")
		if priceStr == "" {
			priceStr = e.ChildText(".poly-price__current-price .andes-money-amount__fraction")
		}

		link, _ := e.DOM.Find("a.poly-component__title").Attr("href")
		if link == "" {
			link, _ = e.DOM.Find("a.ui-search-link").Attr("href")
		}

		stockText := e.ChildText(".ui-search-stock-unavailable")
		hasStock := stockText == ""

		price := parsePrice(priceStr)

		if title != "" && price > 0 {
			results = append(results, Result{
				ProductName: title,
				Price:       price,
				Currency:    "ARS",
				URL:         link,
				HasStock:    hasStock,
			})
		}
	})

	err := c.Visit(searchURL)
	return results, err
}

func parsePrice(priceStr string) float64 {
	priceStr = strings.ReplaceAll(priceStr, "$", "")
	priceStr = strings.ReplaceAll(priceStr, " ", "")
	priceStr = strings.ReplaceAll(priceStr, ".", "")
	priceStr = strings.ReplaceAll(priceStr, ",", ".")

	var price float64
	fmt.Sscanf(priceStr, "%f", &price)
	return price
}
