package main

import (
	"fmt"
	"net/http"
)

// DNSDumpsterResponse represents the JSON response from api.dnsdumpster.com.
type DNSDumpsterResponse struct {
	A []struct {
		Host string `json:"host"`
	} `json:"a"`
	MX []struct {
		Host string `json:"host"`
	} `json:"mx"`
	CNAME []struct {
		Host string `json:"host"`
	} `json:"cname"`
}

// DNSDumpsterEngine is an engine that uses api.dnsdumpster.com to find subdomains.
type DNSDumpsterEngine struct {
	APIKey string
}

// Name returns the name of the engine.
func (e *DNSDumpsterEngine) Name() string {
	return "dnsdumpster"
}

// Fetch returns a list of subdomains found by the engine.
func (e *DNSDumpsterEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	if e.APIKey == "" {
		return nil, fmt.Errorf("dnsdumpster API key not provided")
	}

	url := fmt.Sprintf("https://api.dnsdumpster.com/domain/%s", domain)

	var response DNSDumpsterResponse
	if err := fetchJSON(client, url, map[string]string{"X-API-Key": e.APIKey}, &response); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, record := range response.A {
		subdomains = append(subdomains, record.Host)
	}
	for _, record := range response.MX {
		subdomains = append(subdomains, record.Host)
	}
	for _, record := range response.CNAME {
		subdomains = append(subdomains, record.Host)
	}

	return subdomains, nil
}
