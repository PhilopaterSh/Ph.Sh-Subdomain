<!--
Sync Impact Report
Version change: (none) → 1.0.0
Modified principles: n/a (initial ratification)
Added sections: Core Principles (I-V), Security & Operational Constraints, Development Workflow, Governance
Removed sections: none
Templates requiring updates:
  ✅ .specify/templates/plan-template.md (generic Constitution Check gate, no changes needed)
  ✅ .specify/templates/spec-template.md (generic, no changes needed)
  ✅ .specify/templates/tasks-template.md (generic, no changes needed)
Follow-up TODOs: none
-->

# SubHunter-Go (Ph.Sh_Sub) Constitution

## Core Principles

### I. Engine Contract Integrity
Every subdomain source MUST be implemented as an `Engine` (`Fetch(domain, client) ([]string, error)`
+ `Name() string`) and registered in `main.go`. A single engine's failure or malformed response
MUST NOT crash the run or block other engines — errors are returned and reported per-engine,
never panicked. Raw results MUST pass through `cleanAndUniqueSubdomains` before being presented;
an engine MUST NOT silently truncate multi-value fields (e.g. newline-joined SANs) — every
distinct name a source returns must reach the cleaning stage.

### II. No Secrets in Source
API keys and other credentials MUST NEVER be hardcoded or committed to the repository. The only
supported path for credentials is the user-owned config file (`Ph.Sh_Sub_config.yaml`, loaded via
`InitConfig`/`loadConfig`) or environment-level secrets outside version control. Any code path
that reads a credential MUST degrade gracefully (skip the engine) when the credential is absent,
rather than erroring out the whole run.

### III. Concurrency Safety & Politeness
Network calls run concurrently across engines bounded by a semaphore (`-t`/`threads`). Shared
mutable state accessed from goroutines (e.g. round-robin counters, result maps) MUST use
synchronization primitives (`sync`, `sync/atomic`) — data races are not acceptable. Engines that
scrape rate-limit-sensitive third parties (search engines, crt.sh) MUST keep their existing
politeness delays and MUST NOT be parallelized internally in a way that removes them.

### IV. Passive, Authorized Reconnaissance Only
This tool performs passive/OSINT-style subdomain discovery for security research and bug bounty
use. Contributions MUST NOT add active exploitation, credential brute-forcing against live
services, or denial-of-service-capable behavior. DNS bruteforce is limited to name resolution
against user-supplied or default wordlists — it MUST remain read-only reconnaissance.

### V. Documentation Must Match Behavior (NON-NEGOTIABLE)
`README.md` (features, flags, examples) MUST accurately reflect what the code does. A feature
described in the README that has no corresponding implementation is a constitution violation and
MUST be fixed by either implementing it or removing the claim — whichever is decided, it MUST be
resolved before a release is tagged.

## Security & Operational Constraints

- `--no-ssl-verify` MUST default to `false`; disabling TLS verification is opt-in only and scoped
  to the shared client, except for sources (e.g. ThreatCrowd) that are documented as requiring it.
- Dependencies are kept minimal and vendored via `go.mod`/`go.sum`; both files MUST stay in sync
  (no import without a corresponding `go.sum` entry) so `go build`/`go install` work from a clean
  checkout without extra flags.
- The Python `digger` engine is an optional, embedded, best-effort source: its failure (missing
  Python/`cloudscraper`) MUST be reported as a normal engine error, never a fatal exit.

## Development Workflow

- Before a release: run `go vet ./...` and `go build ./...` clean; verify the startup banner and
  `--help` output render correctly on Windows and POSIX terminals.
- New engines follow the existing file-per-engine convention (`<engine>.go`, package `main`) and
  are added to both the engine list in `main.go` and the "Supported Engines" list in `README.md`
  in the same change.
- Shared scraping logic (pagination, User-Agent, regex extraction) belongs in `scrape_helper.go`;
  new search-engine-style scrapers MUST reuse it instead of duplicating the loop.

## Governance

This constitution supersedes ad-hoc practice for this repository. Amendments are made by editing
this file, incrementing the version per semantic versioning (MAJOR: incompatible principle removal
or redefinition; MINOR: new principle or materially expanded guidance; PATCH: wording/clarification),
and updating the Sync Impact Report at the top of the file. Pull requests that touch engine
behavior, credential handling, or README claims MUST be checked against these principles before
merge; violations block the merge until resolved or the constitution is amended with justification.

**Version**: 1.0.0 | **Ratified**: 2026-07-09 | **Last Amended**: 2026-07-09
