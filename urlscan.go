package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// UrlScanResponse represents the JSON response from urlscan.io.
type UrlScanResponse struct {
	Results []struct {
		Page struct {
			Domain string `json:"domain"`
		} `json:"page"`
	} `json:"results"`
}

// UrlScanEngine is an engine that uses urlscan.io to find subdomains.
type UrlScanEngine struct {
	APIKey string
}

// Name returns the name of the engine.
func (e *UrlScanEngine) Name() string {
	return "urlscan"
}

// Fetch returns a list of subdomains found by the engine.
func (e *UrlScanEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	if e.APIKey == "" {
		return nil, fmt.Errorf("urlscan.io API key not provided")
	}

	url := fmt.Sprintf("https://urlscan.io/api/v1/search/?q=domain:%s", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("API-Key", e.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("urlscan.io returned status: %s", resp.Status)
	}

	var response UrlScanResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, result := range response.Results {
		subdomains = append(subdomains, result.Page.Domain)
	}

	return subdomains, nil
}