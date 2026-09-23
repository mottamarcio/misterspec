# Contract: Architecture Rules & Code Context (`check-architecture`, `prepare` extension)

This documents the two new deterministic capabilities this feature
adds (Constitution Principle IX): a new command,
`internal check-architecture`, and an additive extension to the
existing `internal prepare` response.

## 1. `internal/gosource` (new leaf package)

```go
package gosource

// Imports returns path's own import list (path strings exactly as
// written, e.g. "github.com/mottamarcio/misterspec/internal/ids"),
// parsed via go/parser's ImportsOnly mode — cheap, no type-checking
// (research.md #1/#2).
func Imports(path string) ([]ImportRef, error)

// ImportRef is one import statement's own path and file-absolute line
// (data-model.md).
type ImportRef struct {
	Path string
	Line int
}

// Declarations returns every top-level func/type/const/var
// declaration in path, via go/parser's ParseComments mode — signature
// text, doc comment, and file-absolute line range for each
// (research.md #1/#2, contracts §2).
func Declarations(path string) ([]Declaration, error)

// Declaration is one top-level declaration's own parsed shape
// (data-model.md "CodeDeclaration").
type Declaration struct {
	Name      string
	Kind      string // "func" | "type" | "const" | "var"
	Signature string // signature text + doc comment
	Body      string // full declaration text
	StartLine int
	EndLine   int
}
```

