package pageloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/dyallo/pricenexus/internal/db"
)

type PageLoader struct {
	client *http.Client
	repo   db.Repository
}

func NewPageLoader(repo db.Repository) *PageLoader {
	return &PageLoader{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		repo: repo,
	}
}

// LoadHTML fetches HTML using simple HTTP request (no JavaScript rendering)
// Use this for simple pages that don't require JavaScript
func (pl *PageLoader) LoadHTML(ctx context.Context, url string) (string, error) {
	if pl.repo != nil {
		if cachedHTML, hit, err := pl.repo.GetPageCache(url); err == nil && hit {
			return cachedHTML, nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Use realistic browser headers to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9,en;q=0.8")
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

	if pl.repo != nil {
		if err := pl.repo.SetPageCache(url, string(body), 7*24*time.Hour); err != nil {
			// Cache writes are best-effort only.
		}
	}

	return string(body), nil
}

// LoadRenderedHTML fetches and renders a page using headless Chrome (chromedp)
// Use this for JavaScript-rendered pages where content is loaded dynamically
func (pl *PageLoader) LoadRenderedHTML(ctx context.Context, url string) (string, error) {
	var htmlContent string

	// Create a timeout context for the browser operations
	browserCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create exec allocator to launch Chrome with headless mode
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		browserCtx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("no-first-run", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	// Create a browser context
	browserContext, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// Run chromedp tasks
	err := chromedp.Run(browserContext,
		chromedp.Navigate(url),
		// Wait for the page to be fully loaded by waiting for network idle
		chromedp.WaitReady(`body`, chromedp.ByQuery),
		// Additional wait to ensure any lazy-loaded content is rendered
		chromedp.Sleep(2*time.Second),
		// Capture the rendered HTML
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)

	if err != nil {
		// Try to at least get whatever HTML we have
		if htmlContent != "" {
			return htmlContent, fmt.Errorf("error rendering page with chromedp: %w", err)
		}
		return "", fmt.Errorf("error rendering page with chromedp: %w", err)
	}

	return htmlContent, nil
}

// LoadRenderedHTMLWithSelector fetches page and waits for a specific selector to appear
// This is useful for pages that lazy-load content and we need to wait for specific elements
func (pl *PageLoader) LoadRenderedHTMLWithSelector(ctx context.Context, url string, selector string) (string, error) {
	var htmlContent string

	// Create a timeout context for the browser operations
	browserCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	// Create exec allocator to launch Chrome with headless mode
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		browserCtx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("no-first-run", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	// Create a browser context
	browserContext, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// Build chromedp tasks
	tasks := []chromedp.Action{
		chromedp.Navigate(url),
		chromedp.WaitReady(`body`, chromedp.ByQuery),
	}

	// If selector is provided, wait for it to appear
	if selector != "" {
		tasks = append(tasks, chromedp.WaitVisible(selector, chromedp.ByQuery))
	}

	// Add sleep to ensure any lazy-loaded content is rendered
	tasks = append(tasks, chromedp.Sleep(2*time.Second))

	// Add the final HTML capture
	tasks = append(tasks, chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery))

	err := chromedp.Run(browserContext, tasks...)
	if err != nil {
		// Try to at least get whatever HTML we have
		if htmlContent != "" {
			return htmlContent, fmt.Errorf("error rendering page: %w", err)
		}
		return "", fmt.Errorf("error rendering page: %w", err)
	}

	return htmlContent, nil
}

// ExtractSection extracts an HTML section by tag name (for simple pages)
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
