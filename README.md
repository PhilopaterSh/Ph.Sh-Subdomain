# SubHunter-Go

<p align="center">
  <img src="Ph.Sh-Sudbomain.png" alt="SubHunter-Go Logo" width="150">
</p>

<h3 align="center">A Fast and Concurrent Subdomain Enumeration Tool</h3>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/PhilopaterSh/SubHunter-Ph.Sh"><img src="https://goreportcard.com/badge/github.com/PhilopaterSh/SubHunter-Ph.Sh" alt="Go Report Card"></a>
  <a href="https://github.com/PhilopaterSh/SubHunter-Ph.Sh/releases"><img src="https://img.shields.io/github/v/release/PhilopaterSh/SubHunter-Ph.Sh" alt="GitHub release"></a>
</p>

---

SubHunter-Go is a fast and effective subdomain enumeration tool written in Go. It is designed to discover subdomains using a variety of passive sources, leveraging Go's concurrency for maximum speed and efficiency.

## Features

- **Concurrent Engine Execution**: Runs all engines in parallel for maximum speed.
- **Multiple Subdomain Sources**: Gathers results from the most reliable passive sources.
- **Flexible Input**: Supports scanning a single domain with `-d` or multiple domains from a file with `-dl`.
- **Clean & Unique Output**: Automatically processes, cleans, and de-duplicates results.
- **Cross-Platform**: Can be compiled to run on Windows, Linux, and macOS.

### Supported Engines

- AlienVault OTX
- Ask
- Baidu
- Bing
- Crt.sh
- Digger.tools
- DNSDumpster
- Google
- Netcraft
- SecurityTrails
- Shodan
- ShrewdEye
- **ThreatCrowd** (Restored)
- Urlscan.io
- VirusTotal
- Yahoo

## Requirements

- **Go**: Version 1.18 or newer (only required for installation or building from source).
- **Python**: The `digger` engine requires Python and the `cloudscraper` library to be installed (`pip install cloudscraper`).

## Installation

The recommended way to install is using `go install`, which makes the command available globally from your terminal.

```sh
go install -v github.com/PhilopaterSh/Ph.Sh_Sub@latest
```

**Note:** This command will install the binary (named `Ph.Sh_Sub` or `Ph.Sh_Sub.exe`) into your Go binary path (e.g., `$GOPATH/bin` or `$HOME/go/bin`). Ensure this directory is in your system's `PATH` to run the tool from anywhere.

## Usage

Here are some examples of how to use SubHunter-Go:

*   **Scan a single domain:**
    ```sh
    Ph.Sh_Sub -d example.com
    ```

*   **Scan a list of domains from a file:**
    ```sh
    Ph.Sh_Sub -dl domains.txt
    ```

*   **Save results to a file:**
    ```sh
    Ph.Sh_Sub -d example.com -o results.txt
    ```

*   **Use verbose mode to see which engine found each subdomain:**
    ```sh
    Ph.Sh_Sub -d example.com -v
    ```
    
*   **Handle SSL Certificate Errors:**
    Some servers (like ThreatCrowd) may have invalid SSL certificates. Use the `--no-ssl-verify` flag to bypass these errors.
    ```sh
    Ph.Sh_Sub -d example.com --no-ssl-verify
    ```

### All Options

```
Usage of SubHunter-Go:
  -d string
        Target domain (e.g., example.com)
  -dl string
        File containing a list of domains to scan
  -e string
        Comma-separated engines to run (e.g., crtsh,digger)
  -no-ssl-verify
        Disable SSL verification (insecure)
  -o string
        Save unique subdomains to file
  -t int
        Number of concurrent threads/goroutines (default 10)
  -v
        Show the engine that found each subdomain
  -bruteforce
        Enable DNS bruteforce
  -wordlist string
        Wordlist file for bruteforce
  -resolvers string
        File with DNS resolvers for bruteforce
```

## API Keys

Some engines require API keys to function. The program reads these keys from **environment variables**.

**IMPORTANT:** Do NOT edit the `config.go` file. Set the following environment variables before running the tool.

```sh
# Example for Linux/macOS (.zshrc, .bashrc)
export URLSCAN_API_KEY="YOUR_URLSCAN_KEY"
export DNSDUMPSTER_API_KEY="YOUR_DNSDUMPSTER_KEY"
export VT_API_KEY="YOUR_VT_API_KEY"
export SECURITYTRAILS_API_KEY="YOUR_SECURITYTRAILS_KEY"
export SHODAN_API_KEY="YOUR_SHODAN_KEY"

# Example for Windows (Command Prompt)
set URLSCAN_API_KEY="YOUR_URLSCAN_KEY"
set VT_API_KEY="YOUR_VT_API_KEY"
```
