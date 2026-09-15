# Phase 0 Research: Atomic Entity Creation

As with the prior two features, the frozen envelope
(`docs/architecture-specification.md` §10-12, §32-33, §57-61, §70 and the
ratified constitution) leaves no open `NEEDS CLARIFICATION`. This is the
first feature that writes to the filesystem, so most decisions here are
about *how* to write safely, not *whether* to.

## Decision: Lock as its own package, `internal/lock`

- **Decision**: A new, minimal package `internal/lock` with `Acquire`/
  `Release`, used by both `Create` and `CreateArtifact`.
- **Rationale**: Locking is a distinct responsibility from "how to
  scaffold a Program" or "how to scaffold a Plan" — Constitution
  Principle VI (SRP). It is also the one piece of infrastructure both
  creation operations equally need, so it belongs to neither's file.
- **Alternatives considered**: Inlining lock logic into
  `operations/create.go` — rejected, would duplicate the identical logic
  into `create_artifact.go` or force an awkward shared private helper
  inside a file named for a narrower concept than "locking."

## Decision: Lock implementation — `O_CREATE|O_EXCL` plus a staleness timestamp, not OS-level `flock`

- **Decision**: `.misterspec/.lock` is created with `os.OpenFile(path,
  O_CREATE|O_EXCL|O_WRONLY, ...)`, its content is an RFC3339 timestamp (+
  PID, for diagnostics only). `Acquire` retries on `EEXIST`: if the
  existing lock's file mtime is older than `staleAfter`, it is removed and
  acquisition is retried (racing safely — `O_EXCL` decides exactly one
  winner even if two processes attempt the same removal); otherwise
  `Acquire` waits briefly and retries, up to `timeout`, then returns
  `ErrLockTimeout`.
- **Rationale**: `O_CREATE|O_EXCL` is atomic and available identically on
  Linux, macOS, and Windows via the Go standard library — no OS-specific
  `flock`/`LockFileEx` syscalls needed (Constitution Principle IV: no
  dependency for a capability the stdlib already provides). A timestamp
  threshold is what
  `docs/architecture-specification.md` §58 asks for ("stale locks must be
  safely recoverable") without requiring cross-platform process-liveness
  checking (`/proc/<pid>` doesn't exist on Windows/macOS in the same
  form), which would be real added complexity for a project explicitly
  scoped (per spec.md Assumptions) to one local repository worked on by
  one active workflow at a time — not a distributed system needing exact
  liveness detection.
- **Alternatives considered**: A third-party file-locking library
  (`flock`-style, OS-level advisory locks) — rejected, unjustified
  dependency for a single-machine, single-workflow use case (Principle
  IV). PID-liveness checking — rejected as unnecessary complexity/
  non-portability for the same reason; a time-based staleness threshold is
  simpler and sufficient here.
- **Defaults**: `staleAfter` = 10s, `timeout` (max total wait to acquire)
  = 5s with ~50ms poll interval — generous relative to how long scaffolding
  a handful of files actually takes, so a legitimate concurrent creation
  is never mistaken for staleness, while a crashed process's lock is
  recovered well within one interactive workflow's patience.

## Decision: The lock closes the "already exists" race, not filesystem atomicity alone

- **Decision**: `Create` and `CreateArtifact` hold the lock for their
  entire critical section: existence check → scan/allocate (Create only)
  → template render → atomic write (temp file + `fsync` + `os.Rename`,
  per §57) → release. The "target already exists" rejection (FR-008,
  FR-012) is race-free *with respect to other misterspec-initiated
  creators* because they all serialize on the same lock — not because
  `os.Rename` itself refuses to overwrite (it does not, on POSIX).
- **Rationale**: Go's standard library has no cross-platform
  "rename-if-not-exists." Given this project's own scope (spec.md
  Assumptions: one local repository, one active workflow, not
  multi-machine coordination), serializing all creators through the
  already-required lock is sufficient and simpler than inventing a
  filesystem-level exclusive-rename trick. This is the exact purpose
  §58's lock already exists for.
