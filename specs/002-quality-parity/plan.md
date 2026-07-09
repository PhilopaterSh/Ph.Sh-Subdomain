# Implementation Plan: Quality Parity with Modern Subdomain Enumeration Tools

**Branch**: `002-quality-parity` | **Date**: 2026-07-09 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-quality-parity/spec.md`

## Summary

Add a wildcard-DNS detection/filtering pass (P1), an opt-in active resolution pass reusing the
existing bruteforce resolver machinery (P2), an opt-in JSON output mode (P3), four new keyless
passive engines — HackerTarget, RapidDNS, Wayback/CDX, CertSpotter (P4) — and a shared per-engine
rate-limiter helper replacing scattered `time.Sleep` literals (P5). All five are additive and
independently shippable; none change default output for a run that uses none of the new flags.

## Technical Context

**Language/Version**: Go 1.21, same module as the rest of the project. No new language/runtime.

**Primary Dependencies**: Standard library only (`net`, `net/http`, `encoding/json`, `sync`,
`math/rand`/`crypto/rand` for the wildcard probe label) — no new third-party Go module required.

**Storage**: N/A — all new state (wildcard probe results, resolution results) is in-memory for the
duration of a single run, same lifecycle as existing `allSubdomains`/`engineResults` maps in `main.go`.

**Testing**: Go `testing` package, extending the coverage started in spec 001
(`specs/001-harden-engines-docs`). Wildcard/resolution logic must be unit-testable without a live
network dependency by accepting an injectable resolver function.

**Target Platform**: Same as existing tool — cross-platform CLI (Windows/Linux/macOS).

**Project Type**: Single-module CLI (unchanged).

**Performance Goals**: Wildcard probe adds at most 2-3 DNS lookups per target domain (negligible).
Active resolution (`-resolve`) adds one lookup per unique subdomain, concurrent and bounded by the
existing `-t` semaphore — must not become the dominant cost of a run for typical result set sizes
(hundreds to low thousands of subdomains).

**Constraints**: Every new flag/behavior MUST be opt-in or transparent by default (FR-004, SC-005) —
this plan explicitly rejects any change to default output.

**Scale/Scope**: New files: `wildcard.go`, `resolve.go` (or extend `bruteforce.go`), `jsonoutput.go`,
`hackertarget.go`, `rapiddns.go`, `wayback.go`, `certspotter.go`, `ratelimit.go`. Modifies: `main.go`
(new flags, wiring), `README.md` (new engines + flags), `go.mod` unaffected (stdlib only).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Engine Contract Integrity)**: New engines (US4) MUST implement `Engine` exactly;
  wildcard/resolution results must not silently drop valid entries — PASS by design (see FR-002's
  corroboration rule, which exists precisely to avoid over-filtering).
- **Principle III (Concurrency Safety)**: Active resolution (US2) and the new engines' HTTP calls
  reuse the existing semaphore/goroutine/WaitGroup pattern already in `main.go` and `bruteforce.go` —
  no new unsynchronized shared state. PASS.
- **Principle IV (Passive, Authorized Reconnaissance Only)**: Wildcard probing and active resolution
  are read-only DNS lookups against public resolvers, not active exploitation — PASS, explicitly
  scoped out of split-horizon/internal DNS in spec.md's Edge Cases.
- **Principle V (Documentation Must Match Behavior)**: New flags and engines MUST be added to
  `README.md` in the same change that introduces them (tracked as tasks below) — PASS, enforced by
  process, not automation.
- No violations requiring Complexity Tracking.

*Re-check after Phase 1 (below)*: unchanged — the research decisions in Phase 0 don't introduce any
new dependency, storage, or cross-cutting concern that would trip a principle. Constitution Check:
**PASS**.

## Phase 0: Research

Open technical questions resolved here (full detail in `research.md`):

1. **Wildcard probe resolver choice** — use the same public-resolver approach already used
   optionally in `bruteforce.go` (user-supplied `-resolvers` file) with a hardcoded fallback pair
   (e.g., 8.8.8.8, 1.1.1.1) when the user hasn't supplied one, mirroring puredns's use of trusted
   resolvers for wildcard validation.
2. **Random label generation** — 20+ character random alphanumeric label per probe, 2 probes per
   domain (reduces both false-negative wildcard misses and accidental collision risk, per spec.md
   Edge Cases).
3. **JSON schema shape** — flat array of objects: `{"subdomain":"...","sources":["crtsh","google"],
   "resolved":true,"ip":"1.2.3.4"}` (`resolved`/`ip` omitted unless `-resolve` was used); chosen over
   line-delimited JSON for simplicity given expected result sizes (hundreds–low thousands, not
   millions like massdns-scale tools).
4. **New engine endpoints** (all keyless, confirmed reachable without auth):
   - HackerTarget: `https://api.hackertarget.com/hostsearch/?q={domain}` (plain-text `host,ip` CSV)
   - RapidDNS: `https://rapiddns.io/subdomain/{domain}?full=1` (HTML scrape, same regex-extraction
     pattern already used by `ask.go`/`baidu.go`/`bing.go` via `scrape_helper.go`)
   - Wayback Machine: `http://web.archive.org/cdx/search/cdx?url=*.{domain}&output=json&fl=original&collapse=urlkey`
   - CertSpotter: `https://api.certspotter.com/v1/issuances?domain={domain}&include_subdomains=true&expand=dns_names`
     (unauthenticated tier is rate-limited; treat 429 as a normal per-engine error, not fatal)
