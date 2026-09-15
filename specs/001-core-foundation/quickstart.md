# Quickstart: Core Repository Foundation

This feature has no CLI yet — it's a foundation library. This quickstart
shows how a *future* caller (the Phase 2 `operations` layer, or a test)
will use it, to sanity-check that the package design in `contracts/` and
`data-model.md` actually composes.

## 1. Detect the project

```go
proj, err := project.Detect(cwd)
switch {
case errors.Is(err, project.ErrNotInitialized):
    // handle "no misterspec project here"
case errors.Is(err, project.ErrInvalidConfiguration):
    // handle "config.yaml exists but is broken"
case err != nil:
    // unexpected filesystem error
default:
    // proj.Root, proj.Config are ready to use
}
```

## 2. Resolve a canonical path

```go
specID, _ := ids.Parse(ids.Spec, "SPEC-014", proj.Config.IDWidth)
featID, _ := ids.Parse(ids.Feature, "FEAT-004", proj.Config.IDWidth)

path, err := artifacts.ResolvePath(proj.Root, proj.Config, artifacts.TypeSpec, specID, &featID)
// path.Directory == "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014"
// path.File      == "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md"
```

Note the program parent (`PRG-001`) — a Spec's full ancestry is resolved via
its declared `Parent` chain during path resolution, not guessed.

## 3. Parse an artifact's metadata

```go
meta, err := artifacts.ParseMetadata(filepath.Join(proj.Root, path.File))
switch {
case errors.Is(err, artifacts.ErrArtifactNotFound):
    ...
case errors.Is(err, artifacts.ErrFrontmatterMalformed):
    ...
case errors.Is(err, artifacts.ErrRequiredFieldMissing):
    ...
default:
    // meta.ID, meta.Status, meta.DependsOn, ... are populated
}
```

## 4. Discover existing IDs and compute the next one

```go
result, err := ids.Scan(proj.Root, proj.Config, ids.Spec)
// result.IDs: every valid SPEC-* found
// result.Duplicates: any SPEC-* number claimed by more than one artifact

next := ids.NextID(result.IDs, ids.Spec, proj.Config.IDWidth)
// next == SPEC-<max+1>, derived purely from what Scan found — no counter
// file was read or written to compute it.
```

## Validation

This quickstart is validated by the feature's filesystem-integration tests
(fixture project trees under `t.TempDir()`), not by a live CLI, since no
CLI exists yet in this feature's scope.
