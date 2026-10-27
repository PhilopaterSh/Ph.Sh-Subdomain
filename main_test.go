package main

import "testing"

func TestIsValidSubdomain(t *testing.T) {
	valid := []string{
		"www.example.com",
		"example.com",
		"_dmarc.example.com",
		"_domainkey.mail.example.com",
		"a.b.example.com",
		"sub-domain.example.com",
		"xn--p1ai.example.com",
		"example.xn--p1ai",
		"*.example.com",
	}
	for _, s := range valid {
		if !isValidSubdomain(s) {
			t.Errorf("expected valid subdomain: %q", s)
		}
	}

	invalid := []string{
		"",
		"not a domain",
		"example",
		"-bad.example.com",
		"ex..ample.com",
	}
	for _, s := range invalid {
		if isValidSubdomain(s) {
			t.Errorf("expected invalid subdomain: %q", s)
		}
	}
}

func TestCleanDomainLine(t *testing.T) {
	cases := map[string]string{
		"example.com":           "example.com",
		"  example.com  ":       "example.com",
		"example.com/extra":     "example.com",
		"a.com,b.com":           "a.com",
		"example.com.":          "example.com",
		"https://example.com/x": "https", // cut at ':' — "https" then fails isValidSubdomain and is dropped
		"":                      "",
		",,,":                   "",
	}
	for input, want := range cases {
		if got := cleanDomainLine(input); got != want {
			t.Errorf("cleanDomainLine(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSubdomainRegexMultiLabel(t *testing.T) {
	re := subdomainRegex("example.com")

	html := `Visit a.b.example.com and www.example.com today`
	matches := re.FindAllString(html, -1)
	want := []string{"a.b.example.com", "www.example.com"}
	if len(matches) != len(want) {
		t.Fatalf("got %v, want %v", matches, want)
	}
	for i := range want {
		if matches[i] != want[i] {
			t.Errorf("match[%d] = %q, want %q", i, matches[i], want[i])
		}
	}

	if matches := re.FindAllString("xexample.com", -1); len(matches) != 0 {
		t.Errorf("pattern should not match bare suffix: %v", matches)
	}
}

func TestIsRealAPIKey(t *testing.T) {
	for _, key := range []string{"", "  ", "YOUR_VT_API_KEY", "YOUR_URLSCAN_API_KEY"} {
		if isRealAPIKey(key) {
			t.Errorf("expected placeholder to be rejected: %q", key)
		}
	}
	if !isRealAPIKey("real-key-123") {
		t.Error("expected real key to be accepted")
	}
}
