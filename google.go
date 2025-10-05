package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
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
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("google returned status: %s", resp.Status)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if strings.Contains(strings.ToLower(string(body)), "unusual traffic") {
			return nil, fmt.Errorf("google blocked the request due to unusual traffic")
		}

		re := regexp.MustCompile(fmt.Sprintf(`([a-zA-Z0-9_-]+\.%s)`, regexp.QuoteMeta(domain)))
		matches := re.FindAllString(string(body), -1)
		subdomains = append(subdomains, matches...)

		time.Sleep(1500 * time.Millisecond) // Sleep to be polite
	}

	return subdomains, nil
}