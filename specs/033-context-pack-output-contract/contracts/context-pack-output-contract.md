# Contract: `internal context` output modes and versioned envelope

This documents the machine-readable contract this feature adds to `misterspec internal context <id>` (Constitution Principle IX). It extends, never replaces, `specs/017-internal-context-command/contracts/context-command.md`'s existing default/`--render` contract.

## 1. `--mode` flag

New flag on `internal context <id>`: `--mode {manifest|package|markdown}`. Default: `manifest` (unset is identical to today's no-flag behavior — research.md Decision 2).

| `--mode` | `--render` | Behavior |
|---|---|---|
| unset or `manifest` | unset | **Unchanged**: today's exact default shape. |
| unset or `manifest` | set | **Unchanged**: today's exact `--render` shape (metadata `items[]` + `rendered` Markdown). |
| `package` | unset | New: `items[]` carry full content, location, fingerprint; `rendered` is `null`. |
| `package` | set | **Rejected**: `invalid_argument` — "`--render` cannot be combined with `--mode=package`" (research.md Decision 3). |
| `markdown` | either | New: identical to `--mode unset --render` (an explicit way to name the same request; `--render`'s own presence/absence has no additional effect once `--mode=markdown` is given). |

## 2. Envelope (every mode)

```json
{
  "ok": true,
  "context": {
    "schema_version": 1,
    "target": "SPEC-014",
    "intent": "implementation",
    "budget": 6000,
    "estimated_tokens": 589,
    "budget_exceeded": false,
    "overage": 0,
    "items": [ /* shape depends on mode — see §3/§4 */ ],
    "diagnostics": {
      "candidates_considered": 18,
      "items_selected": 9,
      "tokens_available": 1200,
      "tokens_selected": 589,
      "tokens_excluded": 611,
      "reduction_percent": 61.7,
      "payload_tokens": 742
    },
    "rendered": null
  }
}
```

`schema_version` and `diagnostics.payload_tokens` are the only two fields added to the envelope itself; every other field name and meaning is unchanged from today (data-model.md "ContextPackEnvelope").

## 3. `manifest` mode item shape (unchanged)

```json
{"path": "ai/.../SPEC-014/spec.md", "heading": "R1 — First", "tier": "mandatory", "reasons": ["target"], "score": 100, "tokens": 107}
```

Byte-for-byte identical to today's `renderContextItems` output (`context.go:134-141`).

## 4. `package` mode item shape (new)

```json
{
  "path": "ai/.../SPEC-014/spec.md",
  "heading": "R1 — First",
  "tier": "mandatory",
  "reasons": ["target"],
  "score": 100,
  "tokens": 107,
  "content": "### R1 — First\n\nThe system MUST ...\n",
  "location": {"path": "ai/.../SPEC-014/spec.md", "start_line": 14, "end_line": 18},
  "fingerprint": "sha256:3a7bd3e2360a3d..."
}
```

Every `manifest`-mode field is present unchanged, plus `content`, `location`, and `fingerprint` (data-model.md "PackageItem"). `location.start_line`/`end_line` are file-absolute (research.md Decision 1) — opening the file at `start_line` lands exactly on this item's own first line, frontmatter included in the count.

## 5. `markdown` mode (new name for `--render`'s existing shape)

`items[]` is `manifest`-shaped (no `content`); `rendered` carries the single Markdown string exactly as `contextengine.Render` already produces it — no change to `render.go`'s own output.

## 6. Error case

```json
{"ok": false, "error": {"code": "invalid_argument", "message": "internalcmd: --render cannot be combined with --mode=package"}}
```

Classified via the existing `internalcmd.ErrInvalidArgument` sentinel (already mapped by `classify` to `invalid_argument`, exit code 2) — no new error code needed.

## 7. Compatibility statement (spec.md FR-008, SC-003)

- A caller that never passes `--mode` receives an envelope with two new fields (`schema_version`, `diagnostics.payload_tokens`) and otherwise byte-for-byte identical output to before this feature shipped, for both the no-flag default and `--render`.
- No currently-shipped Skill (`mister-plan`, `mister-tasks`, `mister-analyze`, `mister-implement`, `mister-wrap-up`) passes `--render` or reads `content`/`location`/`fingerprint`/`schema_version` today, so none require any change to keep working (research.md Decision 2). Adopting `--mode=package` to eliminate their own documented re-read step is a follow-up Skill change, not required by this feature to ship safely.
