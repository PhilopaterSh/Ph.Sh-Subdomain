# Tasks: Align Documentation with Behavior & Harden Remaining Engines

**Input**: Design documents from `/specs/001-harden-engines-docs/`

**Prerequisites**: plan.md, spec.md

**Tests**: Included — spec.md User Story 3 explicitly requires automated test coverage.

**Organization**: Tasks are grouped by user story so each can be completed and verified
independently, in priority order (P1 → P2 → P3).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)

## Phase 1: Setup

No setup phase needed — this feature works entirely within the existing flat Go module and
requires no new dependencies, directories, or infrastructure.

---

## Phase 2: Foundational

No foundational/blocking phase — the three user stories below touch disjoint files
(`README.md` / `requirements.txt` + `threatcrowd.go` / new `*_test.go` files) and have no shared
prerequisite work.

---

## Phase 3: User Story 1 - README matches actual CLI behavior (Priority: P1) 🎯 MVP

**Goal**: Remove or implement the unimplemented "Email Discovery" claim; verify every flag listed
in the README's "All Options" block matches `main.go`.

**Independent Test**: Run `-d <domain> -v` and confirm README no longer promises email output;
diff README's flag list against `main.go`'s `flag.*` declarations.

- [ ] T001 [US1] Grep the entire repository for email-parsing/collection logic; confirm none exists.
- [ ] T002 [US1] Remove the "Email Discovery" bullet from `README.md`'s Features list and the
      "and any discovered emails" phrase from the verbose-mode usage example.
- [ ] T003 [US1] Cross-check every flag in `README.md`'s "All Options" block against the
      `flag.String`/`flag.Bool`/`flag.Int` declarations in `main.go`; fix any mismatched
      description or missing/extra flag.

**Checkpoint**: README contains zero claims that don't map to real code (SC-001).

---

## Phase 4: User Story 2 - requirements.txt and engine endpoints reflect reality (Priority: P2)

**Goal**: Trim `requirements.txt` to only what `digger_wrapper.py` imports; verify/fix the
`threatcrowd` engine's target host.

**Independent Test**: `pip install -r requirements.txt` in a clean venv installs only imported
packages; `-e threatcrowd -v` against a domain with known ThreatCrowd data returns results.

- [ ] T004 [P] [US2] Diff `requirements.txt` against the `import` statements in
      `digger_wrapper.py`; remove `requests` and `colorama` if still unused (keep `cloudscraper`).
- [ ] T005 [P] [US2] Manually query `https://www.threatcrowd.org/searchApi/v2/domain/report/?domain=<test-domain>`
      and `https://ci-www.threatcrowd.org/searchApi/v2/domain/report/?domain=<test-domain>` for a
      domain with known subdomains; compare responses.
- [ ] T006 [US2] Based on T005: update `threatcrowd.go`'s target host to whichever one returns real
      data, and add a one-line comment explaining the choice (depends on T005).

**Checkpoint**: `requirements.txt` has no unused packages; `threatcrowd` engine verified to return
real data (or documented as legitimately empty).

---

## Phase 5: User Story 3 - Regression coverage for the subdomain cleaning pipeline (Priority: P3)

**Goal**: Add Go tests for `cleanAndUniqueSubdomains` (dedup, lowercase, wildcard-strip,
invalid-entry rejection, crt.sh multi-SAN newline splitting) and the bruteforce resolver
round-robin, runnable with no network access.

**Independent Test**: `go test ./...` passes with network disabled.

### Tests for User Story 3

- [ ] T007 [P] [US3] Create `main_test.go` with `TestCleanAndUniqueSubdomains` covering: mixed
      case + duplicates collapse to one lowercase entry; `*.` wildcard prefix stripped; an entry
      with an embedded `\n` (post-`strings.Split`, simulating a crt.sh multi-SAN `name_value`)
      yields every valid sub-entry; a bare hostname with no dot is rejected.
- [ ] T008 [P] [US3] Create `bruteforce_test.go` with a test that constructs the custom resolver's
      `Dial` closure logic (or an equivalent extracted helper) with N fake resolver addresses,
      invokes it concurrently more than N times, and asserts every resolver address was selected
      at least once (verifying no starvation of the round-robin counter).

**Checkpoint**: `go test ./...` is green with no network access; both new test files pass.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T009 [P] Run `go vet ./...` and `go build ./...` after all above tasks land; fix any new
      warnings introduced by the test files.
- [ ] T010 Update the "Program Version" badge in `README.md` if this feature ships as part of a
      tagged release.

---

## Dependencies & Execution Order

- **US1, US2, US3 are mutually independent** — no shared files, can be done in any order or in
  parallel by different contributors.
- Within US2: T006 depends on T005 (must know which host is correct before changing code).
- Within US3: T007 and T008 are independent, both `[P]`.
- Phase 6 (Polish) runs after whichever of US1/US2/US3 have been completed.

## Implementation Strategy

**MVP First**: Complete Phase 3 (US1) alone already satisfies the highest-priority constitution
gap (false README claim) and can be shipped independently of US2/US3.

**Incremental Delivery**: US1 → US2 → US3 → Polish, committing after each user story so the
"Documentation Must Match Behavior" fix ships as soon as possible without waiting on test-writing
or the ThreatCrowd host investigation.
