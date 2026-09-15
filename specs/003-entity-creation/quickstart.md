# Quickstart: Atomic Entity Creation

Still no CLI — this shows how a *future* caller uses `internal/operations`'s
new `Create`/`CreateArtifact` on top of the previous two features.

## 1. Create a Program, then a Feature under it

```go
prg, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
    Type: ids.Program,
})
// prg.ID.String() == "PRG-001" (first Program in this project)
// prg.Path == "ai/programs/PRG-001/program.md", already written and valid

feat, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
    Type:   ids.Feature,
    Parent: prg.ID.String(),
})
// feat.ID.String() == "FEAT-001"
```

## 2. Create a Spec under an invalid parent — rejected

```go
_, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
    Type:   ids.Spec,
    Parent: "FEAT-999", // does not exist
})
// errors.Is(err, operations.ErrInvalidParent) == true
// nothing was written to disk
```

## 3. Create a Plan for a Spec

```go
plan, err := operations.CreateArtifact(proj.Root, proj.Config, operations.CreateArtifactRequest{
    Kind: artifacts.TypePlan,
    For:  "SPEC-014",
})
// plan.Path == ".../specs/SPEC-014/plan.md"

// A second Plan for the same Spec is rejected:
_, err = operations.CreateArtifact(proj.Root, proj.Config, operations.CreateArtifactRequest{
    Kind: artifacts.TypePlan,
    For:  "SPEC-014",
})
// errors.Is(err, operations.ErrAlreadyExists) == true
```

## 4. Two concurrent creates never collide

```go
var wg sync.WaitGroup
results := make([]operations.CreateResult, 2)
for i := range results {
    wg.Add(1)
    go func(i int) {
        defer wg.Done()
        results[i], _ = operations.Create(proj.Root, proj.Config, operations.CreateRequest{
            Type:   ids.Spec,
            Parent: "FEAT-001",
        })
    }(i)
}
wg.Wait()
// results[0].ID and results[1].ID are different, sequential SPEC-* IDs
```

## 5. Recover from a stale lock automatically

```go
// Simulate a crashed prior attempt: a lock file older than the staleness
// threshold, left behind.
testutil.WriteFile(t, root, ".misterspec/.lock", oldTimestamp)

// The very next Create still succeeds — no manual cleanup needed.
result, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{Type: ids.Program})
```

## Validation

Validated by this feature's filesystem-integration and concurrency tests,
the same way the prior two features' quickstarts were — no CLI exists yet
to drive this manually.
