package scraper

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

type GarbarinoScraper struct{}

func NewGarbarinoScraper() *GarbarinoScraper {
	return &GarbarinoScraper{}
}

func (s *GarbarinoScraper) Name() string {
	return "Garbarino"
}

func (s *GarbarinoScraper) BaseURL() string {
	return "https://www.garbarino.com"
}

func (s *GarbarinoScraper) Search(query string) ([]Result, error) {
	var results []Result

	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	searchURL := fmt.Sprintf("https://www.garbarino.com/ar/search?searchTerm=%s", query)

	c.OnHTML(".product-card-container", func(e *colly.HTMLElement) {
		title := e.ChildText(".product-card-title")
		priceStr := e.ChildText(".product-price-symbol + span")
		if priceStr == "" {
			priceStr = e.ChildText(".price-container .value")
		}

		link, _ := e.DOM.Find("a").First().Attr("href")
		if link != "" && !strings.HasPrefix(link, "http") {
			link = "https://www.garbarino.com" + link
		}

		stockText := e.ChildText(".stock-label")
		hasStock := !strings.Contains(strings.ToLower(stockText), "sin stock")

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
