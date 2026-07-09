# Feature Specification: Align Documentation with Behavior & Harden Remaining Engines

**Feature Branch**: `001-harden-engines-docs`

**Created**: 2026-07-09

**Status**: Draft

**Input**: User description: "Align README claims with actual implemented functionality and harden the remaining passive-recon engines"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - README matches actual CLI behavior (Priority: P1)

A user reads `README.md` before running the tool and expects every advertised feature and flag to
work exactly as described. Today the README advertises an "Email Discovery" feature ("Gathers
emails associated with the domain... visible in verbose mode") that has no corresponding code
anywhere in the project — running `-v` never shows emails, because no engine collects them.

**Why this priority**: A false feature claim erodes trust in every other README claim and violates
Constitution Principle V (Documentation Must Match Behavior), which is non-negotiable.

**Independent Test**: Grep the codebase for email-collection logic; if none exists, remove the
claim from `README.md` (or implement it). Verify by running `-d <domain> -v` and confirming the
README no longer promises behavior the output doesn't produce.

**Acceptance Scenarios**:

1. **Given** the README lists "Email Discovery" as a feature, **When** the codebase is searched for
   email-parsing logic, **Then** either matching code is found and documented accurately, or the
   claim is removed from the README and its usage example.
2. **Given** the corrected README, **When** a new user follows the "All Options" section, **Then**
   every flag listed there exists in `main.go`'s `flag.*` declarations with matching behavior.

---

### User Story 2 - requirements.txt and engine endpoints reflect reality (Priority: P2)

A user installing the optional `digger` engine dependency runs `pip install -r requirements.txt`
and expects every listed package to be required. Today `requirements.txt` lists `requests` and
`colorama`, but `digger_wrapper.py` only imports `cloudscraper`, `sys`, `json`, and `re`. Separately,
`threatcrowd.go` queries `ci-www.threatcrowd.org` instead of ThreatCrowd's documented
`www.threatcrowd.org` host, which should be verified as intentional (e.g., a working mirror) or
corrected.

**Why this priority**: Unused dependencies waste install time and confuse contributors about what's
load-bearing; a wrong API host silently returns zero results forever, masquerading as "the domain
just has no data" (Constitution Principle I: engine failures must be visible, not silently empty).

**Independent Test**: Run `pip install -r requirements.txt` in a clean venv and confirm every listed
package is actually imported somewhere in the Python source. Run `-e threatcrowd -v` against a
domain with known ThreatCrowd data and confirm non-zero results.

**Acceptance Scenarios**:

1. **Given** `requirements.txt`, **When** cross-referenced against `import` statements in
   `digger_wrapper.py`, **Then** every listed package is actually imported, or removed if unused.
2. **Given** the `threatcrowd` engine, **When** run against a domain with publicly known ThreatCrowd
   subdomain data, **Then** it returns non-empty results (or the code documents why the alternate
   host is intentional).

---

### User Story 3 - Regression coverage for the subdomain cleaning pipeline (Priority: P3)

A contributor changes `cleanAndUniqueSubdomains`, `isValidSubdomain`, or an engine's parsing logic
and wants to know immediately if they broke multi-value splitting (like the crt.sh multi-SAN fix)
or validation. Today the project has zero automated tests, so regressions are only caught by manual
runs against live third-party services, which are slow, rate-limited, and non-deterministic.

**Why this priority**: Lower priority than the two doc/behavior fixes above, but the multi-SAN
crt.sh bug fixed in this codebase's history shows exactly the class of regression tests would catch
for free, without depending on network access.

**Independent Test**: Run `go test ./...` with no network access available and confirm the cleaning
pipeline's core behaviors (dedup, lowercase, wildcard-prefix stripping, invalid-entry rejection,
newline-split multi-SAN handling) are verified without hitting any real engine endpoint.

**Acceptance Scenarios**:

1. **Given** a raw subdomain list with mixed case, duplicates, a `*.` wildcard prefix, and an
   embedded newline (simulating a crt.sh multi-SAN entry), **When** `cleanAndUniqueSubdomains`
   processes it, **Then** the output is a sorted, deduplicated, lowercase list with every valid
   name present and no invalid/empty entries.
2. **Given** the DNS bruteforce custom-resolver dialer, **When** invoked concurrently across many
   goroutines with N configured resolvers, **Then** all N resolvers are observed to be used (no
   resolver starvation), verifying the round-robin fix.

### Edge Cases

- What happens when a source returns a name with no dot at all (e.g., a bare hostname)? It must be
  rejected by `isValidSubdomain`, not silently included.
- What happens when `requirements.txt` is trimmed but a future contributor adds Python code that
  needs `requests`? The dependency list must be re-verified as part of that change, not assumed.
- What happens when ThreatCrowd's host is changed and it *does* start requiring the alternate host
  going forward? The engine's behavior (and any comment explaining the host choice) must be
  updated together so the two never silently drift apart again.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `README.md` MUST NOT describe functionality absent from the codebase; the "Email
  Discovery" claim MUST be removed or implemented before the next release.
- **FR-002**: `README.md`'s "All Options" usage block MUST list exactly the flags defined in
  `main.go`, with descriptions matching the `flag.*` help text.
- **FR-003**: `requirements.txt` MUST list only packages actually imported by
  `digger_wrapper.py` (or any other Python source added later).
- **FR-004**: The `threatcrowd` engine's target host MUST be verified against live ThreatCrowd data;
  if `ci-www.threatcrowd.org` is stale, it MUST be corrected to a host that returns real data, with
  a code comment explaining the choice either way.
- **FR-005**: The repository MUST include a Go test file covering `cleanAndUniqueSubdomains` and
  the crt.sh newline-split behavior, runnable via `go test ./...` with no network access.
- **FR-006**: The repository MUST include a test (or documented manual verification) that the DNS
  bruteforce custom resolver round-robins across all configured resolvers rather than favoring one.

### Key Entities

- **Engine**: a passive subdomain source (`Fetch`/`Name`); this feature does not add new engines,
  only corrects/tests existing ones.
- **README feature claim**: a bullet in `README.md`'s Features/Usage sections; each must map 1:1 to
  actual code.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `README.md` contains zero feature or flag claims that cannot be matched to code within
  the same audit pass (verified by manual cross-reference, not automated in this pass).
- **SC-002**: `pip install -r requirements.txt` installs only packages that are actually imported
  somewhere in the project's Python source.
- **SC-003**: `go test ./...` passes with no network access and exercises at least the multi-SAN
  split and wildcard/dedup behaviors of `cleanAndUniqueSubdomains`.
- **SC-004**: Running `-e threatcrowd -v` against a domain with known public ThreatCrowd data
  returns at least one result, or the code/comment explains why it legitimately returns none.

## Assumptions

- The project owner (PhilopaterSh) is the sole maintainer approving README/behavior changes; no
  external stakeholder sign-off is required.
- Adding Go tests does not require a testing framework beyond the standard library `testing` package
  already implied by `go.mod`'s `go 1.21` toolchain.
- Network-dependent verification (ThreatCrowd host, live engine runs) is done manually by whoever
  implements this feature; it is not expected to be part of CI given the existing GitHub Actions
  workflow only builds the binary (per `.github/workflows/go.yml`).
