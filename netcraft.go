package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// NetcraftEngine is an engine that scrapes subdomains from Netcraft.
type NetcraftEngine struct{}

// Name returns the name of the engine.
func (e *NetcraftEngine) Name() string {
	return "netcraft"
}

// Fetch retrieves subdomains from Netcraft.
func (e *NetcraftEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	var subdomains []string
	nextURL := fmt.Sprintf("https://searchdns.netcraft.com/?restriction=site+ends+with&host=%s", domain)

	// Regex to find the 'Next Page' link
	nextPageRegex := regexp.MustCompile(`<a href="([^"]+)"><b>Next page</b></a>`)

	for i := 0; i < 10 && nextURL != ""; i++ { // Limit to 10 pages to prevent infinite loops
		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			break // Stop if there's an error
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			break
		}

		bodyStr := string(body)

		// Regex to find subdomains
		re := regexp.MustCompile(fmt.Sprintf(`([a-zA-Z0-9_-]+\.%s)`, regexp.QuoteMeta(domain)))
		matches := re.FindAllString(bodyStr, -1)
		subdomains = append(subdomains, matches...)

		// Find the next page URL
		match := nextPageRegex.FindStringSubmatch(bodyStr)
		if len(match) > 1 {
			nextURL = "https://searchdns.netcraft.com" + match[1]
			// Netcraft links can have HTML entities, clean them
			nextURL = strings.ReplaceAll(nextURL, "&amp;", "&")
		} else {
			nextURL = "" // No more pages
		}
	}

	return subdomains, nil
}
