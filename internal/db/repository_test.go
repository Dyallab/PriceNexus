package db

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestRepositoryCacheRoundTrip(t *testing.T) {
	t.Parallel()

	logger := logrus.New()
	repoPath := filepath.Join(t.TempDir(), "prices.db")
	repo, err := NewRepository(repoPath, logger)
	if err != nil {
		t.Fatalf("NewRepository() unexpected error: %v", err)
	}
	defer repo.Close()

	if err := repo.SetSearchCache("ps5", []string{".com.ar"}, []string{"mercadolibre.com.ar"}, 10, []byte(`[{"url":"https://tienda.example.com.ar"}]`), time.Hour); err != nil {
		t.Fatalf("SetSearchCache() unexpected error: %v", err)
	}
	searchPayload, hit, err := repo.GetSearchCache("ps5", []string{".com.ar"}, []string{"mercadolibre.com.ar"}, 10)
	if err != nil {
		t.Fatalf("GetSearchCache() unexpected error: %v", err)
	}
	if !hit {
		t.Fatal("expected search cache hit")
	}
	if string(searchPayload) != `[{"url":"https://tienda.example.com.ar"}]` {
		t.Fatalf("unexpected search cache payload: %s", searchPayload)
	}

	if err := repo.SetPageCache("https://example.com.ar/product", "<html>ok</html>", time.Hour); err != nil {
		t.Fatalf("SetPageCache() unexpected error: %v", err)
	}
	pagePayload, hit, err := repo.GetPageCache("https://example.com.ar/product")
	if err != nil {
		t.Fatalf("GetPageCache() unexpected error: %v", err)
	}
	if !hit {
		t.Fatal("expected page cache hit")
	}
	if pagePayload != "<html>ok</html>" {
		t.Fatalf("unexpected page cache payload: %s", pagePayload)
	}

	if err := repo.SetExtractionCache("hash", "model", "v1", []byte(`[{"product_name":"X"}]`), time.Hour); err != nil {
		t.Fatalf("SetExtractionCache() unexpected error: %v", err)
	}
	extractionPayload, hit, err := repo.GetExtractionCache("hash", "model", "v1")
	if err != nil {
		t.Fatalf("GetExtractionCache() unexpected error: %v", err)
	}
	if !hit {
		t.Fatal("expected extraction cache hit")
	}
	if string(extractionPayload) != `[{"product_name":"X"}]` {
		t.Fatalf("unexpected extraction cache payload: %s", extractionPayload)
	}
}
