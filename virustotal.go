package main

import (
	"fmt"
	"net/http"
)

// VirusTotalEngine searches for subdomains using the VirusTotal API.
type VirusTotalEngine struct {
	APIKey string
}

// Name returns the name of the engine.
func (e *VirusTotalEngine) Name() string {
	return "vt"
}

// VirusTotalResponse represents the structure of the JSON response from VT.
type VirusTotalResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// Fetch retrieves subdomains from VirusTotal.
func (e *VirusTotalEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	if e.APIKey == "" {
		return nil, fmt.Errorf("API key for VirusTotal is not set")
	}

	url := fmt.Sprintf("https://www.virustotal.com/api/v3/domains/%s/subdomains?limit=40", domain)

	var vtResponse VirusTotalResponse
	if err := fetchJSON(client, url, map[string]string{"x-apikey": e.APIKey}, &vtResponse); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, item := range vtResponse.Data {
		subdomains = append(subdomains, item.ID)
	}

	return subdomains, nil
}
