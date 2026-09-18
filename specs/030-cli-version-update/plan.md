# Implementation Plan: `misterspec --version` and `misterspec --update`

**Branch**: `030-cli-version-update` | **Date**: 2026-09-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/030-cli-version-update/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Add two new root-level CLI flags — `--version` (print the embedded release version) and `--update` (check GitHub for a newer release, confirm, verify, replace) — expanding the public command surface the Constitution now allows (v1.1.0 amendment, this session). Both are genuinely new capabilities requiring build/release-process changes this feature's own scope includes: `release.yml` must start embedding a version string via `-ldflags` and publishing a `SHA256SUMS` file alongside binaries, neither of which exists today. `--update` is the one deliberate, Constitution-sanctioned exception to misterspec's otherwise network-free command surface; it downloads and verifies a new binary via the standard library's own `net/http`, then atomically replaces the currently running executable (with a documented Windows-specific rename-then-swap technique, since Windows cannot overwrite a running executable's file directly the way Unix can), and — only inside an already-initialized project — reinstalls Skills for whatever agent integration `.misterspec/install.json` already records, reusing the existing `bootstrap`/`installer` machinery rather than a new copy-paste path.

## Technical Context

**Language/Version**: Go 1.23.4 (repo-wide). New logic: root-command flag handling (`internal/cli`), a small new package for the GitHub-release/download/verify/replace mechanics, and a `release.yml` workflow change.
**Primary Dependencies**: `net/http` (standard library — the GitHub Releases API and binary download; no new HTTP client dependency), `crypto/sha256` (standard library — checksum verification), `golang.org/x/mod/semver` (new, minimal, Go-team-maintained dependency — correct semver comparison, including prerelease ordering, given this project's own tag history mixes `-alpha` and non-suffixed tags: `v1.1.5-alpha` vs `v1.2.0`), `internal/bootstrap`/`internal/installer` (reused, not duplicated, for the Skill-reinstall half of `--update`), `internal/cli/internalcmd` (reused `WriteSuccess`/`WriteError` JSON envelope — `init` already uses the same envelope for a "public" command, confirmed by reading `internal/cli/init.go`, so `--version`/`--update` follow the identical convention for consistency).
**Storage**: N/A beyond the binary file itself and (for `--update`'s Skill-refresh half) the same `ai/`/`.claude/skills/`-style paths every other install already writes to — no new persisted state, no new config field.
**Testing**: `go test ./internal/cli/...` (flag wiring, JSON output shape for both commands), a new package's own unit tests (semver comparison including the project's real mixed-suffix tag history; checksum verification — match and mismatch; the atomic binary-replace recipe, including the Windows-specific rename-then-swap path, tested via a fake "current OS" seam rather than actually requiring a Windows CI runner) — all using an injectable HTTP transport/round-tripper so no test makes a real network call to GitHub.
**Target Platform**: The same 5 release targets `release.yml` already builds for (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`) — `--update` must correctly pick its own current OS/arch's asset name from that exact existing naming convention (`misterspec-<goos>-<goarch>[.exe]`).
**Project Type**: Single Go project with an embedded content kit — unchanged structurally; adds one new small internal package.
**Performance Goals**: N/A — one GitHub API call, one file download, one checksum computation, one atomic rename; not a hot path.
**Constraints**: MUST NOT make any network call from any command other than `--update` (Constitution's own amended "Embedded kit" bullet — `--update` is the sole deliberate exception). MUST verify the downloaded binary's SHA-256 before replacing anything (FR-007/FR-008/FR-009, SC-004) — verification failure MUST leave the existing binary and Skills completely untouched. MUST require explicit confirmation before replacing the binary (FR-005) — no silent auto-update. MUST NOT fail the whole `--update` run just because no project-level `.misterspec/install.json` exists (FR-011) — that is an expected, reportable non-error state, not a failure. MUST reuse `internal/bootstrap`/`internal/installer`'s existing Skill-installation logic for the Skill-refresh half, never a second, parallel implementation (Constitution Principle VI, DRY).
**Scale/Scope**: One new small package (GitHub release lookup, checksum verification, atomic binary replace); `internal/cli/root.go` gains flag handling for `--version`/`--update` (currently has no persistent flags or root-level `RunE` at all); `.github/workflows/release.yml` gains a `-ldflags` version-embedding step and a checksum-generation/publish step; one new contract doc for the checksum file's own format.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I/II (Semantic/Deterministic Separation, Deterministic Ops as sole mutation primitive)**: PASS. Nothing here touches project-state IDs/paths/fingerprints at all — this is tooling self-maintenance (the binary and its own Skill copies), a different concern entirely from the `ai/`-rooted deterministic-operations layer Principles I/II govern.
- **Principle III (Filesystem is Source of Truth)**: PASS. `--version` reads only its own embedded build-time string (no persisted state at all). `--update` reads `.misterspec/install.json` (already the project's own existing, filesystem-based record) — it does not introduce any new persisted "last checked"/"last updated" state.
- **Principle IV (Simplicity First / YAGNI)**: PASS, with one explicit new dependency justified: `golang.org/x/mod/semver` is a minimal, single-purpose, Go-team-maintained module, justified by a real, demonstrated need (this project's own tag history already mixes `-alpha`-suffixed and bare version tags, which naive string comparison would order incorrectly) — not spec-ulative. No new config field; `.misterspec/install.json`'s existing `agent_id` is reused as-is.
- **Principle V (Test-First, NON-NEGOTIABLE)**: Applies fully — this is entirely new deterministic-in-the-broad-sense Go logic (semver comparison, checksum verification, atomic replace), with unit tests written first, using an injectable HTTP transport so tests never hit the real network (matching this project's own existing test conventions elsewhere, e.g. `internal/vcs`'s real-`git`-but-local-fixture pattern, adapted here to a fake HTTP transport instead of a fake filesystem).
- **Principle VI (Clean Code & SOLID, DRY)**: PASS. Skill-reinstall reuses `bootstrap`/`installer` directly rather than a second copy of that logic; the atomic-write recipe is extended (mode-aware), not duplicated, from `internal/installer.WriteAtomicFile`'s own existing pattern.
- **Principle VII (Explicit Mutation Boundaries)**: N/A to the `ai/`-artifact lifecycle this principle is written for — the "artifacts" here are the tool's own binary and Skill copies, not project-state artifacts. The relevant safety property (never partially replace, never touch anything beyond the binary + Skills) is covered under Principle VIII below instead.
- **Principle VIII (Safety by Construction)**: PASS, and directly on-point — mirrors `init`'s own existing "build a plan, preview it, get confirmation before mutating" contract exactly: `--update` reports current vs. available version, requires explicit confirmation, verifies a checksum before use, and uses an atomic replace (temp file + verified content + rename), leaving nothing partially written on any failure path (FR-006, FR-008, FR-009, FR-013).
- **Principle IX (Transparent, Machine-Readable Contracts)**: PASS. Both new commands use the exact same `WriteSuccess`/`WriteError` JSON envelope `init` (a "public" command) already uses — no new, inconsistent output shape introduced.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/030-cli-version-update/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/selfupdate/
├── version.go              # Version string (set via -ldflags at build time,
│                           # "" / "development build" default); semver
│                           # comparison via golang.org/x/mod/semver
├── release.go               # GitHub latest-release lookup + asset selection
│                           # for the current GOOS/GOARCH, via an injectable
│                           # http.RoundTripper (never a real network call
│                           # in tests)
├── checksum.go              # SHA-256 download + verify against a published
│                           # SHA256SUMS-style file
└── replace.go                # Atomic binary replace: Unix direct rename;
                             # Windows rename-current-aside-then-swap, with
                             # best-effort stale-.old cleanup on a later run

internal/cli/
└── root.go                  # + --version and --update persistent flags on
                            # the root command, + a root-level RunE branching
                            # on them (root has neither today); --update
                            # calls internal/selfupdate, then — if
                            # .misterspec/install.json exists in the current
                            # project — internal/bootstrap's existing
                            # Skill-install path, reused not duplicated

internal/installer/
└── filesystem.go            # WriteAtomicFile gains a file-mode parameter
                            # (extended, not duplicated) so the binary
                            # replace can request 0755 instead of the
                            # existing implicit default

.github/workflows/
└── release.yml               # + -ldflags "-X .../selfupdate.Version=$TAG"
                            # on the existing `go build` step; + a step
                            # generating and publishing a SHA256SUMS file
                            # alongside the existing binary assets
```

**Structure Decision**: Single Go project, embedded-content model already in use. One new small package (`internal/selfupdate`) for the genuinely new mechanics; existing packages (`internal/cli`, `internal/installer`, `internal/bootstrap`) are extended, not duplicated, for everything that overlaps with what they already do.

## Complexity Tracking

*No Constitution Check violations — table not needed.*
