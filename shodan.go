package main

import (
	"fmt"
	"net/http"
)

// ShodanEngine searches for subdomains using the Shodan API.
type ShodanEngine struct {
	APIKey string
}

// Name returns the name of the engine.
func (e *ShodanEngine) Name() string {
	return "shodan"
}

// ShodanResponse defines the structure for the API response.
type ShodanResponse struct {
	Subdomains []string `json:"subdomains"`
}

// Fetch retrieves subdomains from Shodan.
func (e *ShodanEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	if e.APIKey == "" {
		return nil, fmt.Errorf("API key for Shodan is not set")
	}

	url := fmt.Sprintf("https://api.shodan.io/dns/domain/%s?key=%s", domain, e.APIKey)

	var shodanResponse ShodanResponse
	if err := fetchJSON(client, url, nil, &shodanResponse); err != nil {
		return nil, err
	}

	// The subdomains from Shodan are just the prefix, so we need to append the domain.
	fullSubdomains := make([]string, len(shodanResponse.Subdomains))
	for i, sub := range shodanResponse.Subdomains {
		fullSubdomains[i] = fmt.Sprintf("%s.%s", sub, domain)
	}

	return fullSubdomains, nil
}
