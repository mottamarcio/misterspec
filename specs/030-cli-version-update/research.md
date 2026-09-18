# Phase 0 Research: `misterspec --version` and `misterspec --update`

No `[NEEDS CLARIFICATION]` markers remained in the Technical Context — the open design questions were resolved directly with the user during `/speckit-specify` (see `checklists/requirements.md` Notes), and the Constitution conflict was resolved via an explicit amendment (v1.1.0) before any requirement was written. This file records the remaining technical decisions grounded in the actual codebase.

## Decision: Version is embedded via `-ldflags`, not a committed file

- **Decision**: `internal/selfupdate.Version` is a package-level `var` (empty by default), set at build time via `go build -ldflags "-X github.com/mottamarcio/misterspec/internal/selfupdate.Version=$TAG"` in `release.yml` — the standard Go mechanism for build-time version injection, requiring no new file to keep in sync.
- **Rationale**: `release.yml`'s own existing `go build -o "${{ steps.asset.outputs.name }}" ./cmd/misterspec` step already has the tag available (`GITHUB_REF_NAME`/`inputs.tag`, per the existing "Determine release tag" step); adding one `-ldflags` argument costs nothing and needs no separate versioning file that could drift from the actual tag a binary was built from.
- **Alternatives considered**:
  - *A committed `VERSION` file read at startup*: rejected — a file can be read from a local, non-release build too, producing a misleading "version" that doesn't correspond to any real released binary; `-ldflags` binds the version to the actual build/release event instead.

## Decision: `golang.org/x/mod/semver` for version comparison

- **Decision**: Use the Go team's own `golang.org/x/mod/semver` package to compare the running version against the latest published tag, rather than a hand-rolled string comparison.
- **Rationale**: This project's own real tag history already mixes suffixed and bare versions (`v1.1.5-alpha` immediately followed by `v1.2.0` — confirmed via `git tag --sort=-v:refname` during this session's own release work) — naive string or even simple dot-split numeric comparison gets prerelease-vs-release ordering wrong in exactly this kind of case. `x/mod/semver` correctly implements the real semver precedence rules (including prerelease ordering), and — being an `x/` module maintained by the Go team itself, single-purpose, with no further transitive dependencies of its own — is a minimal, low-risk, clearly justified addition (Constitution Principle IV: justified by an actual, not hypothetical, need).
- **Alternatives considered**:
  - *Hand-rolled comparison (split on `.`/`-`, compare numerically)*: rejected — this project's own tag history is the concrete counter-example that would break: naive comparison could easily rank `v1.1.5-alpha` above `v1.2.0` (`5 > 2` digit-wise on the wrong field) unless real semver precedence rules are implemented correctly, which is exactly what `x/mod/semver` already does, tested, for free.

## Decision: GitHub Releases API via `net/http`, injectable transport for tests

- **Decision**: `internal/selfupdate/release.go` calls `GET https://api.github.com/repos/mottamarcio/misterspec/releases/latest` via the standard library's `net/http`, with the `*http.Client` (or its `Transport`) passed in as a parameter — never a package-level default client — so every test substitutes a fake transport and never makes a real network call.
- **Rationale**: No third-party HTTP client exists anywhere in this codebase today (confirmed via `go.mod`) and none is needed — the standard library is entirely sufficient for one GET request and one file download. Constitution Principle V requires real tests for this logic; a hard-coded `http.DefaultClient` would make that untestable without hitting the real network, which this project's own testing conventions elsewhere (e.g. `internal/vcs`'s real-`git`-but-local-fixture pattern) never do for external dependencies.
- **Alternatives considered**:
  - *A third-party GitHub API client library (e.g. `go-github`)*: rejected — far more surface area (auth, pagination, rate-limit handling for a wide API) than the one "get latest release" call this feature needs; the standard library's `net/http` plus a small hand-written JSON struct for the one response shape used is simpler and has zero new supply-chain exposure (Principle IV, YAGNI).

## Decision: Checksums published as one `SHA256SUMS` file, not per-asset `.sha256` files

