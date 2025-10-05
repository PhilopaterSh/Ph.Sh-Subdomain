package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
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
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("yahoo returned status: %s", resp.Status)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		re := regexp.MustCompile(fmt.Sprintf(`([a-zA-Z0-9_-]+\.%s)`, regexp.QuoteMeta(domain)))
		matches := re.FindAllString(string(body), -1)
		subdomains = append(subdomains, matches...)

		time.Sleep(600 * time.Millisecond) // Sleep to be polite
	}

	return subdomains, nil
}