package main

import (
	"fmt"
	"net/http"
	"net/url"
)

// AskEngine is an engine that scrapes subdomains from Ask.com search results.
type AskEngine struct{}

// Name returns the name of the engine.
func (e *AskEngine) Name() string {
	return "ask"
}

// Fetch retrieves subdomains from Ask.com.
func (e *AskEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	// Ask.com paginates with the 'page' parameter
	subdomains := scrapeSearchEnginePages(client, domain, 5, func(page int) string {
		query := fmt.Sprintf("site:%s", domain)
		return fmt.Sprintf("https://www.ask.com/web?q=%s&page=%d", url.QueryEscape(query), page+1)
	})

	return subdomains, nil
}
