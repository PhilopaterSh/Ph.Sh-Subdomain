package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// searchEngineUserAgent is the User-Agent header used by the search-engine scraping engines.
const searchEngineUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// subdomainRegex builds a pattern that matches any host ending with domain,
// including multi-label results such as a.b.example.com (a single-label
// pattern would silently reduce them to b.example.com).
func subdomainRegex(domain string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`([a-zA-Z0-9_-]+(?:\.[a-zA-Z0-9_-]+)*\.%s)`, regexp.QuoteMeta(domain)))
}

// scrapeSearchEnginePages fetches the given number of search-result pages, calling urlForPage
// to build the URL for each page index (0-based), and extracts subdomains of domain from the
// returned HTML. A single page's request/read error is skipped rather than aborting the scrape.
func scrapeSearchEnginePages(client *http.Client, domain string, pages int, urlForPage func(page int) string) []string {
	var subdomains []string
	re := subdomainRegex(domain)

	for i := 0; i < pages; i++ {
		// Politeness delay between consecutive result pages — hitting a
		// search engine back-to-back is a quick way to get blocked.
		if i > 0 {
			time.Sleep(600 * time.Millisecond)
		}
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
		_ = resp.Body.Close()
		if err != nil {
			continue
		}

		subdomains = append(subdomains, re.FindAllString(string(body), -1)...)
	}

	return subdomains
}
