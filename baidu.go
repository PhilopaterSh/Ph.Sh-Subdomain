package main

import (
	"fmt"
	"net/http"
	"net/url"
)

// BaiduEngine is an engine that scrapes subdomains from Baidu search results.
type BaiduEngine struct{}

// Name returns the name of the engine.
func (e *BaiduEngine) Name() string {
	return "baidu"
}

// Fetch retrieves subdomains from Baidu.
func (e *BaiduEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	// Baidu paginates with the 'pn' parameter (page number * 10)
	subdomains := scrapeSearchEnginePages(client, domain, 5, func(page int) string {
		query := fmt.Sprintf("site:%s", domain)
		return fmt.Sprintf("https://www.baidu.com/s?pn=%d&wd=%s", page*10, url.QueryEscape(query))
	})

	return subdomains, nil
}
