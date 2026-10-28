package main

import (
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

	var stResponse SecurityTrailsResponse
	if err := fetchJSON(client, url, map[string]string{"APIKEY": e.APIKey}, &stResponse); err != nil {
		return nil, err
	}

	// The subdomains from SecurityTrails are just the prefix, so we need to append the domain.
	fullSubdomains := make([]string, len(stResponse.Subdomains))
	for i, sub := range stResponse.Subdomains {
		fullSubdomains[i] = fmt.Sprintf("%s.%s", sub, domain)
	}

	return fullSubdomains, nil
}
