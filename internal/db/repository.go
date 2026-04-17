package db

import (
	"os"
	"path/filepath"
	"runtime"
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

func (r *sqliteRepository) Close() error {
	return r.db.Close()
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
