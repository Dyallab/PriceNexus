package scraper

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

type TecnoshopsScraper struct{}

func NewTecnoshopsScraper() *TecnoshopsScraper {
	return &TecnoshopsScraper{}
}

func (s *TecnoshopsScraper) Name() string {
	return "Tecnoshops"
}

func (s *TecnoshopsScraper) BaseURL() string {
	return "https://www.tecnoshops.com.ar"
}

func (s *TecnoshopsScraper) Search(query string) ([]Result, error) {
	var results []Result

	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	searchURL := fmt.Sprintf("https://www.tecnoshops.com.ar/search?q=%s", query)

	c.OnHTML(".product-item", func(e *colly.HTMLElement) {
		title := e.ChildText(".product-name")
		priceStr := e.ChildText(".price")
		if priceStr == "" {
			priceStr = e.ChildText(".product-price")
		}

		link, _ := e.DOM.Find("a.product-link").Attr("href")

		stockText := e.ChildText(".stock-status")
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
