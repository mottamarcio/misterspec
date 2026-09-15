# Phase 1 Contracts: Canonical Skills Content

This feature's contract is twofold: a small, additive Go change to
`internal/installer` (the recursive-install enabler, research.md), and
the content contract every `kit/skills/<name>/SKILL.md` must satisfy
(data-model.md). Both are machine-verified.

**Reconciled against the actual implementation (T025)** — three things
worth naming explicitly, none a deviation from what was drafted:

1. **`installOneFS`'s destination construction uses
   `filepath.FromSlash(r.Name)`**, not `r.Name` directly — implied by
   research.md's own path-handling decision but not spelled out in this
   draft's `ListFS`/`InstallFS` signature block below. `r.Name` is
   `fs.FS`-style (forward slashes, possibly multi-segment); converting
   it before an OS `filepath.Join` is what makes the change correct
   beyond Unix, not only coincidentally correct on it.
2. **A nested resource's `ArtifactType` is derived from its own
   basename** (`path.Base(name)`), not the full relative name — harmless
   either way for a `"skill"`-kind resource (the field goes unused
   outside template callers, per 006-agent-adapter's own doc comment),
   but worth confirming explicitly since `artifactTypeFromFilename`
   trims a `".md.tmpl"` suffix that a full nested path could otherwise
   spuriously fail to match.
3. **The structural-conformance test's fixture for the recursive-walk
   change itself uses `sourceDir "."`**, not `"skills"` — matching
   exactly how `kit.SkillsFS` (already `fs.Sub`'d) and `claude.Install`
   actually call `InstallFS` in production. An earlier draft of the test
   used `sourceDir "skills"` (matching `kit.TemplatesFS`'s own
   convention instead) and initially failed for the wrong reason — not
   an implementation bug, but a fixture that didn't match the real
   call shape being tested. Corrected before the implementation task was
   marked complete.

## `internal/installer` (extended, not replaced)

```go
package installer

// ListFS now walks source recursively under sourceDir (fs.WalkDir),
// returning one Resource per file (never a directory) found anywhere
// in the tree — not only sourceDir's immediate children. Resource.Name
// is the file's path relative to sourceDir (research.md's meaning
// change; unchanged in practice for any flat caller). Exported
// signature unchanged: ListFS(source fs.FS, sourceDir, kind string) []Resource
func ListFS(source fs.FS, sourceDir, kind string) []Resource

// InstallFS is unchanged in signature and behavior beyond following
// ListFS's now-recursive resource set — every resource's destination
// is still targetDir/sourceDir/Resource.Name (or targetDir/Resource.Name
// when sourceDir is "."), still one atomic write per resource, still
// no-silent-overwrite, still containment-checked.
func InstallFS(source fs.FS, sourceDir, kind, targetDir string, overwrite bool) ([]Outcome, error)
```

**Guarantees**:
- Every one of 005-embedded-kit's, 006-agent-adapter's, and
  008-cli-cobra's own existing tests pass unmodified — the recursive
  walk produces an identical `Resource` set for every flat fixture and
  for `kit.TemplatesFS` (research.md's verified backward-compatibility
  claim, re-confirmed by the regression suite this feature's tasks.md
  names explicitly).
- A resource nested arbitrarily deep installs at the matching nested
  destination, parent directories created automatically
  (`WriteAtomicFile`'s existing `MkdirAll`, unchanged).
- Containment (`artifacts.RelativeWithinRoot`) and no-silent-overwrite
  guarantees apply identically regardless of nesting depth.

## `kit` package (extended)

```go
package kit

// SkillsFS's embedded content changes from one placeholder file to
// nine real Skill directories (research.md) — its own exported shape
// (fs.FS, rooted at the Skills content itself via fs.Sub, established
// by 008-cli-cobra) is unchanged.
var SkillsFS fs.FS
```

## Skill content contract (data-model.md, machine-checked)

Every `kit/skills/<name>/SKILL.md`, once installed to
`.claude/skills/<name>/SKILL.md`, must satisfy:

| Check | Rule | FR |
|---|---|---|
| Frontmatter | YAML block with non-empty `name` (== directory name) and non-empty `description` | FR-006 |
| Section completeness | All 29 of §39's H2 headings present, in order | FR-002 |
| Operations allowlist | Every `misterspec internal <word>` token in "Deterministic Operations" names one of the ten real commands (data-model.md's per-Skill table) | FR-003, SC-004 |
| Completion Contract | Section text names all five of §50's required concepts | FR-004 |
| Authority sections | "Allowed Reads"/"Allowed Creates"/"Allowed Modifications"/"Forbidden Mutations" all present and non-empty (subset of section completeness, called out separately since it is FR-007's own requirement) | FR-007 |

**Guarantees**: A test suite (Polish) parses every one of the nine
installed `SKILL.md` files and asserts every row above — a structural
regression in any one Skill's content fails the build the same way a
Go compile error would, not only a human content review.

## Cross-cutting: no change to any other package's public API

`internal/operations`, `internal/validation`, `internal/bootstrap`,
`internal/cli`, `internal/agents` (including `claude`/`builtin`) are
all unmodified by this feature — every one of their existing tests is a
named regression gate, not a caller this feature's own tests need to
re-derive.
