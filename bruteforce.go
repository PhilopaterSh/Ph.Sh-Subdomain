package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// defaultWordlist provides a small list of common subdomains for bruteforcing.
var defaultWordlist = []string{
	"www", "mail", "dev", "api", "test", "staging", "m", "admin", "portal", "vpn",
	"shop", "blog", "ftp", "webmail", "remote", "smtp", "ns1", "ns2", "cpanel",
	"whm", "autodiscover", "owa", "support", "docs", "git", "svn", "status",
}

// BruteforceResult holds a found subdomain and its IP.
type BruteforceResult struct {
	Subdomain string
	IP        string
}

// loadWordlist reads a wordlist from a file.
func loadWordlist(path string) ([]string, error) {
	if path == "" {
		fmt.Println("[*] No wordlist file provided, using default list.")
		return defaultWordlist, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open wordlist file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

// loadResolvers reads a list of DNS resolvers from a file.
func loadResolvers(path string) ([]string, error) {
	if path == "" {
		return nil, nil // Use system default
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open resolvers file: %w", err)
	}
	defer file.Close()

	var resolvers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			// Basic validation, could be improved
			if net.ParseIP(line) != nil {
				resolvers = append(resolvers, line+":53") // Append port
			}
		}
	}
	return resolvers, scanner.Err()
}

// Bruteforce performs DNS bruteforcing for a given domain.
func Bruteforce(domain, wordlistFile, resolversFile string, threads int) []string {
	wordlist, err := loadWordlist(wordlistFile)
	if err != nil {
		fmt.Printf("[-] Error loading wordlist: %v\n", err)
		return nil
	}

	resolvers, err := loadResolvers(resolversFile)
	if err != nil {
		fmt.Printf("[-] Error loading resolvers: %v\n", err)
		// Continue with system resolvers
	}

	fmt.Printf("[*] Starting bruteforce with %d words and %d threads.\n", len(wordlist), threads)

	var foundSubdomains []string
	var wg sync.WaitGroup
	subdomainChan := make(chan string, len(wordlist))
	resultsChan := make(chan BruteforceResult, len(wordlist))

	// Create a custom resolver if custom resolvers are provided
	var customResolver *net.Resolver
	if len(resolvers) > 0 {
		customResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				// Round-robin through custom resolvers
				resolverAddr := resolvers[0] // Simple approach, can be improved
				return d.DialContext(ctx, "udp", resolverAddr)
			},
		}
	}

	// Start worker goroutines
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for sub := range subdomainChan {
				host := sub + "." + domain
				var ips []string
				var err error
				if customResolver != nil {
					ips, err = customResolver.LookupHost(context.Background(), host)
				} else {
					ips, err = net.LookupHost(host)
				}

				if err == nil && len(ips) > 0 {
					resultsChan <- BruteforceResult{Subdomain: host, IP: ips[0]}
				}
			}
		}()
	}

	// Feed the subdomains to the workers
	for _, word := range wordlist {
		subdomainChan <- word
	}
	close(subdomainChan)

	// Closer goroutine for results
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		fmt.Printf("[BRUTE] Found: %s -> %s\n", result.Subdomain, result.IP)
		foundSubdomains = append(foundSubdomains, result.Subdomain)
	}

	fmt.Printf("[*] Bruteforce finished. Found %d subdomains.\n", len(foundSubdomains))
	return foundSubdomains
}
