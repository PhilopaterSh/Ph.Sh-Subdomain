package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ShrewdEyeResponse represents the JSON response from shrewdeye.app.
type ShrewdEyeResponse struct {
	Data []struct {
		Name string `json:"name"`
	} `json:"data"`
}

// ShrewdEyeEngine is an engine that uses shrewdeye.app to find subdomains.
type ShrewdEyeEngine struct{}

// Name returns the name of the engine.
func (e *ShrewdEyeEngine) Name() string {
	return "shrewdeye"
}

// Fetch returns a list of subdomains found by the engine.
func (e *ShrewdEyeEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	url := fmt.Sprintf("https://shrewdeye.app/api/v1/domains/%s/resources", domain)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shrewdeye.app returned status: %s", resp.Status)
	}

	var response ShrewdEyeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, record := range response.Data {
		subdomains = append(subdomains, record.Name)
	}

	return subdomains, nil
}
