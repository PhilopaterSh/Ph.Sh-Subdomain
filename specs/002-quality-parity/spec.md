# Feature Specification: Quality Parity with Modern Subdomain Enumeration Tools

**Feature Branch**: `002-quality-parity`

**Created**: 2026-07-09

**Status**: Draft

**Input**: User description: "Bring the tool to feature parity with modern subdomain enumeration
tools: wildcard DNS filtering, optional active resolution, JSON output, additional free passive
sources, unified rate limiting"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Wildcard domains no longer flood results with false positives (Priority: P1)

A user scans a domain that has a wildcard DNS record (`*.example.com` resolves to some IP). Today,
`crt.sh` and other certificate-transparency-based sources return every name ever issued under that
wildcard, and the tool has no way to tell "real" subdomains from wildcard noise — `cleanAndUniqueSubdomains`
only validates string shape, never DNS behavior. The user ends up with thousands of subdomains, the
majority of which don't correspond to distinct, intentionally-provisioned hosts.

**Why this priority**: This is the single biggest quality gap versus subfinder/puredns/amass, all of
which treat wildcard filtering as mandatory, not optional. Without it, every other improvement in this
spec is measured against noisy, low-trust output.

**Independent Test**: Run the tool against a domain with a known wildcard DNS record and a domain
without one. On the wildcard domain, confirm the tool detects the wildcard and does not report every
resolvable random label as a distinct finding. On the non-wildcard domain, confirm behavior is
unchanged (no regressions from this feature).

**Acceptance Scenarios**:

1. **Given** a target domain with a wildcard `A`/`CNAME` record, **When** the tool resolves 2-3
   random, almost-certainly-nonexistent labels under that domain and they all resolve to the same
   IP(s), **Then** the tool marks that domain (or subtree) as wildcarded and flags/filters
   subsequent results that resolve to the same IP(s) without independent corroboration.
2. **Given** a target domain with no wildcard record, **When** the same random-label probe is run,
   **Then** the probe resolves to NXDOMAIN/no answer and no filtering is applied.
3. **Given** wildcard filtering is active, **When** the user passes a flag to disable it (e.g.
   `--no-wildcard-filter`), **Then** the tool falls back to today's unfiltered behavior.

---

### User Story 2 - Optional active resolution confirms subdomains are still alive (Priority: P2)

