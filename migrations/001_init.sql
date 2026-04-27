-- PriceNexus SQLite Schema
-- Migration: 001_init.sql

-- Shops table
CREATE TABLE IF NOT EXISTS shops (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    active INTEGER DEFAULT 1,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    search_term TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Prices table (current prices)
CREATE TABLE IF NOT EXISTS prices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL,
    shop_id INTEGER NOT NULL,
    price REAL NOT NULL,
    currency TEXT DEFAULT 'ARS',
    has_stock INTEGER DEFAULT 0,
    has_shipping INTEGER DEFAULT 0,
    url TEXT NOT NULL,
    scraped_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE
);

-- Price history table (audit log)
CREATE TABLE IF NOT EXISTS price_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL,
    shop_id INTEGER NOT NULL,
    price REAL NOT NULL,
    currency TEXT DEFAULT 'ARS',
    has_stock INTEGER DEFAULT 0,
    has_shipping INTEGER DEFAULT 0,
    url TEXT NOT NULL,
    scraped_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_prices_product ON prices(product_id);
CREATE INDEX IF NOT EXISTS idx_prices_shop ON prices(shop_id);
CREATE INDEX IF NOT EXISTS idx_price_history_product ON price_history(product_id);
CREATE INDEX IF NOT EXISTS idx_price_history_scraped ON price_history(scraped_at);

-- Search cache (OpenRouter search results)
CREATE TABLE IF NOT EXISTS search_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cache_key TEXT NOT NULL UNIQUE,
    query TEXT NOT NULL,
    allowed_domains TEXT NOT NULL,
    excluded_domains TEXT NOT NULL,
    max_results INTEGER NOT NULL,
    result_json TEXT NOT NULL,
    result_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TEXT NULL
);

CREATE INDEX IF NOT EXISTS idx_search_cache_expires ON search_cache(expires_at);

-- Page cache (fetched HTML)
CREATE TABLE IF NOT EXISTS page_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL UNIQUE,
    status_code INTEGER NOT NULL,
    etag TEXT NULL,
    last_modified TEXT NULL,
    content_hash TEXT NOT NULL,
    html_blob TEXT NOT NULL,
    fetched_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TEXT NULL
);

CREATE INDEX IF NOT EXISTS idx_page_cache_content_hash ON page_cache(content_hash);
CREATE INDEX IF NOT EXISTS idx_page_cache_expires ON page_cache(expires_at);

-- Extraction cache (HTML -> products)
CREATE TABLE IF NOT EXISTS extraction_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content_hash TEXT NOT NULL,
    model_name TEXT NOT NULL,
    prompt_version TEXT NOT NULL,
    output_json TEXT NOT NULL,
    output_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TEXT NULL,
    UNIQUE(content_hash, model_name, prompt_version)
);

CREATE INDEX IF NOT EXISTS idx_extraction_cache_expires ON extraction_cache(expires_at);
