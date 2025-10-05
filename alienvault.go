package main

import (
	"encoding/json"
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

	req, err := http.NewRequest("GET", url, nil)
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
		return nil, fmt.Errorf("alienvault returned status: %s", resp.Status)
	}

	var response AlienVaultResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, record := range response.PassiveDNS {
		subdomains = append(subdomains, record.Hostname)
	}

	return subdomains, nil
}