package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

// AskEngine is an engine that scrapes subdomains from Ask.com search results.
type AskEngine struct{}

// Name returns the name of the engine.
func (e *AskEngine) Name() string {
	return "ask"
}

// Fetch retrieves subdomains from Ask.com.
func (e *AskEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	var subdomains []string
	// Ask.com paginates with the 'page' parameter
	for i := 1; i <= 5; i++ {
		query := fmt.Sprintf("site:%s", domain)
		searchURL := fmt.Sprintf("https://www.ask.com/web?q=%s&page=%d", url.QueryEscape(query), i)

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
