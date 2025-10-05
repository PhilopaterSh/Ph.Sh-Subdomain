package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ThreatCrowdEngine is an engine that queries ThreatCrowd.
type ThreatCrowdEngine struct{}

// Name returns the name of the engine.
func (e *ThreatCrowdEngine) Name() string {
	return "threatcrowd"
}

// Fetch returns a list of subdomains found by the engine.
func (e *ThreatCrowdEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	// Create a custom transport and client for ThreatCrowd to always ignore SSL errors,
	// as its certificate is consistently misconfigured.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	customClient := &http.Client{
		Transport: tr,
		Timeout:   client.Timeout, // Use the same timeout as the shared client
	}

	url := fmt.Sprintf("https://ci-www.threatcrowd.org/searchApi/v2/domain/report/?domain=%s", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := customClient.Do(req) // Use the custom client
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var data struct {
		Subdomains []string `json:"subdomains"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		// If JSON parsing fails, it might be an empty response or an error message.
		// We can return an empty slice instead of an error.
		return []string{}, nil
	}

	return data.Subdomains, nil
}
