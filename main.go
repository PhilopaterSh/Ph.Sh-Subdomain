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
var isValidSubdomain = regexp.MustCompile(`(?i)^([a-z0-9\*]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`).MatchString

// cleanAndUniqueSubdomains processes a list of subdomains to make them unique and valid.
func cleanAndUniqueSubdomains(subdomains []string) []string {
	uniqueSubdomains := make(map[string]bool)
	for _, sub := range subdomains {
		sub = strings.ToLower(sub)
		sub = strings.TrimSpace(sub)
		if strings.HasPrefix(sub, "*.") {
			sub = sub[2:]
		}
		if strings.HasPrefix(sub, "0a") {
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

func showAsciiArt() {
	fmt.Println(`
  _____   _            _____ _     
 |  __ \ | |          / ____| |    
 | |__) || |__       | (___ | |__  
 |  ___/ | '_' \     \___ \| '_' \ 
 | |     | | | |  _  ____) | | | |
 |_|     |_| |_| (_) _____/|_| |_|
Built by : PhilopaterSh
# LinkedIn: https://www.linkedin.com/in/philopater-shenouda/
                              `)
}

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
			domains = append(domains, scanner.Text())
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
		&ThreatCrowdEngine{},
	}

	if apiKeys["urlscan"] != "" {
		engines = append(engines, &UrlScanEngine{APIKey: apiKeys["urlscan"]})
	}
	if apiKeys["dnsdumpster"] != "" {
		engines = append(engines, &DNSDumpsterEngine{APIKey: apiKeys["dnsdumpster"]})
	}
	if apiKeys["vt"] != "" {
		engines = append(engines, &VirusTotalEngine{APIKey: apiKeys["vt"]})
	}
	if apiKeys["securitytrails"] != "" {
		engines = append(engines, &SecurityTrailsEngine{APIKey: apiKeys["securitytrails"]})
	}
	if apiKeys["shodan"] != "" {
		engines = append(engines, &ShodanEngine{APIKey: apiKeys["shodan"]})
	}

	// Filter engines if the user provided a specific list
	if *selectedEngines != "" {
		userSelection := strings.Split(*selectedEngines, ",")
		selectedMap := make(map[string]bool)
		for _, e := range userSelection {
			selectedMap[strings.TrimSpace(e)] = true
		}

		var filteredEngines []Engine
		for _, engine := range engines {
			if selectedMap[engine.Name()] {
				filteredEngines = append(filteredEngines, engine)
			}
		}
		engines = filteredEngines
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
