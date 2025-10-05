package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
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
	var subdomains []string
	// Bing paginates with the 'first' parameter (1, 11, 21, etc.)
	for i := 0; i < 5; i++ {
		url := fmt.Sprintf("https://www.bing.com/search?q=domain%%3A%s&first=%d", domain, i*10+1)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			// Don't stop for a single page error
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		bodyStr := string(body)

		// Regex to find subdomains
		re := regexp.MustCompile(fmt.Sprintf(`([a-zA-Z0-9_-]+\.%s)`, regexp.QuoteMeta(domain)))
		matches := re.FindAllString(bodyStr, -1)
		for _, match := range matches {
			if !strings.HasPrefix(match, "www.") { // Simple filter
				subdomains = append(subdomains, match)
			}
		}
	}

	return subdomains, nil
}
