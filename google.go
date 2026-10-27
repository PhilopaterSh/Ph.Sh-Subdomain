package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GoogleEngine is an engine that uses Google to find subdomains.
type GoogleEngine struct{}

// Name returns the name of the engine.
func (e *GoogleEngine) Name() string {
	return "google"
}

// Fetch returns a list of subdomains found by the engine.
func (e *GoogleEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	var subdomains []string

	for start := 0; start <= 40; start += 10 {
		query := fmt.Sprintf("site:%s -www.%s", domain, domain)
		googleURL := fmt.Sprintf("https://www.google.com/search?q=%s&start=%d&gbv=1", url.QueryEscape(query), start)

		req, err := http.NewRequest("GET", googleURL, nil)
		if err != nil {
			return nil, err
		}
		// Google is picky about the User-Agent.
		req.Header.Set("User-Agent", searchEngineUserAgent)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("google returned status: %s", resp.Status)
		}

		if strings.Contains(strings.ToLower(string(body)), "unusual traffic") {
			return nil, fmt.Errorf("google blocked the request due to unusual traffic")
		}

		re := subdomainRegex(domain)
		matches := re.FindAllString(string(body), -1)
		subdomains = append(subdomains, matches...)

		time.Sleep(1500 * time.Millisecond) // Sleep to be polite
	}

	return subdomains, nil
}
