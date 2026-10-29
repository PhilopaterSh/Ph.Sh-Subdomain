package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// fetchJSON performs a GET request against url (attaching the shared
// search-engine User-Agent plus any extra headers) and decodes the JSON
// response body into out. Any non-200 status becomes a descriptive error.
// All JSON-API engines share this path so status handling and header setup
// stay consistent.
func fetchJSON(client *http.Client, url string, headers map[string]string, out interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", searchEngineUserAgent)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
