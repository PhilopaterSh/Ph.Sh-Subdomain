package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SecurityTrailsEngine searches for subdomains using the SecurityTrails API.
type SecurityTrailsEngine struct {
	APIKey string
}

// Name returns the name of the engine.
func (e *SecurityTrailsEngine) Name() string {
	return "securitytrails"
}

// SecurityTrailsResponse defines the structure for the API response.
type SecurityTrailsResponse struct {
	Subdomains []string `json:"subdomains"`
}

// Fetch retrieves subdomains from SecurityTrails.
func (e *SecurityTrailsEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	if e.APIKey == "" {
		return nil, fmt.Errorf("API key for SecurityTrails is not set")
	}

	url := fmt.Sprintf("https://api.securitytrails.com/v1/domain/%s/subdomains", domain)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("APIKEY", e.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var stResponse SecurityTrailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stResponse); err != nil {
		return nil, err
	}

	// The subdomains from SecurityTrails are just the prefix, so we need to append the domain.
	fullSubdomains := make([]string, len(stResponse.Subdomains))
	for i, sub := range stResponse.Subdomains {
		fullSubdomains[i] = fmt.Sprintf("%s.%s", sub, domain)
	}

	return fullSubdomains, nil
}
