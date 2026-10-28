package main

import (
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

	var response ShrewdEyeResponse
	if err := fetchJSON(client, url, nil, &response); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, record := range response.Data {
		subdomains = append(subdomains, record.Name)
	}

	return subdomains, nil
}