Both functions read exactly the one file they're given — no directory
walking, no dependency on any other `internal/*` package (research.md
#1's "leaf" role).

## 2. `internal/architecture` (new package, Facet A)

```go
package architecture

// Rule mirrors project.Configuration's own ArchitectureRule
// (data-model.md).
type Rule struct {
	Kind     string // "forbidden_dependency" | "layer_boundary" | "required_contract"
	From     string
	To       string
	Contract string
}

// Result is one rule's own evaluation outcome (data-model.md "Result").
type Result struct {
	RuleIndex int
	Status    string // "pass" | "fail" | "not_evaluated"
	Path      string
	Line      int
	Message   string
	Reason    string // set only when Status == "not_evaluated"
}

// Report is CheckArchitecture's own full result (data-model.md
// "ArchitectureCheckReport").
type Report struct {
	Adapter string // "go", or "" when no adapter applies
	Results []Result
}

// CheckArchitecture evaluates rules against root's own source tree.
// When root has no go.mod, every rule in rules produces exactly one
// Result with Status "not_evaluated" and Reason "no_adapter_for_project"
// (research.md #8, spec FR-004). exclusions are project.Configuration's
// own CodeExclusions glob patterns (research.md #4) — a matching path
// is never walked, regardless of whether it would otherwise violate a
// rule.
func CheckArchitecture(root string, rules []Rule, exclusions []string) (Report, error)
```

## 3. `internal/context/index` (extended, Facet B)

```go
package index

// CodeFile / CodeDeclaration mirror the new code_files/
// code_declarations tables (data-model.md).
type CodeFile struct {
	Path        string
	Fingerprint string
}

type CodeDeclaration struct {
	Path      string // owning CodeFile's own Path
	Name      string
	Kind      string
	Signature string
	Body      string // "" when not indexed at full-body tier
	StartLine int
	EndLine   int
	IsTest    bool
}

// Store (extended): two new methods, alongside SavePack/LookupPack
// (043).
type Store interface {
	// ... existing methods unchanged ...

	// SyncCode incrementally reconciles code_files/code_declarations
	// with root's current *.go files (excluding exclusions),
	// mirroring Sync's own new/changed/deleted reconciliation for
	// Markdown (research.md #6).
	SyncCode(root string, exclusions []string) (SyncReport, error)

	// DeclarationsForFiles returns every CodeDeclaration whose own
	// Path is in paths, plus every declaration from that path's own
	// associated _test.go file(s) in the same directory (spec FR-007).
	DeclarationsForFiles(paths []string) ([]CodeDeclaration, error)
}
```

`schemaVersion` bumps; `code_files`/`code_declarations` join the
existing `dropStatements`/`schemaStatements` pair — no separate
migration path (data-model.md, research.md #6).

## 4. `internal/prepare` (extended, Facet B)

```go
package prepare

// TaskPreparation (extended) — see data-model.md.
type TaskPreparation struct {
	// ... existing fields unchanged ...
	CodeContext        []contextengine.PackageItem
	CodeScopeNotFound  []string
}

// ResolveCodeContext parses fields.Scope into file paths, resolves
// each against store's own code index (DeclarationsForFiles),
// includes every declaration at signature tier always, promotes a
// declaration to full-body tier when its own owning file's total
// estimated size (via estimator) is at or under maxInlineCodeBodySize
// (research.md #7), and reports any Scope-named path that resolves to
// no indexed file in notFound rather than dropping it silently (spec
// Edge Case).
func ResolveCodeContext(store index.Store, scope string, estimator artifacts.Estimator) (items []contextengine.PackageItem, notFound []string, err error)
```

## 5. CLI: `misterspec internal check-architecture`

```text
misterspec internal check-architecture [--dir <project-dir>]
```

No required flags beyond the universal `--dir`; rules and exclusions
come from the project's own `.misterspec/config.yaml`.

### Success envelope

```json
{
  "ok": true,
  "architecture": {
    "adapter": "go",
    "results": [
      {
        "rule_index": 0,
        "status": "fail",
        "path": "internal/artifacts/parser.go",
        "line": 12,
        "message": "internal/artifacts imports internal/cli/internalcmd, which rule 0 forbids",
        "reason": ""
      },
      {
        "rule_index": 1,
        "status": "not_evaluated",
        "path": "",
        "line": 0,
        "message": "rule kind \"required_contract\" is not supported by the go adapter",
        "reason": "rule_kind_unsupported"
      }
    ]
  }
}
```

An empty `architecture_rules` config section still returns `ok: true`
with `results: []` — zero rules declared is not an error.

### Error envelope

Reuses `{"ok": false, "error": {"code", "message"}}`. A declared rule
with an unrecognized `kind` or an empty `from` is reported by wrapping
the already-existing `project.ErrInvalidConfiguration` sentinel
(`internal/project/config.go`) — the same condition every other
malformed `.misterspec/config.yaml` field already reports through
`project.Load`, not a new, parallel error path invented just for this
one field (Constitution Principle IV/VI). No new error code is added
by this feature.

## 6. CLI: `misterspec internal prepare` (additive response fields)

```json
{
  "ok": true,
  "preparation": {
    "...": "... every existing field unchanged ...",
    "code_context": [
      {
        "path": "internal/auth/refresh.go",
        "heading": "RotateRefreshToken",
        "content": "// RotateRefreshToken ...\nfunc RotateRefreshToken(...) error",
        "location": {"path": "internal/auth/refresh.go", "start_line": 40, "end_line": 40},
        "fingerprint": "sha256:9c1a..."
      }
    ],
    "code_scope_not_found": []
  }
}
```

A Task with an empty or all-non-`.go` `Scope:` gets `code_context: []`
(never omitted, never `null`) and `code_scope_not_found: []`.

## 7. Behavioral guarantees this contract makes (traceable to spec FRs)

- FR-003/FR-005: every `fail` `Result` carries a `path`+`line`
  reproducible across repeated runs on unchanged code; `not_evaluated`
  is never rendered as, or confused with, `pass` — the three `Status`
  values are structurally distinct strings, never inferred from
  absence.
- FR-004: a project with no `go.mod` produces `adapter: ""` and every
  rule as `not_evaluated`/`reason: "no_adapter_for_project"` — never a
  `pass`.
- FR-007/FR-008: `code_context` always includes Scope-declared
  declarations at signature tier at minimum; a declaration's own
  `content` is its full body only when `ResolveCodeContext`'s fixed
  threshold allows it — never unconditionally.
- FR-009: every `code_context` entry carries its own `fingerprint`,
  computed the same way 033 already fingerprints any other
  `PackageItem`.
