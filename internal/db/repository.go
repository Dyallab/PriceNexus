package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dyallo/pricenexus/internal/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	Init() error
	GetAllShops() ([]models.Shop, error)
	GetShopByID(id int) (models.Shop, error)
	GetShopByName(name string) (models.Shop, error)
	AddShop(shop models.Shop) (int64, error)
	GetProduct(searchTerm string) (models.Product, error)
	AddProduct(product models.Product) (int64, error)
	AddPrice(price models.Price) (int64, error)
	GetPricesByProduct(productID int) ([]models.Price, error)
	GetPriceHistoryByProduct(productID int) ([]models.Price, error)
	GetSearchCache(query string, allowedDomains, excludedDomains []string, maxResults int) ([]byte, bool, error)
	SetSearchCache(query string, allowedDomains, excludedDomains []string, maxResults int, payload []byte, ttl time.Duration) error
	GetPageCache(rawURL string) (string, bool, error)
	SetPageCache(rawURL, html string, ttl time.Duration) error
	GetExtractionCache(contentHash, modelName, promptVersion string) ([]byte, bool, error)
	SetExtractionCache(contentHash, modelName, promptVersion string, payload []byte, ttl time.Duration) error
	Close() error
}

type sqliteRepository struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewRepository(dbPath string, log *logrus.Logger) (Repository, error) {
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &sqliteRepository{
		db:  db,
		log: log,
	}

	if err := repo.applyMigrations(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *sqliteRepository) applyMigrations() error {
	migrationFile, err := findMigrationFile()
	if err != nil {
		return err
	}

	migration, err := os.ReadFile(migrationFile)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(string(migration))
	return err
}

func (r *sqliteRepository) Init() error {
	return nil
}

func (r *sqliteRepository) GetAllShops() ([]models.Shop, error) {
	var shops []models.Shop
	err := r.db.Select(&shops, "SELECT id, name, url, active FROM shops WHERE active = 1")
	return shops, err
}

func (r *sqliteRepository) GetShopByID(id int) (models.Shop, error) {
	var shop models.Shop
	err := r.db.Get(&shop, "SELECT id, name, url, active FROM shops WHERE id = ?", id)
	return shop, err
}

func (r *sqliteRepository) GetShopByName(name string) (models.Shop, error) {
	var shop models.Shop
	err := r.db.Get(&shop, "SELECT id, name, url, active FROM shops WHERE lower(name) = lower(?) LIMIT 1", name)
	return shop, err
}

func (r *sqliteRepository) AddShop(shop models.Shop) (int64, error) {
	result, err := r.db.Exec(
		"INSERT INTO shops (name, url, active) VALUES (?, ?, ?)",
		shop.Name, shop.URL, shop.Active,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *sqliteRepository) GetProduct(searchTerm string) (models.Product, error) {
	var product models.Product
	err := r.db.Get(&product,
		"SELECT id, name, search_term, created_at FROM products WHERE search_term = ?",
		searchTerm)
	return product, err
}

func (r *sqliteRepository) AddProduct(product models.Product) (int64, error) {
	result, err := r.db.Exec(
		"INSERT INTO products (name, search_term, created_at) VALUES (?, ?, ?)",
		product.Name, product.SearchTerm, time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *sqliteRepository) AddPrice(price models.Price) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO prices (product_id, shop_id, price, currency, has_stock, has_shipping, url, scraped_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		price.ProductID, price.ShopID, price.Price, price.Currency,
		price.HasStock, price.HasShipping, price.URL, time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}

	r.db.Exec(
		`INSERT INTO price_history (product_id, shop_id, price, currency, has_stock, has_shipping, url, scraped_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		price.ProductID, price.ShopID, price.Price, price.Currency,
		price.HasStock, price.HasShipping, price.URL, time.Now().Format(time.RFC3339),
	)

	return result.LastInsertId()
}

func (r *sqliteRepository) GetPricesByProduct(productID int) ([]models.Price, error) {
	var prices []models.Price
	err := r.db.Select(&prices,
		`SELECT id, product_id, shop_id, price, currency, has_stock, has_shipping, url, scraped_at 
		 FROM prices WHERE product_id = ? ORDER BY scraped_at DESC`,
		productID)
	return prices, err
}

func (r *sqliteRepository) GetPriceHistoryByProduct(productID int) ([]models.Price, error) {
	var prices []models.Price
	err := r.db.Select(&prices,
		`SELECT id, product_id, shop_id, price, currency, has_stock, has_shipping, url, scraped_at
		 FROM price_history WHERE product_id = ? ORDER BY scraped_at DESC`,
		productID)
	return prices, err
}

func (r *sqliteRepository) GetSearchCache(query string, allowedDomains, excludedDomains []string, maxResults int) ([]byte, bool, error) {
	cacheKey := buildSearchCacheKey(query, allowedDomains, excludedDomains, maxResults)
	var payload string
	err := r.db.Get(&payload, `
		SELECT result_json
		FROM search_cache
		WHERE cache_key = ?
		  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		LIMIT 1`, cacheKey)
	if err != nil {
		return nil, false, err
	}

	if _, err := r.db.Exec(`UPDATE search_cache SET last_used_at = CURRENT_TIMESTAMP WHERE cache_key = ?`, cacheKey); err != nil {
		r.log.Warnf("failed to refresh search cache timestamp: %v", err)
	}

	return []byte(payload), true, nil
}

func (r *sqliteRepository) SetSearchCache(query string, allowedDomains, excludedDomains []string, maxResults int, payload []byte, ttl time.Duration) error {
	cacheKey := buildSearchCacheKey(query, allowedDomains, excludedDomains, maxResults)
	expiresAt := cacheExpiresAt(ttl)
	resultHash := sha256Hex(payload)
	_, err := r.db.Exec(`
		INSERT INTO search_cache (
			cache_key, query, allowed_domains, excluded_domains, max_results,
			result_json, result_hash, created_at, last_used_at, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)
		ON CONFLICT(cache_key) DO UPDATE SET
			query = excluded.query,
			allowed_domains = excluded.allowed_domains,
			excluded_domains = excluded.excluded_domains,
			max_results = excluded.max_results,
			result_json = excluded.result_json,
			result_hash = excluded.result_hash,
			last_used_at = CURRENT_TIMESTAMP,
			expires_at = excluded.expires_at`,
		cacheKey, query, strings.Join(normalizeCacheDomains(allowedDomains), ","), strings.Join(normalizeCacheDomains(excludedDomains), ","), maxResults, string(payload), resultHash, expiresAt)
	return err
}

func (r *sqliteRepository) GetPageCache(rawURL string) (string, bool, error) {
	var html string
	err := r.db.Get(&html, `
		SELECT html_blob
		FROM page_cache
		WHERE url = ?
		  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		LIMIT 1`, rawURL)
	if err != nil {
		return "", false, err
	}

	if _, err := r.db.Exec(`UPDATE page_cache SET last_used_at = CURRENT_TIMESTAMP WHERE url = ?`, rawURL); err != nil {
		r.log.Warnf("failed to refresh page cache timestamp: %v", err)
	}

	return html, true, nil
}

func (r *sqliteRepository) SetPageCache(rawURL, html string, ttl time.Duration) error {
	expiresAt := cacheExpiresAt(ttl)
	contentHash := sha256Hex([]byte(html))
	_, err := r.db.Exec(`
		INSERT INTO page_cache (
			url, status_code, etag, last_modified, content_hash, html_blob,
			fetched_at, last_used_at, expires_at
		) VALUES (?, ?, NULL, NULL, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)
		ON CONFLICT(url) DO UPDATE SET
			status_code = excluded.status_code,
			content_hash = excluded.content_hash,
			html_blob = excluded.html_blob,
			fetched_at = CURRENT_TIMESTAMP,
			last_used_at = CURRENT_TIMESTAMP,
			expires_at = excluded.expires_at`,
		rawURL, 200, contentHash, html, expiresAt)
	return err
}

func (r *sqliteRepository) GetExtractionCache(contentHash, modelName, promptVersion string) ([]byte, bool, error) {
	var payload string
	err := r.db.Get(&payload, `
		SELECT output_json
		FROM extraction_cache
		WHERE content_hash = ? AND model_name = ? AND prompt_version = ?
		  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		LIMIT 1`, contentHash, modelName, promptVersion)
	if err != nil {
		return nil, false, err
	}

	if _, err := r.db.Exec(`
		UPDATE extraction_cache
		SET last_used_at = CURRENT_TIMESTAMP
		WHERE content_hash = ? AND model_name = ? AND prompt_version = ?`, contentHash, modelName, promptVersion); err != nil {
		r.log.Warnf("failed to refresh extraction cache timestamp: %v", err)
	}

	return []byte(payload), true, nil
}

func (r *sqliteRepository) SetExtractionCache(contentHash, modelName, promptVersion string, payload []byte, ttl time.Duration) error {
	expiresAt := cacheExpiresAt(ttl)
	outputHash := sha256Hex(payload)
	_, err := r.db.Exec(`
		INSERT INTO extraction_cache (
			content_hash, model_name, prompt_version, output_json, output_hash,
			created_at, last_used_at, expires_at
		) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)
		ON CONFLICT(content_hash, model_name, prompt_version) DO UPDATE SET
			output_json = excluded.output_json,
			output_hash = excluded.output_hash,
			last_used_at = CURRENT_TIMESTAMP,
			expires_at = excluded.expires_at`,
		contentHash, modelName, promptVersion, string(payload), outputHash, expiresAt)
	return err
}

func (r *sqliteRepository) Close() error {
	return r.db.Close()
}

func buildSearchCacheKey(query string, allowedDomains, excludedDomains []string, maxResults int) string {
	parts := []string{
		strings.TrimSpace(strings.ToLower(query)),
		strings.Join(normalizeCacheDomains(allowedDomains), ","),
		strings.Join(normalizeCacheDomains(excludedDomains), ","),
		fmt.Sprintf("%d", maxResults),
	}
	return sha256Hex([]byte(strings.Join(parts, "|")))
}

func normalizeCacheDomains(domains []string) []string {
	result := make([]string, 0, len(domains))
	seen := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		normalized := strings.ToLower(strings.TrimSpace(domain))
		if normalized == "" {
			continue
		}
		if !strings.HasPrefix(normalized, ".") {
			normalized = "." + normalized
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func cacheExpiresAt(ttl time.Duration) any {
	if ttl <= 0 {
		return nil
	}
	return time.Now().UTC().Add(ttl).Format(time.RFC3339)
}

func sha256Hex(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func findMigrationFile() (string, error) {
	candidates := []string{
		"migrations/001_init.sql",
		"../migrations/001_init.sql",
	}

	if _, file, _, ok := runtime.Caller(0); ok {
		baseDir := filepath.Dir(file)
		candidates = append(candidates,
			filepath.Join(baseDir, "../../migrations/001_init.sql"),
			filepath.Join(baseDir, "../../../migrations/001_init.sql"),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", os.ErrNotExist
}
