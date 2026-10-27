// Command Ph.Sh_Sub (SubHunter-Go) is a fast, concurrent passive subdomain enumeration tool that
// aggregates results from certificate-transparency logs, search engines, and threat-intelligence
// APIs, then cleans and deduplicates them into a single sorted list.
package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Engine is the interface for a subdomain enumeration engine.
type Engine interface {
	// Fetch returns a list of subdomains found by the engine.
	Fetch(domain string, client *http.Client) ([]string, error)
	// Name returns the name of the engine.
	Name() string
}

// Result holds the outcome of a single engine's execution.
type Result struct {
	EngineName string
	Subdomains []string
	Error      error
}

// isValidSubdomain checks if a string is a valid subdomain.
// Allows underscores in labels (e.g. _dmarc.example.com), wildcard entries,
// embedded hyphens such as "xn--" punycode labels, and IDN punycode TLDs
// (e.g. example.xn--p1ai).
var isValidSubdomain = regexp.MustCompile(`(?i)^([a-z0-9\*_][a-z0-9\*_-]*\.)+([a-z]{2,}|xn--[a-z0-9-]+)$`).MatchString

// cleanAndUniqueSubdomains processes a list of subdomains to make them unique and valid.
func cleanAndUniqueSubdomains(subdomains []string) []string {
	uniqueSubdomains := make(map[string]bool)
	for _, sub := range subdomains {
		sub = strings.ToLower(sub)
		sub = strings.TrimSpace(sub)
		if strings.HasPrefix(sub, "*.") {
			sub = sub[2:]
		}
		if isValidSubdomain(sub) {

			uniqueSubdomains[sub] = true
		}
	}

	var cleaned []string
	for sub := range uniqueSubdomains {
		cleaned = append(cleaned, sub)
	}
	sort.Strings(cleaned)
	return cleaned
}

// cleanDomainLine extracts the domain candidate from a line in the domain list
// file. It keeps only the leading run of domain-legal characters, so input like
// "example.com,other.com" or "example.com/extra" never fuses into a bogus domain.
func cleanDomainLine(line string) string {
	line = strings.TrimSpace(line)
	for i, r := range line {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' {
			continue
		}
		line = line[:i]
		break
	}
	return strings.Trim(line, ".-")
}

// showAsciiArt prints the tool's startup banner to stdout.
func showAsciiArt() {
	fmt.Println(`
 ____   ____
|  _ \ / ___|
| |_) |\___ \
|  __/  ___) |
|_|    |____/

Built by: PhilopaterSh
# LinkedIn: https://www.linkedin.com/in/philopater-shenouda/`)
}

