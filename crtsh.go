package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CrtShEntry represents a single entry from the crt.sh JSON output.
type CrtShEntry struct {
	NameValue string `json:"name_value"`
}

// CrtShEngine is an engine that uses crt.sh to find subdomains.
type CrtShEngine struct{}

// Name returns the name of the engine.
func (e *CrtShEngine) Name() string {
	return "crtsh"
}

// Fetch returns a list of subdomains found by the engine.
func (e *CrtShEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	var subdomains []string
	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

	time.Sleep(1 * time.Second) // Be polite

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crt.sh returned status: %s", resp.Status)
	}

	var entries []CrtShEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	for _, entry := range entries {
		subdomains = append(subdomains, entry.NameValue)
	}

	return subdomains, nil
}