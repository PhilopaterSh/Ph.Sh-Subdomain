package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// YahooEngine is an engine that uses Yahoo to find subdomains.
type YahooEngine struct{}

// Name returns the name of the engine.
func (e *YahooEngine) Name() string {
	return "yahoo"
}

// Fetch returns a list of subdomains found by the engine.
func (e *YahooEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	var subdomains []string

	for p := 1; p <= 5; p++ {
		yahooURL := fmt.Sprintf("https://search.yahoo.com/search?p=site%%3A%s&b=%d", domain, (p-1)*10+1)

		req, err := http.NewRequest("GET", yahooURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", searchEngineUserAgent)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("yahoo returned status: %s", resp.Status)
		}

		re := subdomainRegex(domain)
		matches := re.FindAllString(string(body), -1)
		subdomains = append(subdomains, matches...)

		time.Sleep(600 * time.Millisecond) // Sleep to be polite
	}

	return subdomains, nil
}
