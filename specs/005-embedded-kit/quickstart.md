# Quickstart: Embedded Kit and Resource Installer

Still no CLI — this shows how a *future* caller (eventually
`misterspec init`) uses `internal/installer` on top of the relocated
`kit` package.

## 1. Discover what the kit provides — no filesystem writes

```go
resources := installer.List()
for _, r := range resources {
    fmt.Println(r.Kind, r.ArtifactType, r.Name)
}
// template  program     program.md.tmpl
// template  feature     feature.md.tmpl
// ...
```

## 2. Install into a fresh directory

```go
outcomes, err := installer.Install(targetDir, false)
for _, o := range outcomes {
    // every o.Status == installer.Installed; every file byte-for-byte
    // identical to its embedded source
}
```

## 3. Re-running without overwrite changes nothing

```go
outcomes, _ = installer.Install(targetDir, false)
// every o.Status == installer.Skipped; nothing on disk changed
```

## 4. Re-running with overwrite replaces everything, still atomically

```go
outcomes, _ = installer.Install(targetDir, true)
// every o.Status == installer.Installed again
```

## 5. Artifact creation still works, sourced from the same kit

```go
result, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{Type: ids.Program})
// result's content is rendered from kit.TemplatesFS via
// internal/templates — the identical source Install materializes raw
// copies of, not a separate copy (FR-008)
```

## Validation

Validated by this feature's filesystem-integration tests (fresh and
already-populated target directories, traversal rejection) plus
003-entity-creation's own existing test suite re-run unmodified — the
same regression discipline every prior feature has used.
