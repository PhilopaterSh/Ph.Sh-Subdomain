# Phase 0 Research: Quality Parity with Modern Subdomain Enumeration Tools

## 1. Wildcard probe resolver choice

**Decision**: Reuse `bruteforce.go`'s existing pattern — if the user supplied `-resolvers`, probe
through that custom resolver; otherwise fall back to a small hardcoded set of trusted public
resolvers (8.8.8.8, 1.1.1.1).

**Rationale**: puredns and subfinder both validate wildcard behavior against trusted public
resolvers specifically to avoid DNS-poisoned or ISP-hijacked answers skewing the result. This project
already has the plumbing (`loadResolvers`, custom `net.Resolver` with a `Dial` override) — no new
dependency needed, just a second call site.

**Alternatives considered**: Using the OS default resolver only — rejected because it's exactly the
failure mode (poisoned/hijacked resolvers) the research flagged as a known false-positive source.

## 2. Random label generation

**Decision**: Generate 2 probe labels per target domain, each 20+ lowercase alphanumeric characters,
using `crypto/rand` (already implicitly available via stdlib, no new import burden beyond what's
already linked for `crypto/tls`).

**Rationale**: Two independent probes reduce the chance a single probe's answer is itself a coincidence;
20+ chars makes accidental collision with a real provisioned hostname astronomically unlikely, which
directly addresses the Edge Case called out in spec.md.

**Alternatives considered**: Single probe — rejected, more prone to one-off resolution flakiness
producing a false "not wildcarded" read. Timestamp-based labels — rejected, less entropy than random.

## 3. JSON schema shape

**Decision**:

```json
[
  {"subdomain": "dev.example.com", "sources": ["crtsh", "alienvault"], "resolved": true, "ip": "1.2.3.4"},
  {"subdomain": "old.example.com", "sources": ["crtsh"], "resolved": false}
]
```

`resolved`/`ip` keys are present only when `-resolve` was used for that run; `sources` always reflects
which engine(s) reported that name (already computable today via `engineResults`, just not currently
surfaced in JSON form).

**Rationale**: A flat JSON array is trivially consumable by `jq`, Python, or another Go program, and
matches the scale of this tool's output (hundreds to low thousands of entries per run) — line-delimited
JSON (JSONL) is more common at massdns/httpx scale (millions of lines) and would be premature
optimization here.

**Alternatives considered**: JSONL — rejected for now given current scale; can be added later as
`-jsonl` without breaking `-json` if usage patterns ever demand it.

## 4. New engine endpoints (all keyless, confirmed reachable without auth)

| Engine | Endpoint | Response format | Notes |
|---|---|---|---|
| HackerTarget | `https://api.hackertarget.com/hostsearch/?q={domain}` | Plain text, `host,ip` per line | Free tier is rate-limited by IP; treat non-200/limit text as a normal engine error |
| RapidDNS | `https://rapiddns.io/subdomain/{domain}?full=1` | HTML | Scrape via the same regex-extraction helper already used by `ask.go`/`baidu.go`/`bing.go` (`scrape_helper.go`) |
| Wayback Machine | `http://web.archive.org/cdx/search/cdx?url=*.{domain}&output=json&fl=original&collapse=urlkey` | JSON array-of-arrays | First row is the header (`["original"]`); extract hostname from each archived URL |
| CertSpotter | `https://api.certspotter.com/v1/issuances?domain={domain}&include_subdomains=true&expand=dns_names` | JSON | Unauthenticated tier is rate-limited (expect occasional 429) — treat as a normal per-engine error, not fatal, matching Constitution Principle I |

**Rationale**: All four are widely used, keyless, free-tier sources that show up repeatedly across
subfinder's and amass's source lists and independent research into keyless recon sources; adding them
increases passive coverage without adding any new configuration burden (Constitution Principle II
already forbids requiring credentials for baseline functionality).

**Alternatives considered**: Sources requiring API keys even on a free tier (e.g., Censys, Chaos) —
deferred to a future feature since this one is explicitly scoped to keyless sources.

## 5. Rate limiter shape

**Decision**: A single `ratelimit.go` file exposing:

```go
var engineDelays = map[string]time.Duration{
    "crtsh":  1 * time.Second,
    "google": 1500 * time.Millisecond,
    "yahoo":  600 * time.Millisecond,
}

func Wait(engineName string) {
    if d, ok := engineDelays[engineName]; ok && d > 0 {
        time.Sleep(d)
    }
}
```

Existing engines replace their inline `time.Sleep(...)` call with `Wait(e.Name())` at the same call
site.

**Rationale**: Preserves today's exact default delays (satisfies SC-005 by construction — the map is
seeded from the literal values already in the code) while giving a single, greppable place to see or
change every engine's politeness delay, instead of five scattered literals.

**Alternatives considered**: A generic token-bucket rate limiter (like subfinder's per-source
`limit/duration` config) — noted as a natural follow-up if a future feature wants user-configurable
rates, but out of scope here since spec.md's FR-007 only requires de-duplicating the *declaration*,
not adding new user-facing configuration surface.
