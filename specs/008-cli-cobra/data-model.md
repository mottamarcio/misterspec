# Phase 1 Data Model: CLI Command Layer (Cobra)

This feature adds no new persisted entities — every "entity" here is a
JSON *shape* one command's output takes, built by hand from 001-007's
already-existing Go types (research.md's envelope-helper decision).
Extracted from `spec.md`'s Key Entities plus
`docs/architecture-specification.md` §6-18's per-command examples.

## Command Result envelope

Every command's stdout is exactly one JSON object, produced by
`internal/cli/internalcmd/envelope.go`.

| Shape | Fields | Notes |
|---|---|---|
| Success | `{"ok": true, "<key>": <value>}` | `<key>` names what the command returns (`"entity"`, `"parent"`, `"children"`, `"created"`, `"source"`, `"files"`, `"valid"`+`"findings"`, `"counts"`+`"specs"`+`"structural_errors"`) — per §6. |
| Error | `{"ok": false, "error": {"code": string, "message": string}}` | Per §7. `code` is always one of the stable codes below. |

## Error Code

A stable, named identifier for one failure category (FR-003, FR-004),
produced by `internalcmd.classify(err error) (code string, exitCode
int)`.

| Code | Source sentinel | Exit code |
|---|---|---|
| `project_not_initialized` | `project.ErrNotInitialized` | 6 |
| `entity_not_found` | `operations.ErrEntityNotFound` | 3 |
| `entity_ambiguous` | `operations.ErrEntityAmbiguous` (`*operations.AmbiguousIDError`) | 3 |
| `invalid_target` | `operations.ErrInvalidTarget`, or `validation.ErrInvalidTarget` (a distinct sentinel of the same name in a different package — both classify to the same code) | 2 |
| `invalid_parent` | `operations.ErrInvalidParent` | 5 |
| `already_exists` | `operations.ErrAlreadyExists` | 5 |
| `unsupported_type` | `operations.ErrUnsupportedType` | 2 |
| `invalid_argument` | `operations.ErrInvalidSlug`; a malformed/missing CLI argument or flag detected before calling into `operations` at all | 2 |
| `path_outside_project` | `artifacts.ErrPathOutsideProject` | 5 |
| `already_initialized` | `bootstrap.ErrAlreadyInitialized` — new, not in §7's illustrative list (research.md) | 5 |
| `unknown_agent` | `bootstrap.ErrUnknownAgent` — new, not in §7's illustrative list (research.md) | 5 |
| `unexpected_failure` | Anything unrecognized (fallback) | 1 |

**Validation rules**: `classify` handles every *error* path (a command
that did not run to completion) — no command's `RunE` hand-picks a code
or exit code itself (research.md). `validate`'s own `findings`-based
exit code (`4` when `valid: false`, still `ok: true`) is a distinct,
success-path rule documented separately below, not part of `classify`
(research.md). Together, exit codes span every category §8 documents: `0` success,
`1` unexpected failure, `2` invalid invocation, `3` not found,
`4` structural validation failure, `5` mutation rejected, `6` not
initialized.

## Per-command JSON shapes

| Command | Go source | Success JSON |
|---|---|---|
| `internal project` *(not added — research.md)* | — | — |
| `internal resolve <id>` | `operations.Resolve` → `ResolvedLocation` | `{"ok":true,"entity":{"id","type","path"}}` (no `"directory"` — research.md) |
| `internal inspect <id>` | `operations.Inspect` → `InspectResult` | `{"ok":true,"entity":{"id","type","status","parent","depends_on","supersedes"}}` — `parent` is `null` when `Metadata.Parent == nil`; `depends_on`/`supersedes` are `[]` (never `null`) when empty |
| `internal parent <id>` | `operations.Parent` → `ParentResult` | `{"ok":true,"parent":{"id","type"}}` or `{"ok":true,"parent":null}` when `!HasParent` |
| `internal children <id> [--type T]` | `operations.Children` → `[]ResolvedLocation` | `{"ok":true,"children":[{"id","type"}, ...]}` (always an array, `[]` not `null` when empty) |
| `internal create <type> [--parent P] [--slug S]` | `operations.Create` → `CreateResult` | `{"ok":true,"created":{"id","type","path"}}` — `type` from `CreateResult.ID.Type.String()` |
| `internal create-artifact <kind> --for <specID>` | `operations.CreateArtifact` → `CreateArtifactResult` | `{"ok":true,"created":{"type","path","for"}}` — `type` from the parsed `kind` argument, `for` echoes the input Spec ID |
| `internal fingerprint <path>` | `operations.Fingerprint` → `FileFingerprint` | `{"ok":true,"source":{"path","algorithm","fingerprint"}}` — `fingerprint` is `FileFingerprint.String()` |
| `internal inventory <dir>` | `operations.Inventory` → `[]FileEntry` | `{"ok":true,"files":[{"path","extension","size"}, ...]}` |
| `internal validate [<id>]` | `validation.ValidateProject` (no arg) or `validation.ValidateEntity` (with arg) → `[]Finding` | `{"ok":true,"valid":bool,"findings":[{"code","severity","path","message"}, ...]}` — `valid` is `len(findings) == 0`; `severity` is always `"error"` today. Exit code is `4` (not `0`) when `valid: false` — research.md's dedicated decision, distinct from `classify`'s error-path table |
| `internal status` | `operations.Status` → `StatusSummary` | `{"ok":true,"counts":{...},"specs":{...},"structural_errors":N}` — `counts` keyed by `EntityType.String()` (research.md), `specs` is `StatusSummary.SpecsByState` verbatim |
| `init --agent A [--dir D]` | `bootstrap.Bootstrap` → `BootstrapOutcome` | `{"ok":true,"bootstrap":{"project_root","config_written","templates":[{...}],"agent":{"adapter_id","integration_path","outcomes":[...]}}}` |

**Validation rules**: No command's JSON shape invents a field its
underlying Go type does not already carry (FR-009) — every value above
traces directly to an existing struct field, with only the
type-renaming and error-classification steps research.md documents as
new.

## `--type` / `<kind>` name tables (CLI-boundary parsing only)

| Table | Maps | Location |
|---|---|---|
| Entity type names | `"program"/"feature"/"spec"/"knowledge"/"learning"` → `ids.EntityType` | `internal/cli/internalcmd/create.go`, `children.go` (research.md — package-local, not added to `ids`) |
| Artifact kind names | `"plan"/"tasks"/"validation"` → `artifacts.ArtifactType` | `internal/cli/internalcmd/create_artifact.go` |

An unrecognized name at either table is `invalid_argument`, exit `2`,
before any operation is called.
