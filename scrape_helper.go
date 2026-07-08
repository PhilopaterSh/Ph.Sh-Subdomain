package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// searchEngineUserAgent is the User-Agent header used by the search-engine scraping engines.
const searchEngineUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// scrapeSearchEnginePages fetches the given number of search-result pages, calling urlForPage
// to build the URL for each page index (0-based), and extracts subdomains of domain from the
// returned HTML. A single page's request/read error is skipped rather than aborting the scrape.
func scrapeSearchEnginePages(client *http.Client, domain string, pages int, urlForPage func(page int) string) []string {
	var subdomains []string
	re := regexp.MustCompile(fmt.Sprintf(`([a-zA-Z0-9_-]+\.%s)`, regexp.QuoteMeta(domain)))

	for i := 0; i < pages; i++ {
		req, err := http.NewRequest("GET", urlForPage(i), nil)
		if err != nil {
			return subdomains
		}
		req.Header.Set("User-Agent", searchEngineUserAgent)

		resp, err := client.Do(req)
		if err != nil {
			continue // Don't stop for a single page error
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		subdomains = append(subdomains, re.FindAllString(string(body), -1)...)
	}

	return subdomains
}
