package models

type Shop struct {
	ID     int    `json:"id" db:"id"`
	Name   string `json:"name" db:"name"`
	URL    string `json:"url" db:"url"`
	Active bool   `json:"active" db:"active"`
}

type Product struct {
	ID         int    `json:"id" db:"id"`
	Name       string `json:"name" db:"name"`
	SearchTerm string `json:"search_term" db:"search_term"`
	CreatedAt  string `json:"created_at" db:"created_at"`
}

type Price struct {
	ID          int     `json:"id" db:"id"`
	ProductID   int     `json:"product_id" db:"product_id"`
	ShopID      int     `json:"shop_id" db:"shop_id"`
	Price       float64 `json:"price" db:"price"`
	Currency    string  `json:"currency" db:"currency"`
	HasStock    bool    `json:"has_stock" db:"has_stock"`
	HasShipping bool    `json:"has_shipping" db:"has_shipping"`
	URL         string  `json:"url" db:"url"`
	ScrapedAt   string  `json:"scraped_at" db:"scraped_at"`
}
