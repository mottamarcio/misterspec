# Quickstart: Internal Context Command

Builds directly on 015/016's own fixture: SPEC-014 depends on SPEC-011,
links to KNOW-003, the project has a Constitution, and a free-text query
also matches an unrelated chunk elsewhere.

## 1. First invocation against a project with no cache yet

```sh
$ misterspec internal context SPEC-014 \
    --intent implementation \
    --query "refresh token rotation"
```

Behind the scenes (User Story 3): `.misterspec/cache/context.db` does
not exist yet → `index.Open` creates it and its schema → `store.Sync`
indexes every eligible artifact for the first time (every artifact
looks "new") → `Collect`/`Rank`/`ApplyBudget` run exactly as they would
against a warm cache. The command still returns a correct, successful
result on this very first call — no separate "build the index first"
step is ever required.

```json
{"ok": true, "context": {"target": "SPEC-014", "intent": "implementation", "budget": 6000, "estimated_tokens": 4382, "budget_exceeded": false, "overage": 0, "items": [...], "diagnostics": {...}, "rendered": null}}
```

## 2. Repeating the same request — cheap, identical output

```sh
$ misterspec internal context SPEC-014 --intent implementation --query "refresh token rotation"
```

`store.Sync` finds every artifact's fingerprint unchanged and skips
reindexing all of them; `Collect`/`Rank`/`ApplyBudget` produce
byte-for-byte identical `items`/`diagnostics` (FR-009).

## 3. An unknown target

```sh
$ misterspec internal context SPEC-999
```

```json
{"ok": false, "error": {"code": "entity_not_found", "message": "..."}}
```

Exit code 3. No `context` key is present.

## 4. An unsupported intent

```sh
$ misterspec internal context SPEC-014 --intent bogus
```

```json
{"ok": false, "error": {"code": "unsupported_intent", "message": "contextengine: unsupported intent: bogus"}}
```

Exit code 2.

## 5. A non-numeric budget

```sh
$ misterspec internal context SPEC-014 --budget notanumber
```

```json
{"ok": false, "error": {"code": "invalid_argument", "message": "..."}}
```

Exit code 2. (A parseable negative budget, e.g. `--budget -1`, is
**not** an error — see step 6.)

## 6. A deliberately tiny budget — mandatory content still wins

```sh
$ misterspec internal context SPEC-014 --budget 10
```

```json
{"ok": true, "context": {"target": "SPEC-014", "budget": 10, "budget_exceeded": true, "overage": 812, "items": [/* Constitution + SPEC-014 chunks only, in full */], ...}}
```

Exactly 016's own already-tested behavior (FR-006/FR-007), now reachable
through the CLI.

## 7. Rendered mode

```sh
$ misterspec internal context SPEC-014 --intent implementation --render
```

```json
{"ok": true, "context": {"target": "SPEC-014", ..., "rendered": "# Context for SPEC-014 (implementation)\n\n## Mandatory\n\n### ai/.../constitution.md\n...\n\n## Structural\n\n### ai/.../SPEC-011/spec.md — Intent\n...\n"}}
```

Every item present in `items[]` also appears in `rendered`, grouped
under the same `Tier.String()` headings, in the same order — `--render`
never changes selection, ranking, or budgeting (FR-011).