- **Decision**: `release.yml` generates a single `SHA256SUMS` text file (the standard `sha256sum`-style format: `<hex-digest>  <filename>` per line, one line per published binary asset) and uploads it alongside the 5 existing binary assets.
- **Rationale**: One file to fetch and parse is simpler than 5 separate `<asset>.sha256` files, and matches a widely-recognized convention (the same format `sha256sum -c` already validates directly) — easy for a human to also verify by hand if they ever want to, not just for `--update` itself.
- **Alternatives considered**:
  - *Per-asset `.sha256` files*: rejected — more assets to fetch and manage for no real benefit over one combined file, for only 5 total binaries.
  - *GPG-signing releases instead of/in addition to checksums*: considered, but rejected as out of scope — the spec's own clarified requirement is checksum verification specifically (protecting against a corrupted/incomplete download and network-level tampering via HTTPS-adjacent risks), not a full code-signing trust chain, which would be a substantially larger, separate effort.

## Decision: Atomic binary replace — direct rename on Unix, rename-current-aside on Windows

- **Decision**: `internal/selfupdate/replace.go` downloads the new binary to a temp file in the same directory as the currently running executable (via `os.Executable()`), verifies its checksum, `chmod`s it to `0755`, then: on Unix (`linux`/`darwin`), a plain `os.Rename(tmpPath, exePath)` — safe even while the old binary is still executing, since Unix `rename(2)` just repoints the directory entry, and the running process keeps its already-open reference to the old (now-unlinked) inode until it exits. On Windows, `os.Rename` cannot overwrite a currently-running `.exe` directly (the OS holds an exclusive lock on it while it's executing), so the sequence is instead: rename the *currently running* exe aside to `<name>.old` first (Windows permits renaming, just not overwriting, a running exe), then rename the verified temp file into the original path. A best-effort attempt to delete the leftover `.old` file happens on this same run if possible, and is retried opportunistically the next time `--version` or `--update` runs, in case Windows still held it locked immediately after the swap.
- **Rationale**: This is a well-established, standard technique for self-updating tools on Windows (the OS's own file-locking semantics leave no simpler direct-overwrite path); documenting it explicitly here means the plan doesn't quietly assume Unix-only behavior for a feature whose own release matrix explicitly includes `windows/amd64`.
- **Alternatives considered**:
  - *Spawn a helper process/script to perform the swap after the main process exits*: rejected as unjustified complexity for this project's scope (Principle IV) — the rename-aside trick achieves the same safety property (never a moment where the exe path has no valid file) without a second process or any OS-specific scripting.
  - *Punt on Windows entirely (report "unsupported" and exit)*: rejected — Windows is one of the project's own 5 supported release targets; leaving it deliberately broken would contradict SC-002's "one command" promise for a real fraction of this tool's own users.

## Decision: `WriteAtomicFile` gains a mode parameter, reused not duplicated

- **Decision**: Extend `internal/installer.WriteAtomicFile`'s signature to accept an `os.FileMode` (all existing callers pass their current implicit default explicitly, preserving today's behavior exactly), rather than writing a second, parallel atomic-write helper inside `internal/selfupdate` just to get `0755` instead of the default.
- **Rationale**: Constitution Principle VI (DRY) — the underlying recipe (temp file in the same directory, write, fsync, rename) is identical; only the desired file permissions differ between an ordinary Markdown/YAML artifact write and an executable binary write.
- **Alternatives considered**:
  - *A second, `internal/selfupdate`-local copy of the same ~20-line recipe*: rejected — `internal/installer.WriteAtomicFile`'s own doc comment already explains why this recipe was kept out of `internal/operations` in the first place (avoiding a backwards architectural dependency); no equivalent justification exists for a *third* copy specifically for a permission difference a single parameter already solves.

## Decision: Skill-reinstall reuses `internal/bootstrap`, gated on `.misterspec/install.json` presence

- **Decision**: After a successful binary replace, `--update` checks for `.misterspec/install.json` in the current project (the same resolution every other command already uses); if present, it reads the recorded `agent_id` and calls the same `bootstrap`/`installer` path `misterspec init --agent <id>` already uses to (re-)install Skills, rather than a new, separate copy mechanism. If absent, this step is skipped and reported plainly (FR-011) — never a failure.
- **Rationale**: This is the exact manual process already carried out by hand earlier this session for the `pregoeiros` project (copy the new binary, re-copy `kit.SkillsFS` content into `.claude/skills/`) — automating precisely that, through the existing installer rather than a new implementation, directly satisfies Constitution Principle VI (DRY) and the spec's own FR-010.
- **Alternatives considered**:
  - *A new, `--update`-specific Skill-copy routine*: rejected — `internal/bootstrap`/`internal/installer` already do exactly this, byte-identically (verified this session via `TestSkillsContent_AllCanonicalSkillsInstalled`'s own byte-for-byte comparison); reusing it is strictly simpler and cannot drift from `init`'s own behavior over time.
