# Quickstart: References and Backlinks

Builds directly on 011-wikilink-foundation's example: SPEC-014 depends
on SPEC-011 and links to KNOW-003.

## 1. Query what SPEC-014 points to

```bash
$ misterspec internal references SPEC-014
{"ok":true,"target":"SPEC-014","references":{
  "formal":[
    {"relation":"parent","target":"FEAT-004"},
    {"relation":"depends_on","target":"SPEC-011"}
  ],
  "semantic":[
    {"relation":"wikilink","target":"KNOW-003"}
  ]
}}
```

Nothing new to author — this reads exactly the frontmatter and wikilinks
that already exist on disk.

## 2. Query what points at SPEC-011

```bash
$ misterspec internal backlinks SPEC-011
{"ok":true,"target":"SPEC-011","backlinks":{
  "formal":[
    {"relation":"depends_on","source":"SPEC-014"}
  ],
  "semantic":[]
}}
```

## 3. A broken or ambiguous wikilink contributes nothing

```markdown
<!-- SPEC-014/spec.md's body -->
- [[KNOW-999]]   <!-- broken: doesn't exist -->
```

```bash
$ misterspec internal references SPEC-014
{"ok":true,"target":"SPEC-014","references":{"formal":[...],"semantic":[{"relation":"wikilink","target":"KNOW-003"}]}}
```

`KNOW-999` simply doesn't appear — 011's own `internal validate SPEC-014`
is still where that problem is reported, as `broken_wikilink`.

## 4. An artifact nothing points to

```bash
$ misterspec internal backlinks KNOW-999-unused
{"ok":true,"target":"KNOW-999","backlinks":{"formal":[],"semantic":[]}}
```

Well-formed and empty — never an error (FR-006).

## 5. An unsupported target is rejected consistently

```bash
$ misterspec internal references TASK-001
{"ok":false,"error":{"code":"invalid_target","message":"operations: invalid target: ... is not a standalone-referenceable entity type"}}
$ echo $?
2
```

Same code, same exit convention `inspect`/`children` already use for
their own unsupported-target cases (FR-008) — nothing new to learn.

## Validation

Validated by: `internal/ids`'s new `ResolveTarget` test (mirroring
011's own three-way classification); `internal/validation`'s full
011 suite re-run unmodified after `classifyWikilink`'s refactor;
`internal/operations`'s new `References`/`Backlinks` tests (formal only,
semantic only, both, empty, self-reference, duplicate reference, broken/
malformed/ambiguous exclusion, unsupported-target rejection, ordering);
`internal/cli/internalcmd`'s new command tests (JSON shape, exit codes);
004-structural-validation's and 008-cli-cobra's own full suites re-run
unmodified as the explicit non-regression gate (US3, SC-003).
