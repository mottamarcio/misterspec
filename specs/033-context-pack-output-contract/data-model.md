# Phase 1 Data Model: Context Pack completo e contrato de saída versionado

No persisted storage is added (Constitution Principle III) — every type below is computed fresh, in memory, from the same `Collect`/`Rank`/`ApplyBudget` pipeline `internal context` already runs on each call. This extends `internal/context`'s existing types (`Candidate`, `ScoredCandidate`, `ResultItem`, `Result`, `Diagnostics`) rather than replacing them — see research.md Decision 1: their field *names* are unchanged, only `StartLine`/`EndLine`'s *values* become file-absolute.

## OutputMode

The caller's explicit choice of response shape (spec.md "Output Mode").

| Value | Meaning |
|---|---|
| `manifest` (default) | Today's existing shape: `items[]` carry metadata only (`path`, `heading`, `tier`, `reasons`, `score`, `tokens`) — no `content`, no `rendered`. Byte-for-byte identical to today's no-flag output (research.md Decision 2). |
| `package` | `items[]` carry every manifest field **plus** `content`, `location` (file-absolute), and `fingerprint`. No `rendered` field in the same response (FR-005). |
| `markdown` | Equivalent to today's `--render`: `items[]` stay metadata-only (as in `manifest`); `rendered` carries the single Markdown string. |

`--render` (unchanged flag) is equivalent to requesting `markdown` mode's `rendered` field *in addition to* the default `manifest`-shaped `items[]` — exactly today's existing combination, preserved verbatim.

## ItemLocation

Replaces the implicit, currently-wrong body-relative `StartLine`/`EndLine` pairing with an explicit, correctly-scoped concept (spec.md FR-003, "Context Pack Item").

| Field | Type | Description |
|---|---|---|
| `Path` | `string` | The source artifact's path, exactly as already carried (unchanged). |
| `StartLine` | `int` | 1-indexed, **file-absolute** (research.md Decision 1) — the first line of this item's content in the whole source file, frontmatter included. |
| `EndLine` | `int` | 1-indexed, file-absolute — the last line of this item's content. |

**Validation rule**: `StartLine <= EndLine`, both `>= 1`. For an artifact with no frontmatter, `StartLine`/`EndLine` are unchanged from today's values (the offset is `0`) — research.md Decision 1's Edge Case.

**Task-derived items** (spec.md FR-009): no separate shape — a Chunk from a Task's own `## TASK-NNN — Title` heading already has `StartLine`/`EndLine` bounded to that Task's own section (research.md Decision 7); `heading` (already present) names it (e.g. `"TASK-003 — Add session persistence"`).

## ItemFingerprint

A deterministic digest of one item's own selected content (research.md Decision 4).

| Field | Type | Description |
|---|---|---|
| (rendered as) | `string` | `"sha256:<hex>"` — SHA-256 of the item's own `Content` bytes, matching `operations.FileFingerprint`'s existing string format. |

**Validation rule**: Two items with byte-identical `Content` produce the identical fingerprint string; any difference in `Content` (even whitespace) produces a different one — a pure function of `Content` alone, computed fresh per response, never cached.

## PackageItem (package-mode item shape)

Extends today's manifest item (`path`, `heading`, `tier`, `reasons`, `score`, `tokens` — unchanged) with the new fields `package` mode adds.

| Field | Type | Description |
|---|---|---|
| `content` | `string` | The item's own selected text, verbatim — the same bytes `ResultItem.Content` already holds in memory (no re-read). |
| `location` | `ItemLocation` | File-absolute `start_line`/`end_line`, alongside the existing `path`. |
| `fingerprint` | `string` | This item's `ItemFingerprint`, per Decision 4. |

## ContextPackEnvelope

The versioned wrapper around every mode's response (spec.md "Context Pack Envelope").

| Field | Type | Description |
|---|---|---|
| `schema_version` | `int` | `1` for this feature's contract (research.md Decision 5). Present in every mode, including the unchanged default. |
| `target`, `intent`, `budget` | (unchanged) | Carried exactly as today. |
| `estimated_tokens`, `budget_exceeded`, `overage` | (unchanged) | Carried exactly as today — content-only token accounting, per Decision 6. |
| `items` | `[]ManifestItem` or `[]PackageItem` | Shape depends on `OutputMode`, per the table above. |
| `diagnostics` | `Diagnostics` | Existing fields unchanged, plus the new `payload_tokens` field (below). |
| `rendered` | `string` or `null` | Unchanged: the Markdown string when `markdown` mode (or `--render`) is requested, `null` otherwise. |

## Diagnostics (extended)

| Field | Type | Description |
|---|---|---|
| `candidates_considered`, `items_selected`, `tokens_available`, `tokens_selected`, `tokens_excluded`, `reduction_percent` | (unchanged) | Content-only accounting, exactly as today (research.md Decision 6) — an existing consumer reading these keeps getting identical numbers. |
| `payload_tokens` | `int` (new) | Estimated size of the actual response body for the requested `OutputMode` — item metadata plus content (package mode) or the rendered Markdown string (markdown mode) — always `>= tokens_selected` (spec.md FR-007). |

## State / Lifecycle

No new lifecycle or persisted state. `OutputMode` is a per-request choice, not a stored setting; `ItemFingerprint` and `payload_tokens` are recomputed from the same in-memory `Result` on every call.
