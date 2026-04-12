package pageloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type PageLoader struct {
	client *http.Client
}

func NewPageLoader() *PageLoader {
	return &PageLoader{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (pl *PageLoader) LoadHTML(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Use realistic browser headers to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Cache-Control", "max-age=0")

	resp, err := pl.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	return string(body), nil
}

func (pl *PageLoader) ExtractSection(html string, selector string) string {
	startTag := fmt.Sprintf("<%s", selector)
	endTag := fmt.Sprintf("</%s>", selector)

	startIdx := strings.Index(html, startTag)
	if startIdx == -1 {
		return ""
	}

	endIdx := strings.Index(html[startIdx:], endTag)
	if endIdx == -1 {
		return ""
	}

	return html[startIdx : startIdx+endIdx+len(endTag)]
}