- **Alternatives considered**: `O_CREATE|O_EXCL` directly on the final
  artifact path (atomic create-and-check in one syscall, no rename needed)
  — rejected because it writes content directly to the final path with no
  temporary intermediate, meaning a reader could observe a partially
  written file mid-write (violates FR-007), which the lock+temp+rename
  approach avoids entirely.

## Decision: Templates as embedded `text/template` files, in their own package

- **Decision**: A new package `internal/templates`, embedding one
  `text/template` file per entity/artifact type (`program.md.tmpl`,
  `feature.md.tmpl`, `spec.md.tmpl`, `knowledge.md.tmpl`,
  `learning.md.tmpl`, `plan.md.tmpl`, `tasks.md.tmpl`,
  `validation.md.tmpl`) via `go:embed`, with a typed `Render` function per
  kind.
- **Rationale**: The frontmatter + body shape for each type is already
  frozen exactly in `docs/architecture-specification.md` §22-31 — these
  are data, not logic, and are far more legible as separate `.tmpl` files
  than as Go string constants. `go:embed` makes the binary self-contained
  for exactly these 8 files, without building the broader "Embedded Kit"
  system (Skills, integrations) `misterspec init` will need later
  (Constitution Principle IV — don't build that scope now).
- **Alternatives considered**: Go string constants inline in
  `operations/create.go` — rejected as harder to review/diff for content
  that is fundamentally a document template, not program logic. Building
  out the full `kit/templates/` embedded-kit structure now — rejected as
  premature scope (this feature needs 8 specific files, not an
  installable kit).

## Decision: Reuse `operations.Resolve` for parent validation

- **Decision**: `Create`'s parent-existence/type check calls
  `operations.Resolve` (from `002-read-operations`) on the declared parent
  ID, rejecting with a specific reason if it returns
  `ErrEntityNotFound`/`ErrEntityAmbiguous`/`ErrInvalidTarget`, or if the
  resolved entity's type doesn't match what the requested child type
  requires (e.g. a Feature's declared parent must resolve as a Program).
- **Rationale**: `Resolve` already implements exactly "does this ID exist,
  unambiguously, and what type is it" — reimplementing that check inside
  `Create` would duplicate `002-read-operations`'s logic (Constitution
  Principle VI, DRY).
- **Alternatives considered**: A bespoke, lighter-weight parent check
  inside `create.go` (e.g. just `ids.Scan` without going through
  `Resolve`) — rejected; it would silently drift from `Resolve`'s own
  ambiguity/not-found semantics over time.

## Decision: `CreateArtifact` reuses `artifacts.ArtifactType`, restricted to Plan/Tasks/Validation

- **Decision**: `CreateArtifactRequest.Kind` is an `artifacts.ArtifactType`
  value; `CreateArtifact` rejects any value other than `TypePlan`,
  `TypeTasks`, or `TypeValidation` as `ErrUnsupportedType`.
- **Rationale**: `artifacts.ArtifactType` already models exactly this
  vocabulary (`001-core-foundation`); inventing a second, narrower enum
  just for this feature would duplicate it for no benefit (Principle VI).
- **Alternatives considered**: A dedicated `operations.ArtifactKind` enum
  with only three values — rejected as unjustified duplication of an
  existing, already-correct type.

## Decision: Knowledge/Learning slug is caller-supplied, not generated

- **Decision**: `CreateRequest.Slug` is required for Knowledge/Learning
  and used verbatim in the filename (`KNOW-<NNN>-<slug>.md`); `Create`
  validates it is non-empty and filesystem-safe (no path separators, no
  `..`), but does not derive it from anything.
- **Rationale**: Choosing a meaningful slug describing the topic is a
  semantic judgment (Constitution Principle I) — squarely the agent's job,
  not this deterministic layer's.
- **Alternatives considered**: Deriving a slug automatically from a title
  string (slugification) — rejected as unrequested scope; nothing in
  spec.md asks for it, and it would blur the semantic/deterministic
  boundary Principle I exists to keep sharp.

## Output

All unknowns resolved. No `NEEDS CLARIFICATION` markers remain.