A user wants to know which of the passively-discovered subdomains actually resolve today, since
certificate-transparency-derived results can be years old and point at decommissioned hosts. The
tool already contains DNS resolution logic (`bruteforce.go`'s custom resolver) but only applies it
to the bruteforce wordlist, never to results from the other 15 passive engines.

**Why this priority**: High value, but secondary to wildcard filtering — an alive/dead flag is much
less useful if the underlying list is already polluted with wildcard noise from User Story 1.

**Independent Test**: Run the tool with a new `-resolve` flag against a domain with a mix of live and
long-dead historical subdomains (e.g., one visible in old crt.sh certs). Confirm the output
distinguishes resolved (alive) from unresolved (dead) entries, and that omitting `-resolve` preserves
today's behavior exactly (no DNS lookups beyond bruteforce).

**Acceptance Scenarios**:

1. **Given** the `-resolve` flag is passed, **When** enumeration finishes collecting all passive (and
   optional bruteforce) results, **Then** each unique subdomain is resolved once, concurrently, reusing
   the existing resolver/semaphore patterns already used by bruteforce.
2. **Given** `-resolve` is passed with `-v`, **When** results are printed, **Then** each line indicates
   resolved IP or a dead/unresolved marker.
3. **Given** `-resolve` is NOT passed, **When** the tool runs, **Then** no additional DNS traffic is
   generated and output is byte-for-byte identical to the current behavior.

---

### User Story 3 - JSON output for pipeline interoperability (Priority: P3)

A user wants to pipe results into other recon tools (e.g., httpx-style HTTP probers, nuclei) that
expect structured input. Today the tool only supports plain-text stdout or a plain-text `-o` file.

**Why this priority**: Widely expected interoperability convention among modern recon tools, but
purely additive — no existing behavior changes, so it's safely lower priority than the two
correctness-affecting stories above.

**Independent Test**: Run with a new `-json`/`-oJ` flag and confirm the output is valid, line-delimited
or array JSON containing at minimum `{subdomain, source}` (and `{resolved, ip}` when `-resolve` was
also used), parseable by a standard JSON tool without errors.

**Acceptance Scenarios**:

1. **Given** `-json` is passed, **When** the tool finishes, **Then** stdout (or the `-o` file, if also
   given) contains valid JSON reflecting every unique subdomain and the engine(s) that found it.
2. **Given** neither `-json` nor `-oJ` is passed, **When** the tool runs, **Then** output format is
   unchanged from today (plain text).

---

### User Story 4 - Additional free, no-API-key passive sources (Priority: P4)

A user wants broader passive coverage without configuring more API keys. Competing tools integrate
dozens of keyless sources; this project currently has ~11 keyless engines. Adding a few well-known
keyless sources (HackerTarget, RapidDNS, Wayback Machine/CDX, CertSpotter) increases coverage with no
new configuration burden for the user.

**Why this priority**: Pure coverage/breadth improvement, valuable but not correctness-critical —
ranked below the three stories that fix output quality and interoperability.

**Independent Test**: Run `-e hackertarget,rapiddns,wayback,certspotter -v` against a domain known to
have historical presence in each source and confirm each engine returns at least one result and
integrates cleanly into the existing `Engine` interface and cleaning pipeline.

**Acceptance Scenarios**:

1. **Given** a new source engine (e.g., `hackertarget.go`), **When** it is added, **Then** it
   implements the existing `Engine` interface exactly like current engines, requires no API key, and
   is added to both the engine list in `main.go` and the README's "Supported Engines" list.
2. **Given** one of these new sources is temporarily down or rate-limits the request, **When** it
   errors, **Then** the run continues and reports the per-engine error exactly like existing engines
   (Constitution Principle I).

---

### User Story 5 - Unified, configurable per-engine rate limiting (Priority: P5)

A user running many domains via `-dl` hits inconsistent, hardcoded delays scattered across engine
files (`crt.sh` = 1s, `google` = 1.5s, `yahoo` = 600ms, others = none), making total run time
unpredictable and impossible to tune without editing source.

**Why this priority**: Lowest priority — an internal-quality/maintainability improvement with no
user-visible correctness impact, appropriate as the final polish item in this feature.

**Independent Test**: Inspect engine source after the change and confirm delay values are declared in
one place (not scattered `time.Sleep` literals) and are overridable, and that scraping-heavy engines
(google, yahoo, bing, baidu, ask) still behave politely by default with no regression in delay
duration.

**Acceptance Scenarios**:

1. **Given** the existing per-engine `time.Sleep` calls, **When** replaced with a shared rate-limiter
   helper, **Then** default delay values are unchanged (no more/less polite than today).
2. **Given** the shared rate limiter, **When** a user wants to tune it, **Then** there is a single,
   documented place to do so (e.g., a config value or flag) rather than editing multiple `.go` files.

### Edge Cases

- What happens when the wildcard-probe random label itself collides with a real, coincidentally-named
  subdomain? Use a sufficiently long random label (e.g., 20+ chars) to make collision negligible, and
  document the residual risk rather than claiming zero false negatives.
- What happens when `-resolve` is used on a domain with thousands of subdomains? Resolution MUST be
  concurrent and bounded by the existing `-t`/threads semaphore, not sequential.
- What happens when a target uses split-horizon DNS (different answers for internal vs. public
  resolvers)? Out of scope — this feature only reasons about public resolver behavior, consistent with
  the tool's existing passive/public-recon scope (Constitution Principle IV).
- What happens when `-json` and `-v` are both passed? `-json` takes precedence for the structured
  output; verbose grouping becomes a field in the JSON (e.g., `source` per entry) rather than two
  competing text formats.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST detect wildcard DNS behavior per target domain before finalizing output,
  using a random-label resolution probe against public resolvers.
- **FR-002**: The tool MUST filter or clearly flag subdomains that resolve to a detected wildcard's
  IP(s) without independent corroboration, and MUST support disabling this via a flag.
- **FR-003**: The tool MUST support an opt-in `-resolve` flag that performs concurrent DNS resolution
  of all collected unique subdomains and reports resolved/unresolved status.
- **FR-004**: The tool MUST NOT perform any additional DNS resolution beyond today's behavior when
  `-resolve` is not passed (no default behavior change).
- **FR-005**: The tool MUST support an opt-in JSON output mode (`-json`/`-oJ`) producing valid,
  parseable JSON reflecting subdomain, source engine, and (when available) resolution status.
- **FR-006**: New passive source engines added under this feature MUST implement the existing
  `Engine` interface and require no API key.
- **FR-007**: Per-engine scraping delays MUST be defined in a single shared location rather than
  duplicated `time.Sleep` literals per file, with unchanged default values.

### Key Entities

- **WildcardProbe**: a per-domain check result — probed random labels, resolved IP(s) (if any), and
  whether the domain is classified as wildcarded.
- **ResolutionResult**: a per-subdomain outcome when `-resolve` is used — subdomain, resolved IP(s) or
  dead/unresolved marker.
- **Engine** (existing, unchanged contract): new keyless sources plug into this same interface.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On a known-wildcarded test domain, output size (unique subdomain count) drops
  significantly after wildcard filtering versus today's unfiltered baseline, without losing
  non-wildcard entries that were present before.
- **SC-002**: `-resolve` correctly distinguishes at least one known-dead historical subdomain from
  live ones on a test domain with both, adding no more than one resolution round-trip per unique
  subdomain.
- **SC-003**: `-json` output validates against a JSON parser with zero errors across at least one
  real run.
- **SC-004**: At least the four named new engines (HackerTarget, RapidDNS, Wayback/CDX, CertSpotter)
  are added, each returning results for at least one domain with known historical presence in that
  source, with zero required API keys.
- **SC-005**: Default run time and request cadence for existing scraping engines is unchanged
  (±10%) after the rate-limiter refactor, confirmed by comparing a timed run before/after.

## Assumptions

- Public DNS resolvers (e.g., 8.8.8.8, 1.1.1.1) are reachable from wherever the tool runs; split-horizon
  or fully offline environments are out of scope, consistent with this being a passive/public recon
  tool (Constitution Principle IV).
- "Independent corroboration" for User Story 1 means: a subdomain resolving to the wildcard's IP is
  still kept if 2+ independent engines separately reported that exact name (reduces false negatives
  for legitimately-hosted names that happen to share the wildcard's IP).
- JSON output schema is versioned informally via this spec (not a separate formal contract file) since
  there is exactly one consumer pattern described (piping to other recon tools) and no external API
  contract to maintain.
- New engines (User Story 4) are added one at a time behind the existing `-e` engine-selection flag,
  so partial delivery (e.g., only HackerTarget lands first) is still immediately useful.
