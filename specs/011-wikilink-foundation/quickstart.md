# Quickstart: Wikilink Graph Foundation

No CLI change — this shows how a Spec artifact's author uses the new
syntax, and what `misterspec internal validate` now catches.

## 1. Author an explicit link

```markdown
<!-- ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md -->

## Relevant Knowledge

- [[KNOW-003|Authentication Model]]
- [[KNOW-008]]
```

Nothing installs this automatically — an author (human or coding agent)
writes it deliberately, the same way any other prose is written.

## 2. Extraction (Go-level, no CLI yet)

```go
body, _ := artifacts.ReadBody("ai/.../SPEC-014/spec.md")
links, _ := artifacts.ExtractWikiLinks(body)
// links == []WikiLink{
//   {Target: "KNOW-003", Alias: "Authentication Model", Line: 3},
//   {Target: "KNOW-008", Alias: "",                      Line: 4},
// }
```

## 3. Validation catches a broken link

```markdown
## Relevant Knowledge

- [[KNOW-999]]
```

```bash
$ misterspec internal validate SPEC-014
{"ok":true,"valid":false,"findings":[
  {"code":"broken_wikilink","severity":"error",
   "path":"ai/.../SPEC-014/spec.md",
   "message":"link target KNOW-999 does not resolve to an existing artifact"}
]}
$ echo $?
4
```

Same command, same JSON shape, same exit-code convention
(008-cli-cobra) — nothing new to learn.

## 4. A malformed link is distinguished from a broken one

```markdown
- [[not-a-real-id]]
```

```json
{"code":"invalid_wikilink","severity":"error", ...}
```

## 5. An artifact with no links validates exactly as before

```bash
$ misterspec internal validate SPEC-001   # no links anywhere in its body
{"ok":true,"valid":true,"findings":[]}
```

Identical to this project's behavior before this feature existed — no
opt-in required, no new failure mode for artifacts that never use a
link (US3).

## Validation

Validated by: `internal/artifacts`'s own new tests (extraction, code-
span/fenced-block exclusion, `ReadBody`, `ParseMetadata` regression);
`internal/validation`'s new tests (the three codes, plus its full
004-structural-validation suite re-run unmodified as the explicit
non-regression gate for US3); `internal/cli/internalcmd`'s existing
`validate` tests re-run unmodified, confirming the new codes flow
through the already-generic JSON mapping with zero code changes there.
