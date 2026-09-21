# Quickstart: Validating Lean Skills & Integration Contracts

Prerequisites: this branch's Go toolchain (1.23.4) with `go test`
available; no built `misterspec` binary or project fixture is required
beyond what `internal/example`'s own test fixtures already set up.

## 1. Edit a shared rule once, see it everywhere (User Story 1)

```sh
grep -rn "mechanical-steps-note\|resolve-preamble" internal/skillgen/fragment.go
```

**Expected**: each canonical fragment's `Body` appears exactly once in
Go source. Then:

```sh
go generate ./internal/skillgen/...
git diff --stat kit/skills
```

**Expected**: running this a second time with no further fragment/
manifest edit produces an *unchanged* `git diff --stat` — generation is
deterministic, so re-running it never adds further changes on top of
whatever is already committed (contract §1's drift-check invariant,
exercised manually here the same way CI exercises it via
`skillgen_drift_test.go`, `go test ./internal/example/... -run
TestSkillgenDrift`).

## 2. Skills never cite a command or promise a check the binary can't back (User Story 2)

```sh
go test ./internal/example/... -run TestSkillsContent -v
```

**Expected**: passes, and its output confirms three checks per Skill —
the existing `internal <op>` allowlist check, the new example-syntax
check (FR-005), and the new verification-capability allowlist check
(FR-006). To see a deliberate failure, temporarily add `internal
does-not-exist` to any `kit/skills/*/SKILL.md` and re-run — the test
fails naming that Skill and the unknown command.

## 3. Representative integrations behave consistently (User Story 3)

```sh
go test ./internal/example/... -run TestSkillSmoke -v
```

**Expected**: passes for all three rows (`claude-code`, `cursor-agent`,
`copilot`), each showing its own `TargetPath` and the
`resolve → context/prepare → validate` chain completing with the
documented JSON shape. A failure names the specific adapter and step
(spec Acceptance Scenario US3.1) — for example, temporarily renaming an
adapter's `TargetPath()` return value reproduces a single-row failure
without affecting the other two.

## 4. Context Engine fallback is visible to evaluation (User Story 4)

```sh
go test ./internal/eval/... -run TestMetrics -v
```

**Expected**: `Metrics{ContextFallbacks: 2}.Validate()` returns `nil`;
`Metrics{ContextFallbacks: -1}.Validate()` returns a non-nil
`ErrInvalidRunRecord`-wrapped error.

```sh
go test ./internal/eval/... -run TestCompare_ContextFallbacksSurvivesComparison -v
```

**Expected**: passes. Correction found during implementation:
`compare.go`'s per-task diffing (`compareTaskResults`) only ever
compares `TaskResult.Outcome` — it has never diffed individual
`Metrics` fields (`ExtraReads`/`Rework` aren't surfaced by `Compare()`
either, both pre-dating this feature), so there was no existing
per-field comparison path for `ContextFallbacks` to join. This test
instead confirms the field survives `Compare()` and a `RunRecord`
round-trip intact — satisfying FR-008/SC-005 by making the figure a
durable, structured, JSON field any consumer of two `RunRecord`s can
read and diff directly, without adding new comparison-report logic
that would be scope beyond what this feature needs.

## 5. Size finding (SC-001)

```sh
go test ./internal/example/... -run TestSkillsContent_SizeDoesNotRegressUnexpectedly -v
```

**Expected**: passes, logging the current total line count. Correction
found during implementation, documented in full in
`internal/example/skills_size_test.go`'s own header comment: extracting
the real duplicated fragments (User Story 1) does not shrink
`kit/skills/*/SKILL.md`'s own rendered size, because each generated file
stays byte-identical, full Markdown — deduplication happens in the
authoring source (`internal/skillgen/fragment.go`), not in what an
agent actually reads, and no single agent invocation reads more than
one Skill file. SC-001's original 80%-of-baseline target is therefore
not met (current total is 2393 lines, up from the 2378-line baseline,
due to the FR-008 fallback-recording text) and is not achievable by
this feature's approach without cutting real Skill content against the
spec's own quality-preservation assumption. This test instead guards
against further, larger regressions. SC-001 itself should be revisited
in a follow-up spec amendment.
