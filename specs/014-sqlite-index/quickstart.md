# Quickstart: Disposable SQLite Index

Builds directly on 011 (wikilinks), 012 (reference graph), and 013
(document/chunk model).

## 1. Build the index for the first time

```go
store, _ := index.Open(".misterspec/cache/context.db")
defer store.Close()

report, _ := store.Sync(root, cfg)
// report.Indexed  == number of eligible artifacts found
// report.Updated  == 0 (nothing existed before)
// report.Skipped  == 0
// report.Removed  == 0
```

## 2. Search it

```go
results, _ := store.Search("refresh token rotation", 10)
// results[0].Path    == "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md"
// results[0].Heading == "Intent"
```

## 3. Change one artifact, then synchronize incrementally

```markdown
<!-- SPEC-014/spec.md edited: one sentence changed in "Intent" -->
```

```go
report, _ := store.Sync(root, cfg)
// report.Updated == 1 (only SPEC-014)
// report.Skipped  == every other previously-indexed artifact,
//                    provably untouched (SC-003)
```

## 4. Delete an artifact, then synchronize

```bash
$ rm -rf ai/.../SPEC-999/
```

```go
report, _ := store.Sync(root, cfg)
// report.Removed == 1
```

## 5. Look up a relationship from the index

```go
outgoing, _ := store.Outgoing("SPEC-014")
// outgoing contains the exact same entries
// operations.References(root, cfg, "SPEC-014") would report (FR-009)
```

## 6. Discard and rebuild from scratch

```go
report, _ := store.Rebuild(root, cfg)
// Every artifact is treated as new; report.Indexed == the full corpus.
// A subsequent Search for the same term returns the same underlying
// content as before the rebuild (SC-004).
```

## 7. An artifact that fails to parse doesn't stop the run

```go
report, _ := store.Sync(root, cfg)
// report.Errors == []SyncError{{Path: "ai/knowledge/KNOW-005-x.md", Err: ...}}
// Every other artifact was still indexed (FR-010).
```

## Validation

Validated by: `internal/context/index`'s new tests covering
`docs/context-engine-implementation.md` §29.5's own list (initial
build, unchanged sync, changed file, new file, deleted file,
transaction rollback, database rebuild, FTS synchronization, incoming/
outgoing links); 011's, 012's, and 013's own full suites re-run
unmodified, since this feature adds no call site requiring any change
to them.
