---

description: "Task list template for feature implementation"
---

# Tasks: Internal Context Command

**Input**: Design documents from `/specs/017-internal-context-command/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/context-command.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 011-wikilink-foundation's, 012-references-backlinks's, 013-document-model-chunking's, 014-sqlite-index's, 015-context-collector's, and 016-ranking-budgeting's full test suites are explicit, named regression gates (Polish), since this feature calls their already-shipped functions unchanged.

**Organization**: Tasks are grouped by user story. **User Story 1** (the command's happy path) must exist before any other story has anything to attach to — every later story's tests exercise the same command. **User Story 2** (error classification) and **User Story 4** (rendering) are genuinely additive, independent extensions of User Story 1's command once it exists. **User Story 3** (transparent index readiness), like 011's own User Story 3 and 016's own User Story 3/4, is a *proving* story: research.md #2 establishes that `index.Open`'s existing schema-recreation and `Store.Sync`'s existing incremental reconciliation (both already shipped in 014) already deliver everything User Story 3 asks for the moment User Story 1's command calls them — so User Story 3 has dedicated proving tests but no new implementation task of its own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US4)
- Every task names its exact file path

## Path Conventions

```text
internal/context/                    # existing package (contextengine) — gains render.go here; request.go, candidate.go extended
internal/context/index/              # existing package (014) — reused via Open/Sync, unmodified
internal/cli/internalcmd/            # existing package — gains context.go here; errors.go extended
internal/cli/                        # existing package — internal.go extended (command registration)
internal/example/                    # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: This feature adds no new package and no new dependency (research.md). There is nothing to front-load before Foundational.

