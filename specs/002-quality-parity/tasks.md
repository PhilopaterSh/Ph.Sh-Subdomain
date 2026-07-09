# Tasks: Quality Parity with Modern Subdomain Enumeration Tools

**Input**: Design documents from `/specs/002-quality-parity/`

**Prerequisites**: plan.md, research.md, spec.md

**Tests**: Included — wildcard/resolution logic must be unit-testable without live network access
per plan.md's Technical Context.

**Organization**: Tasks grouped by user story, in priority order (P1 → P5). Each story is
independently shippable — none change default behavior on their own.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

No new dependencies or directories — stdlib only (per plan.md, Primary Dependencies).

## Phase 2: Foundational

No blocking shared infrastructure — each user story below touches its own new file(s) plus a small,
additive wiring change in `main.go`.

---

## Phase 3: User Story 1 - Wildcard filtering (Priority: P1) 🎯 MVP

**Goal**: Detect wildcard DNS per target domain and filter/flag results that resolve to the
wildcard's IP without independent corroboration.

**Independent Test**: Run against a known-wildcarded domain and a known-clean domain; confirm
filtering only triggers on the former (spec.md Acceptance Scenarios 1-2).

### Tests for User Story 1

- [ ] T001 [P] [US1] `wildcard_test.go`: given an injectable fake resolver function, test that 2
      random labels resolving to the same IP marks the domain as wildcarded, and that NXDOMAIN
      answers do not.
- [ ] T002 [P] [US1] `wildcard_test.go`: test that a subdomain reported by 2+ independent engines is
      kept even if it resolves to the wildcard IP (corroboration rule, FR-002).

### Implementation for User Story 1

- [ ] T003 [US1] Create `wildcard.go`: `DetectWildcard(domain string, resolver *net.Resolver) (isWildcard bool, wildcardIPs []string)` using 2× 20-char `crypto/rand`-generated labels (research.md #1-2).
- [ ] T004 [US1] Wire wildcard detection into `main.go`'s per-domain loop, after collecting
      `allSubdomains`/`engineResults` for that domain, before final output.
- [ ] T005 [US1] Apply the corroboration rule: keep any subdomain seen in 2+ `engineResults` keys
      even if it matches a detected wildcard IP; otherwise filter/flag it.
- [ ] T006 [US1] Add `--no-wildcard-filter` flag to `main.go` to fall back to today's unfiltered
      behavior (FR-002).

**Checkpoint**: Wildcarded test domain shows a meaningfully smaller, corroboration-respecting result
set; non-wildcarded domain's output is byte-for-byte unchanged (SC-001).

---

## Phase 4: User Story 2 - Optional active resolution (Priority: P2)

**Goal**: `-resolve` flag concurrently resolves every unique collected subdomain and reports
alive/dead status, reusing the existing bruteforce resolver machinery.

**Independent Test**: Run with and without `-resolve` on a domain with known dead historical
subdomains; confirm status differentiation and zero behavior change when the flag is absent.

### Tests for User Story 2

- [ ] T007 [P] [US2] `resolve_test.go`: given a fake resolver, test that a mix of resolvable and
      NXDOMAIN names produces the correct alive/dead classification per name.
- [ ] T008 [P] [US2] `resolve_test.go`: test that resolution concurrency respects the passed
      semaphore/thread-count bound (no unbounded goroutine fan-out).

### Implementation for User Story 2

- [ ] T009 [US2] Extract the resolver-construction logic already in `bruteforce.go`
      (`customResolver`/round-robin dialer) into a shared helper usable by both `bruteforce.go` and
      the new `resolve.go`, so User Story 2 doesn't duplicate it.
- [ ] T010 [US2] Create `resolve.go`: `ResolveAll(subdomains []string, resolver *net.Resolver,
      threads int) map[string]ResolutionResult`, concurrent and semaphore-bounded like the existing
      bruteforce worker pool.
- [ ] T011 [US2] Add `-resolve` flag to `main.go`; when set, call `ResolveAll` on the final unique
      subdomain list after wildcard filtering (Phase 3) and before output.
- [ ] T012 [US2] Update verbose (`-v`) text output to show resolved IP or a dead/unresolved marker
      per line when `-resolve` was used; leave output unchanged when it wasn't (FR-004).

**Checkpoint**: `-resolve` correctly separates alive/dead on a mixed test domain; omitting it produces
identical output to before this feature (SC-002).

---

## Phase 5: User Story 3 - JSON output (Priority: P3)

**Goal**: `-json`/`-oJ` flag emits the schema from research.md #3.

**Independent Test**: Output validates against a JSON parser; absence of the flag leaves text output
unchanged (spec.md Acceptance Scenarios 1-2).

### Tests for User Story 3

- [ ] T013 [P] [US3] `jsonoutput_test.go`: given a sample `engineResults`/resolution map, test the
      marshaled JSON matches the documented schema, with `resolved`/`ip` omitted when resolution
      wasn't performed.

### Implementation for User Story 3

- [ ] T014 [US3] Create `jsonoutput.go`: builds the `[]SubdomainRecord` structure from
      `engineResults` (+ resolution results if present) and marshals via `encoding/json`.
- [ ] T015 [US3] Add `-json`/`-oJ` flag to `main.go`; when set, write JSON to stdout or `-o` file
      instead of the current plain-text loop (FR-005).

**Checkpoint**: `-json` output parses cleanly; default text output unchanged (SC-003).

---

## Phase 6: User Story 4 - New keyless passive engines (Priority: P4)

**Goal**: Add HackerTarget, RapidDNS, Wayback/CDX, and CertSpotter as new `Engine` implementations.

**Independent Test**: `-e hackertarget,rapiddns,wayback,certspotter -v` against a domain with known
presence in each returns results per engine (spec.md Acceptance Scenarios 1-2).

- [ ] T016 [P] [US4] Create `hackertarget.go` per research.md #4 endpoint/format; parse `host,ip` CSV
      lines, keep host only.
- [ ] T017 [P] [US4] Create `rapiddns.go`; reuse `scrapeSearchEnginePages`-style regex extraction from
      `scrape_helper.go` against the RapidDNS HTML response.
- [ ] T018 [P] [US4] Create `wayback.go`; parse the CDX JSON array-of-arrays, skip the header row,
      extract hostnames from each archived URL.
- [ ] T019 [P] [US4] Create `certspotter.go`; parse `dns_names` from each issuance JSON object,
      treating 429 as a normal per-engine error (not fatal), per Constitution Principle I.
- [ ] T020 [US4] Register all four new engines in `main.go`'s engine list (depends on T016-T019).
- [ ] T021 [US4] Add all four to the README's "Supported Engines" list (Constitution Principle V —
      depends on T020).

