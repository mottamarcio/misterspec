# Quickstart: Read-Only Deterministic Operations

Still no CLI — this shows how a *future* caller (the eventual
`misterspec internal ...` CLI layer, or a test) uses `internal/operations`
on top of `001-core-foundation`.

## 1. Resolve and inspect by ID alone

```go
loc, err := operations.Resolve(proj.Root, proj.Config, "SPEC-014")
switch {
case errors.Is(err, operations.ErrEntityNotFound):
    // no such Spec anywhere in the project
case errors.Is(err, operations.ErrEntityAmbiguous):
    var amb *operations.AmbiguousIDError
    errors.As(err, &amb)
    // amb.Locations lists every claiming path
case err != nil:
    // unexpected failure
default:
    // loc.Path is the exact canonical path — never guessed
}

result, err := operations.Inspect(proj.Root, proj.Config, "SPEC-014")
// result.Metadata.Status, .Parent, .DependsOn, .Supersedes are populated
```

## 2. Walk structural relationships

```go
parent, err := operations.Parent(proj.Root, proj.Config, "SPEC-014")
if parent.HasParent {
    // parent.Parent.ID == FEAT-004
}

featureType := ids.Spec
children, err := operations.Children(proj.Root, proj.Config, "FEAT-004", &featureType)
// children lists every Spec nested under FEAT-004 — and only those
```

## 3. Discover files and verify content

```go
files, err := operations.Inventory(proj.Root, proj.Config.RawDir)
// files lists every file under ai/raw, or an empty slice if none exist yet

fp, err := operations.Fingerprint(proj.Root, "ai/raw/architecture.pdf")
// fp.String() == "sha256:...", identical every time for the same content
```

## Validation

Validated by this feature's filesystem-integration tests (fixture project
trees), the same way 001-core-foundation's quickstart was — no CLI exists
yet to drive this end-to-end manually.