// main parses CLI flags, runs the selected engines concurrently for each target domain, and
// prints/saves the resulting unique subdomains.
func main() {
	showAsciiArt()
	// Define command-line flags
	domain := flag.String("d", "", "Target domain (e.g., example.com)")
	domainListFile := flag.String("dl", "", "File containing a list of domains to scan")
	outputFile := flag.String("o", "", "Save unique subdomains to file")
	verbose := flag.Bool("v", false, "Show the engine that found each subdomain")
	selectedEngines := flag.String("e", "", "Comma-separated engines to run (e.g., crtsh,digger)")
	noSslVerify := flag.Bool("no-ssl-verify", false, "Disable SSL verification (insecure)")
	threads := flag.Int("t", 10, "Number of concurrent threads/goroutines")
	bruteforce := flag.Bool("bruteforce", false, "Enable DNS bruteforce")
	wordlistFile := flag.String("wordlist", "", "Wordlist file for bruteforce")
	resolversFile := flag.String("resolvers", "", "File with DNS resolvers for bruteforce")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg, err := InitConfig()
	if err != nil {
		fmt.Printf("[-] Failed to load configuration: %v\n", err)
		cfg = &Config{}
	}

	var domains []string
	if *domainListFile != "" {
		file, err := os.Open(*domainListFile)
		if err != nil {
			fmt.Printf("[-] Failed to open domain list file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			cleanedLine := cleanDomainLine(scanner.Text())
			if cleanedLine != "" {
				domains = append(domains, cleanedLine)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("[-] Failed to read domain list file: %v\n", err)
			os.Exit(1)
		}
	}

	if *domain != "" {
		domains = append(domains, *domain)
	}

	if len(domains) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Create a shared HTTP client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: *noSslVerify},
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second, // Generous timeout for all engines
	}

	// Create a list of engines
	engines := []Engine{
		&CrtShEngine{},
		&AlienVaultEngine{},
		&DiggerToolsEngine{},
		&ShrewdEyeEngine{},
		&GoogleEngine{},
		&YahooEngine{},
		&BingEngine{},
		&BaiduEngine{},
		&AskEngine{},
		&NetcraftEngine{},
	}

	// API-key engines are only enabled when a real key is configured
	// (placeholder values such as "YOUR_..." written by the config template
	// must not activate the engines and fail on every run).
	if isRealAPIKey(cfg.APIKeys.Urlscan) {
		engines = append(engines, &UrlScanEngine{APIKey: cfg.APIKeys.Urlscan})
	}
	if isRealAPIKey(cfg.APIKeys.DNSDumpster) {
		engines = append(engines, &DNSDumpsterEngine{APIKey: cfg.APIKeys.DNSDumpster})
	}
	if isRealAPIKey(cfg.APIKeys.VT) {
		engines = append(engines, &VirusTotalEngine{APIKey: cfg.APIKeys.VT})
	}
	if isRealAPIKey(cfg.APIKeys.SecurityTrails) {
		engines = append(engines, &SecurityTrailsEngine{APIKey: cfg.APIKeys.SecurityTrails})
	}
	if isRealAPIKey(cfg.APIKeys.Shodan) {
		engines = append(engines, &ShodanEngine{APIKey: cfg.APIKeys.Shodan})
	}

	// Filter engines if the user provided a specific list
	if *selectedEngines != "" {
		userSelection := strings.Split(*selectedEngines, ",")
		selectedMap := make(map[string]bool)
		knownEngines := make(map[string]bool)
		for _, engine := range engines {
			knownEngines[engine.Name()] = true
		}
		for _, e := range userSelection {
			name := strings.TrimSpace(e)
			if name == "" {
				continue
			}
			if knownEngines[name] {
				selectedMap[name] = true
			} else {
				fmt.Printf("[!] Unknown engine '%s' — skipping (available: ", name)
				var names []string
				for _, engine := range engines {
					names = append(names, engine.Name())
				}
				fmt.Println(strings.Join(names, ", ") + ")")
			}
		}

		var filteredEngines []Engine
		for _, engine := range engines {
			if selectedMap[engine.Name()] {
				filteredEngines = append(filteredEngines, engine)
			}
		}
		engines = filteredEngines
		if len(engines) == 0 {
			fmt.Println("[-] No valid engines selected — nothing to run.")
			os.Exit(1)
		}
	}

	allSubdomains := make(map[string]bool)
	engineResults := make(map[string][]string)

	for _, currentDomain := range domains {
		fmt.Printf("[*] Starting subdomain enumeration for: %s\n", currentDomain)

		resultsChan := make(chan Result)
		var wg sync.WaitGroup
		// Use a semaphore to limit concurrency
		sem := make(chan struct{}, *threads)

		// Run engines concurrently
		for _, engine := range engines {
			wg.Add(1)
			go func(engine Engine) {
				defer wg.Done()
				sem <- struct{}{}        // Acquire a token
				defer func() { <-sem }() // Release the token
				subdomains, err := engine.Fetch(currentDomain, httpClient)
				resultsChan <- Result{
					EngineName: engine.Name(),
					Subdomains: subdomains,
					Error:      err,
				}
			}(engine)
		}

		// Closer goroutine
		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		// --- Collect all engine results ---
		for result := range resultsChan {
			if result.Error != nil {
				fmt.Printf("[-] Error fetching subdomains from %s: %v\n", result.EngineName, result.Error)
				continue
			}
			fmt.Printf("[%s] done: %d subs\n", result.EngineName, len(result.Subdomains))

			cleanedSubs := cleanAndUniqueSubdomains(result.Subdomains)
			engineResults[result.EngineName] = append(engineResults[result.EngineName], cleanedSubs...)
			for _, sub := range cleanedSubs {
				allSubdomains[sub] = true
			}
		}

		// --- Optional Bruteforce Step ---
		if *bruteforce {
			bruteResults := Bruteforce(currentDomain, *wordlistFile, *resolversFile, *threads)
			engineResults["bruteforce"] = append(engineResults["bruteforce"], bruteResults...)
			for _, sub := range bruteResults {
				allSubdomains[sub] = true
			}
		}
	}

	// Create a final sorted slice of unique subdomains
	var finalSubdomains []string
	for sub := range allSubdomains {
		finalSubdomains = append(finalSubdomains, sub)
	}
	sort.Strings(finalSubdomains)

	// --- Output Results ---
	fmt.Printf("\nTotal Unique Subdomains: %d\n\n", len(finalSubdomains))

	if *verbose {
		// Print output grouped by engine
		for engineName, subs := range engineResults {
			if len(subs) > 0 {
				sort.Strings(subs)
				fmt.Printf("--- %s (%d) ---\n", engineName, len(subs))
				for _, sub := range subs {
					fmt.Println(sub)
				}
				fmt.Println() // Add a newline for spacing
			}
		}
	} else {
		// Print unique list only
		for _, sub := range finalSubdomains {
			fmt.Println(sub)
		}
	}

	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Printf("\n[!] Failed to save file: %v\n", err)
		} else {
			defer file.Close()
			for _, sub := range finalSubdomains {
				_, _ = file.WriteString(sub + "\n")
			}
			fmt.Printf("\n[+] Saved to: %s\n", *outputFile)
		}
	}

	fmt.Println("\n[*] Enumeration finished.")

}
