package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

// BaiduEngine is an engine that scrapes subdomains from Baidu search results.
type BaiduEngine struct{}

// Name returns the name of the engine.
func (e *BaiduEngine) Name() string {
	return "baidu"
}

// Fetch retrieves subdomains from Baidu.
func (e *BaiduEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	var subdomains []string
	// Baidu paginates with the 'pn' parameter (page number * 10)
	for i := 0; i < 5; i++ {
		query := fmt.Sprintf("site:%s", domain)
		searchURL := fmt.Sprintf("https://www.baidu.com/s?pn=%d&wd=%s", i*10, url.QueryEscape(query))

		req, err := http.NewRequest("GET", searchURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			continue // Don't stop for a single page error
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
		subdomains = append(subdomains, matches...)
	}

	return subdomains, nil
}
