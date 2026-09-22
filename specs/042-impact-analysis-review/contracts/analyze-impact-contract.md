# Contract: Impact Analysis (`internal analyze-impact`)

This documents the new deterministic operation and CLI command this
feature adds (Constitution Principle IX). No existing command's output
shape changes.

## 1. `internal/vcs` (extended)

```go
package vcs

// DiffEntry is one changed path between two revisions (data-model.md
// "ChangeSet"/"ChangedElement" inputs).
type DiffEntry struct {
	Path   string
	Status string // "A" | "M" | "D" (git's own --name-status letters)
}

// DiffNameStatus reports every path that differs between from and to,
// via `git diff --name-status <from> [<to>]`. to == "" diffs against
// the working tree (plain `git diff <from>`), matching
// IsWorkingTreeDirty's own working-tree scope (research.md #2).
func DiffNameStatus(root, from, to string) ([]DiffEntry, error)

// FileAtRevision returns path's content at rev via `git show
// <rev>:<path>`. found is false, with no error, when path did not
// exist at rev (an added file — mirrors CommitsSinceFileAdded's own
// "bool separate from error" convention). rev == "" reads the
// working-tree file directly (os.ReadFile), not through git show.
func FileAtRevision(root, path, rev string) (content []byte, found bool, err error)
```

## 2. `internal/validation` (extended)

```go
package validation

// RequirementSections scans body (a Spec's own spec.md body) the same
// way parseSpecRequirements already does, additionally returning each
// found "R<N>" heading's own Section (data-model.md "New, additive
// exports"). parseSpecRequirements itself is unchanged; this is a new,
// additive sibling (research.md #3).
func RequirementSections(spec ids.EntityID, body []byte) map[int]artifacts.Section
```

## 3. `internal/impact` (new package)

```go
package impact

// AnalyzeImpactRequest names the two revisions to compare and an
// optional path scope (data-model.md).
type AnalyzeImpactRequest struct {
	From string // required; a resolvable Git revision
	To   string // optional; "" means working tree
	Path string // optional; when set, the diff is scoped to this one
	      // artifact path instead of the whole project
}

// AnalyzeImpact computes the ChangeSet between req.From/req.To
// (scoped to req.Path when set), then walks every changed element's
// reverse relations (formal, coverage, evidence, wikilink) to a fixed
// point (bounded by a visited-element set, never re-expanding the same
// element — research.md #6), and returns the full ImpactReport
// (data-model.md). Returns ErrRevisionNotFound (wrapping the
// underlying git error) when From or a non-empty To does not resolve
// in root's repository; ErrNotARepository when root is not a Git
// repository at all — errors.go's existing classify() maps both to
// stable JSON error codes, the same pattern every other internal
// command already follows.
func AnalyzeImpact(root string, cfg project.Configuration, req AnalyzeImpactRequest) (ImpactReport, error)
```

`AnalyzeImpact` never mutates the filesystem (Constitution Principle
VII) — read-only, safe to run repeatedly and from CI.

## 4. CLI: `misterspec internal analyze-impact`

```text
misterspec internal analyze-impact --from <revision> [--to <revision>] [--path <artifact-path>] [--dir <project-dir>]
```

- `--from` is required; `--to` defaults to the working tree; `--path`
  defaults to the whole project.
- `--dir` defaults to `.`, same convention as every other `internal`
  command.

### Success envelope (`{"ok": true, ...}`)

```json
{
  "ok": true,
  "analyze_impact_schema_version": 1,
  "change_set": {
    "from": "a1b2c3d",
    "to": "working-tree",
    "elements": [
      {
        "id": "SPEC-014",
        "path": "programs/PROG-001/features/FEAT-002/specs/SPEC-014/spec.md",
        "status": "modified",
        "requirement_number": null
      },
      {
        "id": "SPEC-014",
        "path": "programs/PROG-001/features/FEAT-002/specs/SPEC-014/spec.md",
        "status": "modified",
        "requirement_number": 2
      }
    ],
    "unmapped_code_paths": 0
  },
  "affected_items": [
    {
      "id": "SPEC-014:TASK-003",
      "path": "programs/PROG-001/features/FEAT-002/specs/SPEC-014/tasks.md#TASK-003",
      "classification": "deterministic_invalidation",
      "severity": "high",
      "reason": "TASK-003 serves SPEC-014:R2, whose content changed between a1b2c3d and working-tree",
      "paths": [
        {
          "hops": [
            {
              "relation": "coverage",
              "from": "SPEC-014",
              "to": "SPEC-014:TASK-003",
              "source_path": "programs/PROG-001/features/FEAT-002/specs/SPEC-014/tasks.md",
              "source_section": "",
              "source_line": 0
            }
          ]
        }
      ],
      "reverification_candidate": "SPEC-014:TASK-003"
    }
  ],
  "no_known_relation_elements": []
}
```

`analyze_impact_schema_version` follows the same additive-versioning
precedent as `contextSchemaVersion` (033/035/036/038) — bumped only if
this response's shape changes in a way existing consumers must be
aware of.

### Error envelope (`{"ok": false, ...}`)

Reuses the existing `{"ok": false, "error": {"code", "message"}}`
shape (`internalcmd.WriteError`). New stable error codes:

| Code | Condition |
|---|---|
| `revision_not_found` | `--from` or a non-empty `--to` does not resolve in the project's Git repository. |
| `not_a_repository` | The target project directory is not a Git repository at all (`vcs.IsRepo` false) — `analyze-impact` cannot run without Git history to diff. |

## 5. Behavioral guarantees this contract makes (traceable to spec FRs)

- FR-004/data-model.md "Classification": no input produces
  `classification: "deterministic_invalidation"` for an `AffectedItem`
  reached only by a path whose last hop is `"wikilink"`.
- FR-006: `AnalyzeImpact` always returns (never hangs) on a project
  containing a reference/dependency cycle — enforced structurally by
  the visited-element set (research.md #6), not by a timeout.
- FR-008: `no_known_relation_elements` is populated, never merely
  implied by an empty `affected_items`, whenever a `ChangeSet` element
  has zero reachable relations.
- FR-010: identical `(from, to, path)` input against unchanged project
  state always yields byte-identical `affected_items` ordering and
  `severity` values (no time-based, random, or map-iteration-order
  nondeterminism).
