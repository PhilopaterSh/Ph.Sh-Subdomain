package main

import (
	"fmt"
	"net/http"
)

// AlienVaultResponse represents the JSON response from AlienVault OTX.
type AlienVaultResponse struct {
	PassiveDNS []struct {
		Hostname string `json:"hostname"`
	} `json:"passive_dns"`
}

// AlienVaultEngine is an engine that uses AlienVault OTX to find subdomains.
type AlienVaultEngine struct{}

// Name returns the name of the engine.
func (e *AlienVaultEngine) Name() string {
	return "alienvault"
}

// Fetch returns a list of subdomains found by the engine.
func (e *AlienVaultEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)

	var response AlienVaultResponse
	if err := fetchJSON(client, url, nil, &response); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, record := range response.PassiveDNS {
		subdomains = append(subdomains, record.Hostname)
	}

	return subdomains, nil
}