**Checkpoint**: Nothing to verify — proceed directly to Foundational.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Stand up the two small, shared pieces of `internal/context` plumbing every later story needs but no single story "owns" as its own value proposition: a stable, reusable tier-label mapping (`Tier.String()`, needed by both User Story 1's JSON `tier` field and User Story 4's `Render` headings — research.md #5) and an `errors.Is`-matchable sentinel for an unsupported intent (`ErrUnsupportedIntent`, needed by User Story 2's classify case — research.md #4).

**⚠️ CRITICAL**: No user story implementation may begin until this phase's regression checkpoint (T005) passes.

- [X] T001 [P] Unit tests for `Tier.String()` in `internal/context/result_test.go` (Tier's actual home file — result.go, not candidate.go): each of `TierMandatory`, `TierStructural`, `TierSemantic`, `TierText`, `TierSecondHop` returns its documented lowercase label (`"mandatory"`, `"structural"`, `"semantic"`, `"text"`, `"second_hop"`); an out-of-range `Tier` value returns `"unknown"` (data-model.md, research.md #5).
- [X] T002 [P] Unit tests for the `ErrUnsupportedIntent` sentinel in `internal/context/request_test.go`: `validateIntent` given an unrecognized `Intent` value returns an error satisfying `errors.Is(err, ErrUnsupportedIntent)`; `""` and every recognized `Intent` constant still return `nil` unchanged (data-model.md, research.md #4).
- [X] T003 [P] Implement `Tier.String()` in `internal/context/result.go` (data-model.md; corrected from tasks.md's original `candidate.go` reference — Tier is actually declared in result.go). Depends on T001.
- [X] T004 [P] Implement `var ErrUnsupportedIntent = errors.New(...)` and change `validateIntent` to return `fmt.Errorf("%w: %v", ErrUnsupportedIntent, i)` in `internal/context/request.go` (data-model.md). Depends on T002.
- [X] T005 Regression checkpoint: `go build ./...` succeeds; `go vet ./internal/context/...` clean; 011-016's own full suite (`go test ./internal/context/...`) still green (both changes are additive — a new method, a wrapped-but-unchanged error condition). Depends on T003, T004.

**Checkpoint**: Foundation ready — `Tier.String()` and `ErrUnsupportedIntent` exist with proven behavior. User Story 1 implementation can begin.

---

## Phase 3: User Story 1 - Request a Context Pack for a Target (Priority: P1) 🎯 MVP

**Goal**: `misterspec internal context <id> [--intent] [--task] [--query] [--budget]` orchestrates `index.Open`+`Store.Sync` (014), `contextengine.Collect` (015), `contextengine.Rank`+`ApplyBudget` (016) and returns the budgeted, ranked Context Result as one stable JSON envelope.

**Independent Test**: Invoke the command against a known target/intent in a fixture project; confirm the returned JSON matches what directly calling `Collect`→`Rank`→`ApplyBudget` with the same `Request` would produce.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T007-T008.

- [X] T006 [US1] CLI tests in new file `internal/cli/internalcmd/context_test.go` (fixture project matching 015/016's own: SPEC-014 depends on SPEC-011, links to KNOW-003, a Constitution exists): a valid `<target>` + `--intent` returns `ok:true` with `context.items` reflecting `Collect`/`Rank`/`ApplyBudget`'s own output, `context.target`/`context.intent`/`context.budget` echoing the resolved request, and `context.diagnostics` matching `Result.Diagnostics` field-for-field (Acceptance Scenario 1); `--task`, `--query`, and `--budget` all passed together are reflected in the built `Request` and influence `context.items`/`context.budget` accordingly (Acceptance Scenario 2); omitting task/query/budget still succeeds, using `""`/`""`/`contextengine.DefaultBudget` (Acceptance Scenario 3); invoking the identical command twice in a row produces byte-for-byte identical JSON output (Acceptance Scenario 4, FR-009).

### Implementation for User Story 1

- [X] T007 [US1] Implement `NewContextCmd` in new file `internal/cli/internalcmd/context.go`: flags `--intent`, `--task`, `--query`, `--budget` (declared as a **string** flag, not `int` — corrected during implementation, see research.md #6 amendment below — tracked via `cmd.Flags().Changed("budget")` to distinguish "omitted" from "explicit zero/negative"), `--dir` (default `"."`, matching every existing internal command); `RunE` calls `project.Detect(dir)`, opens the fixed cache path `<root>/.misterspec/cache/context.db` via `index.Open`, calls `store.Sync(root, cfg)` (research.md #1, #2), builds a `contextengine.Request` from the flags, calls `Collect`→`Rank`→`ApplyBudget` in sequence, and maps the resulting `Result` into the JSON shape data-model.md defines (`target`, `intent`, `budget`, `estimated_tokens`, `budget_exceeded`, `overage`, `items[]` with `tier` computed from each item's own `Reasons` inline — `minTier` is unexported so `internalcmd` recomputes the same minimum locally rather than duplicating it as a public helper — via `Tier.String()`, `diagnostics`, `rendered: null`), passing every error straight through the existing `WriteError` (contracts/context-command.md). Depends on T006, T005.
- [X] T008 [US1] Register `internalcmd.NewContextCmd()` in `newInternalCmd()` in `internal/cli/internal.go`. Depends on T007.

**Implementation-time correction to research.md #6**: a non-numeric `--budget` cannot be caught by Cobra's own `IntVar` parse-failure fallback in `internal/cli/root.go`, because that fallback only fires when the full command tree is invoked through `root.Execute()` (production's `cli.Execute()`) — every existing per-command test (including this feature's own `context_test.go`, matching every prior command's test file) calls the leaf command's own `cmd.Execute()` directly, which returns pflag's raw parse error un-wrapped (no JSON envelope, no `*ExitCodeError`), exactly as `TestContextCmd_NonNumericBudget` demonstrated (red) before this fix. Corrected design: `--budget` is a **string** flag, parsed manually via `strconv.Atoi` inside `RunE`; a parse failure returns `internalcmd.ErrInvalidArgument` through the existing `WriteError`/`classify` machinery — self-contained and testable at the single-command level like every other internal command's own validation, with no dependency on `root.Execute()`'s cross-cutting fallback. The budget *semantics* research.md #6 actually cared about (a parseable negative/zero value is accepted and passed through unchanged, per 016) are unaffected.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/cli/internalcmd/... -run TestContext` (happy-path cases) passes on its own.

---

## Phase 4: User Story 2 - Get a Clear, Actionable Error for Invalid Input (Priority: P2)

**Goal**: An unknown target, an ambiguous target, an unsupported intent, or a non-numeric budget each produce a distinct, well-formed JSON error — never a success-shaped result, never a crash.

**Independent Test**: Invoke the command with a nonexistent target, then an unsupported intent, then a non-numeric budget; confirm each produces a distinct, well-formed error rather than a success result or a crash.

### Tests for User Story 2

> Write these tests FIRST; confirm the `unsupported_intent` case fails before implementing T010 (the not-found/ambiguous/invalid-argument cases already pass today, since they reuse existing, unmodified classification — these tests exist to prove and guard that reuse).

- [X] T009 [US2] CLI tests in `internal/cli/internalcmd/context_test.go` (extending T006's file): a target ID that does not resolve returns `{"ok": false, "error": {"code": "entity_not_found", ...}}` with the matching exit code (Acceptance Scenario 1); a target that resolves ambiguously returns `"code": "entity_ambiguous"` (Acceptance Scenario 2); `--intent bogus` returns `"code": "unsupported_intent"` (Acceptance Scenario 3); `--budget notanumber` returns `"code": "invalid_argument"` via `RunE`'s own manual `strconv.Atoi` check (see T007's correction note), without ever substituting a default budget (Acceptance Scenario 4); every one of these responses is confirmed to carry `"ok": false` and no `context` key at all (Acceptance Scenario 5).

### Implementation for User Story 2

- [X] T010 [US2] Add one `classify` case in `internal/cli/internalcmd/errors.go`: `errors.Is(err, contextengine.ErrUnsupportedIntent) → ("unsupported_intent", 2)` (data-model.md, contracts/context-command.md). Depends on T009, T004, T007.

**Checkpoint**: User Story 2 is independently complete and testable — `go test ./internal/cli/internalcmd/... -run TestContext_Error` passes on its own.

---

## Phase 5: User Story 3 - Trust the Index Stays Usable Without Manual Repair (Priority: P3)

**Goal**: Prove that `NewContextCmd`'s existing `index.Open`+`Store.Sync` call (already implemented in User Story 1) transparently handles a missing index, a stale index, and a schema-incompatible index — no new implementation of its own (research.md #2).

**Independent Test**: Delete the disposable index entirely, then invoke the command against a target that exists on disk; confirm the command still returns a correct, successful result, having rebuilt whatever derived state it needed.

### Tests for User Story 3

> This story is a regression/correctness guarantee about User Story 1's own already-implemented command (mirroring 016's own User Story 3/4 precedent) — no new implementation task of its own.

- [X] T011 [US3] CLI tests in `internal/cli/internalcmd/context_test.go` (extending T006's file) directly proving spec.md's own User Story 3 acceptance scenarios end-to-end through the real command: invoking the command in a fixture project with no `.misterspec/cache/context.db` yet still returns a correct, successful result, and the file exists afterward (Acceptance Scenario 1); modifying an artifact's content between two invocations changes the second invocation's own result accordingly, proving `Sync` reconciled the change rather than serving stale data (Acceptance Scenario 2); pre-creating a real SQLite `context.db` file with a stale `documents` table and `PRAGMA user_version = 999` (matching `internal/context/index/schema_test.go`'s own `TestEnsureSchema_RecreatesOnVersionMismatch` approach, opened directly via `database/sql`+`modernc.org/sqlite` — a plain-garbage/non-database file was tried first and correctly rejected, which is a different, out-of-scope failure mode from "schema incompatible"), then invoking the command, still returns a correct, successful result rather than an error (Acceptance Scenario 3, exercising `index.Open`'s own existing `ensureSchema` recreation); across every one of the above, `SPEC-014/spec.md` is confirmed byte-for-byte unchanged before and after the invocation (Acceptance Scenario 4, FR-008). Depends on T008.

**Checkpoint**: User Story 3 is independently complete and testable — `go test ./internal/cli/internalcmd/... -run TestContext_IndexReadiness` passes on its own, proving a guarantee that already held once User Story 1 existed.

---

## Phase 6: User Story 4 - Render a Human/Agent-Readable Context Pack (Priority: P4)

**Goal**: An optional `--render` flag adds a Markdown context pack (`context.rendered`) to the same JSON envelope, grouping items by `Tier.String()`, containing every item the structured result already returns, in the same order — never altering selection, ranking, or budgeting.

**Independent Test**: Invoke the command in rendered mode against a known request; confirm the output is well-formed Markdown whose sections correspond exactly to the same items the structured result would return for the identical request.

### Tests for User Story 4

> Write these tests FIRST; confirm they fail before implementing T014-T015.

- [X] T012 [P] [US4] Unit tests for `Render` in new file `internal/context/render_test.go`: the output is well-formed Markdown containing every `ResultItem`'s own `Path`/`Heading`/`Content`; items are grouped under `"## " + Tier.String()` headings in the same ascending-Tier order `Result.Items` already carries, with no re-sorting of items within a group (FR-010, FR-011); an empty `Result` (no items) still produces well-formed, non-panicking output.
- [X] T013 [P] [US4] CLI tests in `internal/cli/internalcmd/context_test.go` (extending T006's file): `--render` populates `context.rendered` with a Markdown string containing every item present in `context.items` (Acceptance Scenario 1); omitting `--render` leaves `context.rendered` as `null` and does not otherwise change `context.items`/`context.diagnostics` (Acceptance Scenario 2); a request that would fail under User Story 2 (e.g. an unknown target) returns the identical well-formed error whether or not `--render` was passed (Acceptance Scenario 3).

### Implementation for User Story 4

- [X] T014 [P] [US4] Implement `Render(req Request, result Result) string` in new file `internal/context/render.go`: group `result.Items` by `minTier(item.Reasons)` (reusing the existing unexported helper, no re-sorting of its own), emit one `"## " + Tier.String()` heading per group present, in ascending Tier order, followed by one subsection per item naming its `Path`/`Heading`, its `Content`, and its own `Reasons`/`Tokens`; a short header line names `req.Target` and `req.Intent` (data-model.md). Depends on T012, T005 (needs `Tier.String()`).
- [X] T015 [US4] Wire a `--render` bool flag into `NewContextCmd` (`internal/cli/internalcmd/context.go`): when set, call `contextengine.Render(req, result)` and set `context.rendered` to its output; otherwise leave it `null` — no other field of the response changes (research.md #7). Depends on T013, T014, T007.

**Implementation note**: T014/T015 (and T012) were implemented immediately after T007, ahead of T009/T010/T011, because `NewContextCmd`'s own `RunE` calls `contextengine.Render` unconditionally in its dead branch (`if render { ... }`), so the package would not compile without `Render` already existing — the actual build order was T001-T008, then T012/T014 (to unblock compilation), then T015, T009/T010, T011, matching Go's own compile-time dependency rather than tasks.md's story-priority ordering. Every acceptance scenario each task names was still verified via its own dedicated test before being marked complete.

**Checkpoint**: User Story 4 is independently complete and testable — `go test ./internal/context/... -run TestRender` and `go test ./internal/cli/internalcmd/... -run TestContext_Render` both pass on their own.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all four stories, no new capability.

- [X] T016 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 017-internal-context-command's packages included) and fix any findings. Clean on first run.
- [X] T017 [P] Verify/extend `internal/context`'s and `internal/cli/internalcmd`'s own package-level doc comments, cross-checked against `specs/017-internal-context-command/contracts/context-command.md`.
- [X] T018 Add a compiled, run-in-CI example in `internal/example` (new file `internal_context_command_quickstart_test.go`, extending the existing package) exercising `quickstart.md`'s flow end-to-end via `internalcmd.NewContextCmd` directly (matching every other `example` file's own convention of exercising the package surface rather than a `root.Execute()`/os.Args-based full binary invocation): a cold-cache first invocation, a repeated invocation proving determinism, an unknown-target error, an unsupported-intent error, a non-numeric-budget error, a deliberately tiny budget proving mandatory content still wins, and a `--render` invocation — each matching `quickstart.md`'s own documented result. Also manually exercised the real compiled `misterspec` binary end-to-end against every quickstart.md scenario during implementation (cold cache, tiny budget, non-numeric budget, `--render`) — output matched exactly.
- [X] T019 Reconciled `specs/017-internal-context-command/contracts/context-command.md`'s signatures/JSON shape against the actual implementation — two drifts fixed: `--budget` is a string flag manually parsed via `strconv.Atoi` (not an `int` flag relying on Cobra's own fallback — see T007's correction note); `Tier.String()` lives in `result.go`, not `candidate.go`.
- [X] T020 Full regression run: `go test ./...` across the entire module (001 through 017) green, `go vet ./...` clean, `gofmt -l .` empty, `go build ./cmd/misterspec` succeeds, `go test ./internal/cli/... ./internal/context/... -race` clean. All confirmed.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do — proceed directly to Foundational.
- **Foundational (Phase 2)**: BLOCKS every user story (both `Tier.String()` and `ErrUnsupportedIntent` are needed no later than User Story 1's own JSON mapping).
- **User Story 1 (Phase 3)**: Depends on Foundational. Nothing else in this feature exists until this command exists.
- **User Story 2 (Phase 4)**: Depends on User Story 1 (tests exercise the same command) and Foundational (the sentinel it classifies).
- **User Story 3 (Phase 5)**: Depends on User Story 1 (proving-only, no new implementation).
- **User Story 4 (Phase 6)**: Depends on User Story 1 (needs a `Result` to render) and Foundational (`Tier.String()`); independent of User Story 2 and User Story 3.
- **Polish (Phase 7)**: Depends on all four user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start immediately after Foundational. No dependency on any other story.
- **User Story 2 (P2)**: Depends on User Story 1 (same command, additive classify case) and Foundational.
- **User Story 3 (P3)**: Depends on User Story 1 (proving-only).
- **User Story 4 (P4)**: Depends on User Story 1 and Foundational; independent of User Story 2/3 — could be built by a different developer in parallel with either.

This feature's chain is Foundational → US1, then fans out into three
independent extensions of US1 (US2, US3-proving, US4) — the same shape
016's own chain took after its own Foundational/US1/US2 core, and
matching this project's established precedent (011, 015, 016) of
letting later, narrower stories attach to an already-complete core
rather than forcing a single linear chain through all four.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V) — except User Story 3, a regression/correctness guarantee about already-implemented behavior, matching 011's and 016's own precedent.
- Foundational's regression checkpoint (T005) passes before any user story's implementation begins.

### Parallel Opportunities

- T001 and T002 (Foundational's two independent test files) can be written in parallel.
- T003 and T004 (Foundational's two independent implementations) can proceed in parallel once their respective tests exist.
- T012 and T013 (User Story 4's two independent test files) can be written in parallel with each other, and with User Story 2's/3's own tasks — all touch different files or different, non-overlapping concerns once User Story 1 (T007-T008) exists.
- Within Polish: T016 and T017 in parallel.

---

## Parallel Example: Foundational

```bash
# These two can be authored together — no shared file, no dependency:
Task: "Unit tests for Tier.String() in internal/context/candidate_test.go"
Task: "Unit tests for the ErrUnsupportedIntent sentinel in internal/context/request_test.go"
```

## Parallel Example: User Story 4 (once User Story 1 exists)

```bash
# These two can be authored together — different files:
Task: "Unit tests for Render in internal/context/render_test.go"
Task: "CLI tests for --render in internal/cli/internalcmd/context_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (nothing to do).
2. Complete Phase 2: Foundational (`Tier.String()`, `ErrUnsupportedIntent`, proven).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/cli/internalcmd/... -run TestContext` green, independently.
5. This alone already makes 011-016's entire pipeline reachable through the actual `misterspec` binary for the first time — the whole point of Phase 8.

### Incremental Delivery

1. Setup (nothing) + Foundational (tier labels, unsupported-intent sentinel, proven).
2. Add User Story 1 → validate independently → the command exists and returns correct results (MVP!).
3. Add User Story 2 (additive classify case) → validate independently → callers can branch reliably on failure.
4. Add User Story 3 (proves User Story 1's own transparent index handling) → validate independently → the "no manual repair step ever needed" guarantee formally proven.
5. Add User Story 4 (adds `Render` + `--render`) → validate independently → a human/agent-readable pack available on request.
6. Polish (Phase 7), including the full-module `-race` regression run (T020).

### Team Strategy

Once User Story 1 exists, User Story 2, User Story 3's proving tests,
and User Story 4 are all independent of one another — three developers
could take one each in parallel, the same shape 016's own User Story
3/User Story 4 fan-out took.

---

## Notes

- [P] tasks touch different files, or the same file with no dependency on an incomplete task's own content.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement — except User Story 3's own proving tests, which pass immediately once User Story 1's implementation is correct.
- This feature is strictly read-only with respect to authoritative project artifacts (FR-008) — the only file this command ever writes is its own disposable cache database (`.misterspec/cache/context.db`); T011 explicitly proves this holds even across index-repair paths.
- `--render` (User Story 4) never changes `context.items`/`context.diagnostics`/`context.budget_exceeded` — it only adds one more field to the same envelope (research.md #7); T013's own tests are the permanent guard against that.
- No new ranking or budgeting logic is introduced anywhere in this feature (research.md) — `Rank`/`ApplyBudget` are called exactly as 016 already shipped them.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