**Checkpoint**: All four engines return real data for a domain with known presence, or a clean
per-engine error; README lists them (SC-004).

---

## Phase 7: User Story 5 - Unified rate limiting (Priority: P5)

**Goal**: Replace scattered `time.Sleep` literals with a shared `ratelimit.go` helper, preserving
today's exact default delays.

**Independent Test**: Timed before/after comparison of a default run shows no meaningful change
(spec.md Acceptance Scenarios 1-2).

- [ ] T022 [US5] Create `ratelimit.go` per research.md #5, seeded with today's literal values
      (crtsh=1s, google=1.5s, yahoo=600ms, others=0).
- [ ] T023 [P] [US5] Replace `time.Sleep(1 * time.Second)` in `crtsh.go` with `Wait(e.Name())`.
- [ ] T024 [P] [US5] Replace `time.Sleep(1500 * time.Millisecond)` in `google.go` with
      `Wait(e.Name())`.
- [ ] T025 [P] [US5] Replace `time.Sleep(600 * time.Millisecond)` in `yahoo.go` with `Wait(e.Name())`.

**Checkpoint**: Default run timing unchanged within ±10% (SC-005).

---

## Phase 8: Polish & Cross-Cutting Concerns

- [ ] T026 [P] Run `go vet ./...` and `go build ./...` after all above tasks land.
- [ ] T027 Update `README.md`'s "All Options" block with `-resolve`, `-json`/`-oJ`,
      `--no-wildcard-filter` (Constitution Principle V).
- [ ] T028 Bump the "Program Version" badge in `README.md` if this ships as a tagged release.

---

## Dependencies & Execution Order

- **US1 → US2 → US3 are best done in this order** (US2's "final unique subdomain list" input is
  cleaner after US1's wildcard filtering; US3's JSON schema includes US2's resolution fields when
  present) — but each is independently testable/shippable per its own Independent Test.
- **US4 and US5 are fully independent** of US1-US3 and each other; can be done in parallel by
  different contributors.
- Within US4: T016-T019 are `[P]` (different files); T020 depends on all four; T021 depends on T020.
- Within US5: T023-T025 are `[P]` (different files), all depend on T022.

## Implementation Strategy

**MVP First**: Phase 3 (US1, wildcard filtering) alone delivers the highest-value fix identified by
the competitive research and can ship independently of everything else in this feature.

**Incremental Delivery**: US1 → US2 → US3 → US4 → US5 → Polish, committing after each story/phase so
value ships continuously rather than as one large batch.
