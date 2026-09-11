# Quickstart: Structural Validation and Project Status

Still no CLI — this shows how a *future* caller uses
`internal/validation` and `operations.Status` on top of the prior three
features.

## 1. Validate a single entity

```go
findings, err := validation.ValidateEntity(proj.Root, proj.Config, "SPEC-014")
if err != nil {
    // a malformed rawID, or an unsupported target type — a usage error
}
if len(findings) == 0 {
    // SPEC-014 is structurally clean
}
for _, f := range findings {
    fmt.Println(f.Code, f.Path, f.Message)
}
```

## 2. Validate the whole project

```go
findings, err := validation.ValidateProject(proj.Root, proj.Config)
// findings aggregates every problem across every entity — including
// duplicate IDs no single-entity check could see on its own.
```

## 3. Get a status summary

```go
summary, err := operations.Status(proj.Root, proj.Config)
// summary.Counts[ids.Spec]           == 21
// summary.SpecsByState["ready"]      == 6
// summary.StructuralErrors           == len(findings) from step 2
```

## 4. A broken entity reports specific findings, not a generic failure

```go
// A Feature whose parent doesn't exist:
findings, _ := validation.ValidateEntity(proj.Root, proj.Config, "FEAT-099")
// findings[0].Code == "missing_parent"

// A project with two Specs sharing an ID under different Features:
findings, _ = validation.ValidateProject(proj.Root, proj.Config)
// one entry has Code == "duplicate_id"
```

## Validation

Validated by this feature's filesystem-integration tests (fixture
projects mixing valid and deliberately broken entities), the same way the
prior three features' quickstarts were.
