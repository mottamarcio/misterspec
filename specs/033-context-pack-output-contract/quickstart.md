# Quickstart: Validating the Context Pack output contract

This guide proves the feature works end-to-end once implemented, against a disposable temp project. It mirrors what the Go tests for this feature assert.

## Prerequisites

- A built `misterspec` binary reflecting this feature's changes (`go build ./cmd/misterspec`).
- A scratch project with `.misterspec/config.yaml` and at least one Spec whose `spec.md` has frontmatter followed by real body content (so the frontmatter-offset fix has something to prove).

## Scenario 1 — Full content delivered without re-reading (User Story 1)

```sh
misterspec internal context SPEC-014 --mode package
```

**Pass condition**: every entry in `context.items` has a non-empty `content` field whose text matches the corresponding section of `SPEC-014`'s own `spec.md`, verified without opening the file again.

```sh
misterspec internal context SPEC-014
```

**Pass condition**: the default (manifest) response has no `content` field on any item — confirming the lightweight mode stays explicitly metadata-only, never silently including partial content.

## Scenario 2 — File-absolute location (User Story 2)

```sh
misterspec internal context SPEC-014 --mode package
```

**Pass condition**: for each item, open `spec.md` locally and confirm that line `location.start_line` is exactly the first line of that item's own heading/content — not shifted by the frontmatter's own line count. Repeat against an artifact with no frontmatter and confirm the numbers are still correct (Edge Case).

## Scenario 3 — Fingerprint detects source drift (Edge Case)

```sh
misterspec internal context SPEC-014 --mode package | jq '.context.items[0].fingerprint'
# Edit that item's own source section in spec.md.
misterspec internal context SPEC-014 --mode package | jq '.context.items[0].fingerprint'
```

**Pass condition**: the fingerprint differs after the edit; it is identical across two calls with no edit in between.

## Scenario 4 — Versioned envelope, no duplication (User Story 3)

```sh
misterspec internal context SPEC-014 --mode package
```

**Pass condition**: `context.schema_version` is present and equals `1`; `context.rendered` is `null`; no item's `content` text also appears inside a `rendered` field in the same response (there is none to duplicate into).

```sh
misterspec internal context SPEC-014 --mode package --render
```

**Pass condition**: the command fails with `{"ok":false,"error":{"code":"invalid_argument",...}}`, exit code 2 — the two are never silently combined.

## Scenario 5 — Backward compatibility (User Story 3, spec FR-008)

```sh
misterspec internal context SPEC-014
misterspec internal context SPEC-014 --render
```

**Pass condition**: both outputs are identical to a captured pre-change baseline, field-for-field, except for the two new additive fields (`schema_version`, `diagnostics.payload_tokens`) — every existing field name, value, and the absence of `content`/`location`/`fingerprint` is unchanged.

## Scenario 6 — Output-size diagnostics reflect what was returned (spec FR-007)

```sh
misterspec internal context SPEC-014 --mode package | jq '.context.diagnostics'
```

**Pass condition**: `diagnostics.payload_tokens >= diagnostics.tokens_selected` (package mode's per-item metadata adds real size beyond raw content), and `tokens_selected` itself is unchanged from what the same request would have reported in manifest mode.

## Cleanup

```sh
rm -rf /tmp/ms-quickstart-033
```
