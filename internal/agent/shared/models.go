package shared

type AgentType string

const (
	AgentOrchestrator  AgentType = "orchestrator"
	AgentWebSearcher   AgentType = "websearcher"
	AgentDataExtractor AgentType = "dataextractor"
	AgentStorage       AgentType = "storage"
)

type SearchResult struct {
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	URL         string  `json:"url"`
	HasStock    bool    `json:"has_stock"`
	HasShipping bool    `json:"has_shipping"`
	ShopName    string  `json:"shop_name"`
}

type SearchRequest struct {
	Query string `json:"query"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

type StorageRequest struct {
	ProductID   int     `json:"product_id"`
	ShopID      int     `json:"shop_id"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	HasStock    bool    `json:"has_stock"`
	HasShipping bool    `json:"has_shipping"`
	URL         string  `json:"url"`
}

type StorageResponse struct {
	Success bool  `json:"success"`
	ID      int64 `json:"id"`
}

type AgentMessage struct {
	Type    string      `json:"type"`
	From    AgentType   `json:"from"`
	To      AgentType   `json:"to"`
	Payload interface{} `json:"payload"`
}
