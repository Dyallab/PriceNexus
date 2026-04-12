package models

import "time"

type Shop struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Active bool   `json:"active"`
}

type Product struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	SearchTerm string    `json:"search_term"`
	CreatedAt  time.Time `json:"created_at"`
}

type Price struct {
	ID          int       `json:"id"`
	ProductID   int       `json:"product_id"`
	ShopID      int       `json:"shop_id"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	HasStock    bool      `json:"has_stock"`
	HasShipping bool      `json:"has_shipping"`
	URL         string    `json:"url"`
	ScrapedAt   time.Time `json:"scraped_at"`
}
