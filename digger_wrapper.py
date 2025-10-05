
import sys
import json
import re

try:
    import cloudscraper
except ImportError:
    sys.stderr.write("Error: cloudscraper library not found. Please install it with 'pip install cloudscraper'\n")
    sys.exit(1)

def get_digger_subdomains(domain):
    """
    Fetches subdomains for a given domain from digger.tools using cloudscraper.
    """
    subs = set()
    try:
        scraper = cloudscraper.create_scraper()
        url = f"https://digger.tools/lookup/{domain}/subdomains"
        r = scraper.get(url, timeout=30)

        if r.status_code != 200:
            sys.stderr.write(f"Digger.tools returned status: {r.status_code}\n")
            return list(subs)

        # Digger may return JSON or HTML, try JSON first
        try:
            data = r.json()
            for s in data.get("subdomains", []):
                if "." in s:
                    fq = s
                else:
                    fq = f"{s}.{domain}"
                subs.add(fq.lower())
        except ValueError:
            # Fallback to regex on HTML if JSON fails
            pat = re.compile(r"([a-zA-Z0-9_-](?:[a-zA-Z0-9_-]*\.)+%s)" % re.escape(domain))
            found = pat.findall(r.text)
            for sub in found:
                subs.add(sub.lower())
    except Exception as e:
        sys.stderr.write(f"An unexpected error occurred in digger_wrapper: {e}\n")
        pass
    
    return sorted(list(subs))

if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit(1)
    
    domain_to_check = sys.argv[1]
    subdomains = get_digger_subdomains(domain_to_check)
    for sub in subdomains:
        print(sub)
