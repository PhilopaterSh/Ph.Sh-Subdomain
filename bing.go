package main

import (
	"fmt"
	"net/http"
	"strings"
)

// BingEngine is an engine that scrapes subdomains from Bing search results.
type BingEngine struct{}

// Name returns the name of the engine.
func (e *BingEngine) Name() string {
	return "bing"
}

// Fetch retrieves subdomains from Bing.
func (e *BingEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	// Bing paginates with the 'first' parameter (1, 11, 21, etc.)
	matches := scrapeSearchEnginePages(client, domain, 5, func(page int) string {
		return fmt.Sprintf("https://www.bing.com/search?q=domain%%3A%s&first=%d", domain, page*10+1)
	})

	var subdomains []string
	for _, match := range matches {
		if !strings.HasPrefix(match, "www.") { // Simple filter
			subdomains = append(subdomains, match)
		}
	}

	return subdomains, nil
}
