# Phase 1 Data Model: Project Documentation Site and README

No Go types. This feature's only "data model" is the site's own
content structure and the shared navigation script's own data shape.

## Navigation entity (`site/assets/nav.js`)

```js
// One flat array, in display order. `group` clusters related entries
// under one sidebar heading; entries with the same `group` render
// together. `href` is relative to the site root.
const NAV = [
  { group: "Overview",        title: "Introduction",              href: "index.html" },
  { group: "Overview",        title: "Getting Started",           href: "getting-started.html" },
  { group: "Overview",        title: "The Spec-Driven Workflow",  href: "workflow.html" },
  { group: "Features",        title: "Core Foundation",           href: "foundation.html" },
  { group: "Features",        title: "Agent Adapters & CLI",      href: "agents-cli.html" },
  { group: "Features",        title: "The Context Engine",        href: "context-engine.html" },
  { group: "Features",        title: "Multi-Agent Skill Integration", href: "multi-agent-skills.html" },
  { group: "Features",        title: "Dogfooding & Evaluation",   href: "dogfooding.html" },
  { group: "Reference",       title: "Command Reference",         href: "commands.html" },
];
```

`nav.js` renders this into `#sidebar` on `DOMContentLoaded`, marking
the current page (matched against `location.pathname`) with a
`slate-900` highlight — satisfying FR-010 ("orient themselves... from
any page") without duplicating markup per page (research.md #2).

## Feature Section content model (one per `site/*.html` feature page)

Every feature page follows the same fixed shape, so a future addition
follows an obvious template (Edge Cases: "obvious where a new entry
belongs"):

| Element | Content |
|---|---|
| What it is | One paragraph, plain language |
| Why it exists | The specific problem it solves, referencing the real prior gap |
| How it fits | Its real relationship to the feature(s) immediately before/after it (FR-007) |
| Example | At least one concrete, runnable example (a real command invocation and its real output, or a real before/after artifact snippet) |

## Feature-to-page inventory (FR-007, 100% of Specs 001-019)

| Spec | Real feature name | Page |
|---|---|---|
| 001 | Core Repository Foundation | `foundation.html` |
| 002 | Read-Only Deterministic Operations | `foundation.html` |
| 003 | Atomic Entity Creation | `foundation.html` |
| 004 | Structural Validation & Project Status | `foundation.html` |
| 005 | Embedded Kit & Resource Installer | `foundation.html` |
| 006 | Agent Adapter Layer | `agents-cli.html` |
| 007 | Project Bootstrap | `agents-cli.html` |
| 008 | CLI (Cobra) & the `internal` Command Tree | `agents-cli.html` |
| 009 | Canonical Skills Content | `agents-cli.html` |
| 010 | Interactive `init` TUI | `agents-cli.html` |
| 011 | Wikilink Graph Foundation | `context-engine.html` §1 |
| 012 | References and Backlinks | `context-engine.html` §2 |
| 013 | Document Model and Chunking | `context-engine.html` §3 |
| 014 | Disposable SQLite Index | `context-engine.html` §4 |
| 015 | Context Collector and Retrieval | `context-engine.html` §5 |
| 016 | Ranking and Budgeting | `context-engine.html` §6 |
| 017 | Internal Context Command | `context-engine.html` §7 |
| 018 | Multi-Agent Skill Integration | `multi-agent-skills.html` |
| 019 | Dogfooding and Evaluation | `dogfooding.html` |

## Command Reference Entry content model (`commands.html`, one per command)

| Element | Content |
|---|---|
| Signature | `misterspec [internal] <command> <args> [flags]`, exactly as registered |
| Purpose | One sentence, from the command's own `Short:` field or its governing spec |
| Flags | Every real flag, its default, and its meaning (read from the command's own source) |
| Example | One real invocation against a small example project, with its real (or representative-real-shape) JSON output |

## Command inventory (FR-008, 100% of the 14 currently-registered commands)

| Command | Kind | Real flags (beyond shared `--dir`) |
|---|---|---|
| `init` | Public | `--agent` |
| `internal resolve <id>` | Internal | — |
| `internal inspect <id>` | Internal | — |
| `internal parent <id>` | Internal | — |
| `internal children <id>` | Internal | `--type` |
| `internal create <type>` | Internal | `--parent`, `--slug` |
| `internal create-artifact <kind>` | Internal | `--for` |
| `internal fingerprint <path>` | Internal | — |
| `internal inventory <dir>` | Internal | — |
| `internal validate [<id>]` | Internal | — |
| `internal status` | Internal | — |
| `internal references <id>` | Internal | — |
| `internal backlinks <id>` | Internal | — |
| `internal context <id>` | Internal | `--intent`, `--task`, `--query`, `--budget`, `--render` |

All 14 verified directly against each command's own current
`internal/cli/{init.go,internalcmd/*.go}` source at planning time
(research.md #4) — not from memory.
