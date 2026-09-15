# Phase 1 Contracts: Agent Adapter Package APIs

Like the prior features, no HTTP/CLI surface yet. This feature's contract
is: new exported surface on the existing `internal/installer` package
(additive — see `internal/installer`'s own contract in
005-embedded-kit for what's unchanged), and three new packages
(`internal/agents`, `internal/agents/claude`, `internal/agents/builtin`).

**Reconciled against the actual implementation (T019)** — four things
discovered during implementation, not present in the original draft
below:

1. **`InstallFS`/`ListFS` gained a `kind` parameter.** The original draft
   omitted it, but `Resource.Kind` was hardcoded `"template"` inside the
   old `List()` — a generalized function needs the caller to say what
   kind of resource it's listing/installing (`"template"` vs. this
   feature's `"skill"`), or every non-template caller would get
   mislabeled resources.
2. **`internal/installer`'s private `writeAtomicFile` is now exported as
   `WriteAtomicFile`.** `agents.RecordInstall` needed the identical
   atomic-write recipe to write `install.json`, and — being a different
   package — could not call an unexported function. This mirrors
   `artifacts.RelativeWithinRoot`'s export story in 002-read-operations.
3. **A real bug, found and fixed**: `installOneFS` originally built an
   `fs.FS` read path via raw string concatenation
   (`sourceDir+"/"+r.Name`). With `sourceDir == "."` (exactly what
   `claude.Install` passes, since `req.Skills` is already scoped to the
   Skills directory itself) this produced `"./name"`, which `io/fs`
   rejects as an invalid path — `fstest.MapFS` and `os.DirFS` both
   correctly refused it. Fixed by using `path.Join(sourceDir, r.Name)`
   (the `path` package's join, which collapses `"." + "/" + "name"` to
   `"name"` — not `path/filepath`, since `fs.FS` paths are always
   forward-slash `path`-style regardless of host OS). This was caught by
   `TestClaude_InstallFreshDirectory` actually failing, not discovered by
   inspection — the value of writing the integration test first.
4. **User Story 3 turned out not to require User Story 2 after all.**
   `agents.RecordInstall`/`CurrentInstall` operate on an `InstallResult`
   value, not on a real `Adapter` — so their own tests build a fixture
   `InstallResult` directly rather than needing `claude.New().Install(...)`
   to have run. The full real-adapter round trip is still covered (by
   `TestClaude_InstallFreshDirectory` in `claude_test.go`, which does
   call the real adapter and checks the resulting `install.json`) — it's
   just not `CurrentInstall`'s own tests' only coverage, a looser (and
   arguably better-isolated) coupling than tasks.md predicted.

## `internal/installer` (extended) — prerequisite for User Story 2

```go
package installer

// InstallFS is the generalized form of Install: source, sourceDir, and
// kind replace the hardcoded kit.TemplatesFS/"templates"/"template".
// Install(targetDir, overwrite) is now literally
// InstallFS(kit.TemplatesFS, "templates", "template", targetDir, overwrite)
// — unchanged behavior, unchanged exported signature.
func InstallFS(source fs.FS, sourceDir, kind, targetDir string, overwrite bool) ([]Outcome, error)

// ListFS is the generalized form of List.
func ListFS(source fs.FS, sourceDir, kind string) []Resource

// WriteAtomicFile is 005-embedded-kit's atomic-write recipe, exported so
// internal/agents can reuse it for install.json rather than a third
// implementation (item 2 above).
func WriteAtomicFile(targetAbs string, content []byte) error
```

## `internal/agents` — User Story 1: Discovery & Selection

```go
package agents

// Adapter is one coding-agent integration
// (docs/architecture-specification.md §34).
type Adapter interface {
    ID() string
    Name() string
    TargetPath() string
    Install(ctx context.Context, req InstallRequest) (InstallResult, error)
}

// Registry is an immutable-after-construction lookup of adapters —
// never global mutable state (research.md).
type Registry struct{ /* unexported */ }

func NewRegistry(adapters ...Adapter) *Registry

// List returns every registered adapter, sorted by ID for deterministic
// output. No filesystem access (FR-002).
func (r *Registry) List() []Adapter

// Get returns (adapter, true) for a registered id, or (nil, false) —
// never an error, never an arbitrary default — for one that isn't
// (FR-003).
func (r *Registry) Get(id string) (Adapter, bool)
```

**Guarantees**: `List()` performs no filesystem access; two `Registry`
instances built from the same adapters produce identical `List()` output
(sorted, deterministic).

## `internal/agents` — User Story 2: Install & Record

```go
package agents

// InstallRequest is the input to Adapter.Install.
type InstallRequest struct {
    ProjectRoot string
    Skills      fs.FS  // caller-supplied; kit.SkillsFS once Phase 6 exists
    Overwrite   bool
}

// InstallResult is the in-memory outcome of one Install call (FR-006).
type InstallResult struct {
    AdapterID       string
    IntegrationPath string
    Outcomes        []installer.Outcome // reused directly, not duplicated
}

// InstallRecord mirrors docs/architecture-specification.md §22's
// install.json schema exactly.
type InstallRecord struct {
    SchemaVersion     int    `json:"schema_version"`
    MisterspecVersion string `json:"misterspec_version"`
    Agent             struct {
        ID              string `json:"id"`
        IntegrationPath string `json:"integration_path"`
    } `json:"agent"`
}

// RecordInstall writes result as install.json, atomically. Exported —
// not the "recordInstall" internal helper the original draft assumed —
// because each concrete Adapter lives in its own package (e.g.
// internal/agents/claude) and must be able to call it (item 3 above's
// export story, same reasoning).
func RecordInstall(projectRoot string, result InstallResult) error

// CurrentInstall reads a project's installation record. It returns
// (record, true, nil) if one exists and parses, (zero, false, nil) if
// none exists yet — a distinct "not installed" result, not an error
// (FR-007) — or (zero, false, err) only for a genuine read/parse
// failure distinct from simple absence.
func CurrentInstall(projectRoot string) (InstallRecord, bool, error)
```

**Guarantees**:
- `Adapter.Install` materializes every Skill resource under the same
  atomic-write and no-silent-overwrite guarantees `internal/installer`
  already proves (FR-005, SC-002, SC-005), then writes `install.json`
  atomically (FR-006).
- An adapter never writes outside its own `TargetPath()` plus
  `.misterspec/install.json` — no path into a project's `ai/` artifact
  tree exists in this feature at all (FR-008, mirroring
  005-embedded-kit's identical boundary).
- `CurrentInstall` never re-installs, never re-derives its answer by
  scanning `TargetPath()`'s contents — it only reads the record
  `Install` wrote (FR-007).

## `internal/agents/claude` & `internal/agents/builtin`

```go
package claude

// New constructs the Claude Code adapter: ID() == "claude-code",
// TargetPath() == ".claude/skills" — this project's own established
// convention.
func New() agents.Adapter
```

```go
package builtin

// Default returns a Registry pre-populated with every concrete adapter
// this build knows about (today: Claude Code only). A separate package
// from agents/claude specifically to avoid the import cycle
// agents ↔ claude would otherwise create (research.md).
func Default() *agents.Registry
```

## Cross-cutting: error/field → future JSON mapping

Extends the running convention from 001-005. This feature introduces no
new sentinel errors — `Registry.Get`'s "not found" and
`CurrentInstall`'s "not installed" are both `bool`-signaled, matching
Go idiom for an expected, non-exceptional absence rather than an error
condition:

| Condition | Future JSON shape |
|---|---|
| `Registry.Get(id)` → `(nil, false)` | `{"ok": true, "adapter": null}` — a successful query that found nothing, not a failure. |
| `CurrentInstall` → `(_, false, nil)` | `{"ok": true, "installed": false}` — likewise. |
| `InstallRecord` | Serializes directly to §22's `install.json` shape via its `json:` tags — no translation layer needed when the CLI layer exists. |