5. **Rate limiter shape** — a small `ratelimit.go` exposing `Wait(engineName string)` backed by a
   `map[string]time.Duration` of default per-engine delays (seeded with today's exact values: crtsh=1s,
   google=1.5s, yahoo=600ms, others=0) so SC-005 (unchanged default timing) holds by construction.

All decisions above resolve every `NEEDS CLARIFICATION` that would otherwise appear in Technical
Context — none remain open.

## Project Structure

### Documentation (this feature)

```text
specs/002-quality-parity/
├── plan.md              # This file
├── research.md          # Phase 0 output (the 5 decisions above, detailed)
└── tasks.md             # Phase 2 output
```

No `data-model.md`/`contracts/` — the only "data model" is the two small structs (`WildcardProbe`,
`ResolutionResult`) captured in spec.md's Key Entities, which is sufficient without a separate file.
No `quickstart.md` — usage is fully covered by new README flag documentation (a task below).

### Source Code (repository root)

```text
Ph.Sh-Subdomain-main/
├── main.go                 # Wire new flags: -resolve, -json/-oJ, --no-wildcard-filter
├── wildcard.go              # NEW — US1: per-domain wildcard detection + filtering
├── resolve.go                # NEW — US2: active resolution pass over collected subdomains
├── jsonoutput.go              # NEW — US3: JSON marshaling of final results
├── ratelimit.go                # NEW — US5: shared per-engine delay helper
├── hackertarget.go              # NEW — US4
├── rapiddns.go                   # NEW — US4
├── wayback.go                     # NEW — US4
├── certspotter.go                  # NEW — US4
├── bruteforce.go                    # MODIFIED — reuse/share resolver dialer with resolve.go
├── crtsh.go, google.go, yahoo.go,
│   bing.go, ask.go, baidu.go        # MODIFIED — call ratelimit.Wait(name) instead of time.Sleep
└── README.md                          # MODIFIED — document new flags + 4 new engines
```

**Structure Decision**: Same flat single-package `main` layout as spec 001 — every new capability is
one new file plus minimal wiring in `main.go`, consistent with this repo's existing file-per-concern
convention (file-per-engine, `scrape_helper.go` for shared scraping logic).

## Complexity Tracking

*No Constitution Check violations — this section is intentionally empty.*
