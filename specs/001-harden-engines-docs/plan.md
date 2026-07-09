# Implementation Plan: Align Documentation with Behavior & Harden Remaining Engines

**Branch**: `001-harden-engines-docs` | **Date**: 2026-07-09 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-harden-engines-docs/spec.md`

## Summary

Bring `README.md` back in sync with actual CLI behavior (remove the unimplemented "Email
Discovery" claim, verify every documented flag exists), trim `requirements.txt` to only what
`digger_wrapper.py` actually imports, verify/fix the `threatcrowd` engine's target host, and add
a first `_test.go` file covering the subdomain-cleaning pipeline (dedup, wildcard-stripping,
crt.sh multi-SAN newline splitting) plus a check that the DNS bruteforce resolver round-robin
fix actually distributes across all configured resolvers.

## Technical Context

**Language/Version**: Go 1.21 (module `github.com/PhilopaterSh/Ph.Sh_Sub`); Python 3.6+ only for
the optional `digger` engine's embedded wrapper script.

**Primary Dependencies**: `gopkg.in/yaml.v2` (Go); `cloudscraper` (Python, digger engine only).

**Storage**: N/A — no persistence beyond the optional YAML config file and plain-text output files.

**Testing**: Go standard library `testing` package (`go test ./...`); no existing test files today.

**Target Platform**: Cross-platform CLI (Windows/Linux/macOS), built via `go build`/`go install`.

**Project Type**: Single-module CLI tool (not a library, no frontend/backend split).

**Performance Goals**: N/A — this feature is documentation/reliability hardening, not a performance
change. Existing concurrency model (`-t` threads, semaphore-bounded) is unaffected.

**Constraints**: Tests added by this feature MUST run with no network access (Constitution does not
mandate network-free CI today, but User Story 3's independent test requires it so contributors get
fast, deterministic feedback).

**Scale/Scope**: Touches `README.md`, `requirements.txt`, `threatcrowd.go`, and adds one or more
new `*_test.go` files. No new engines, no new CLI flags, no breaking changes to existing flags.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle V (Documentation Must Match Behavior)** is the direct driver of User Story 1 — PASS,
  this plan exists specifically to satisfy it.
- **Principle I (Engine Contract Integrity)** governs User Story 2's ThreatCrowd host check — PASS,
  plan verifies rather than changes the `Engine` interface or error-handling contract.
- **Principle III (Concurrency Safety)** governs User Story 3's round-robin resolver test — PASS,
  plan only adds a test around the existing `sync/atomic` counter, no new shared state introduced.
- No principle is violated; no complexity exceptions required. Constitution Check: **PASS** (no
  re-check needed after Phase 1 — this feature has no design/architecture phase, only fixes + tests).

## Project Structure

### Documentation (this feature)

```text
specs/001-harden-engines-docs/
├── plan.md              # This file
└── tasks.md             # Phase 2 output (/speckit-tasks / manual equivalent)
```

No `research.md`, `data-model.md`, `quickstart.md`, or `contracts/` are needed — this feature has no
unresolved technical unknowns (Technical Context above has zero `NEEDS CLARIFICATION` markers) and
introduces no new data model or API contract.

### Source Code (repository root)

```text
# Option 1: Single project (this repository's actual layout — flat Go package `main`)
Ph.Sh-Subdomain-main/
├── main.go, *.go              # Engine implementations + orchestration (existing, package main)
├── digger_wrapper.py          # Embedded Python helper for the digger engine (existing)
├── requirements.txt           # Python deps for digger_wrapper.py (User Story 2 touches this)
├── README.md                  # User Story 1 touches this
├── cleaning_test.go            # NEW — User Story 3: cleanAndUniqueSubdomains + crt.sh split coverage
├── bruteforce_test.go          # NEW — User Story 3: resolver round-robin coverage
└── threatcrowd.go              # User Story 2 touches this if host is stale
```

**Structure Decision**: This is a flat single-package Go CLI (no `src/`/`tests/` split, no
frontend/backend). New tests are added as `*_test.go` files alongside the code they cover,
matching Go convention and this repo's existing flat layout — no new directories needed.

## Complexity Tracking

*No Constitution Check violations — this section is intentionally empty.*
