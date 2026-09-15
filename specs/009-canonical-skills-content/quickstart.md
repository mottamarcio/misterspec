# Quickstart: Canonical Skills Content

For the first time, this quickstart is genuinely end-user-facing — not
"a future caller," but what `misterspec init` plus a real Claude Code
session actually does once this feature ships.

## 1. Bootstrap a project — Skills install for real now

```bash
$ misterspec init --agent claude-code --dir ./my-project
{"ok":true,"bootstrap":{...,"agent":{"adapter_id":"claude-code","integration_path":".claude/skills","outcomes":[
  {"name":"create-knowledge-base/SKILL.md","kind":"skill","status":"installed",...},
  {"name":"create-constitution/SKILL.md","kind":"skill","status":"installed",...},
  {"name":"create-program/SKILL.md","kind":"skill","status":"installed",...},
  {"name":"create-feature/SKILL.md","kind":"skill","status":"installed",...},
  {"name":"create-specs/SKILL.md","kind":"skill","status":"installed",...},
  {"name":"create-plan/SKILL.md","kind":"skill","status":"installed",...},
  {"name":"create-tasks/SKILL.md","kind":"skill","status":"installed",...},
  {"name":"implement/SKILL.md","kind":"skill","status":"installed",...},
  {"name":"analyze/SKILL.md","kind":"skill","status":"installed",...}
]}}}
```

Nine entries, one per canonical Skill — no more single-`README.md`
placeholder (research.md).

## 2. Claude Code discovers them natively

```text
$ ls my-project/.claude/skills/
analyze/  create-constitution/  create-feature/  create-knowledge-base/
create-plan/  create-program/  create-specs/  create-tasks/  implement/

$ cat my-project/.claude/skills/create-plan/SKILL.md
---
name: create-plan
description: Design a technical implementation plan for a Spec.
---

## Purpose
...
```

## 3. The dogfooding flow (§67) — the actual product, end to end

```text
place a PRD in ai/raw/
       ↓
/create-knowledge-base   → KNOW-001, KNOW-002, ...
       ↓
/create-constitution     → ai/memory/constitution.md
       ↓
/create-program          → PRG-001
       ↓
/create-feature PRG-001  → FEAT-001
       ↓
/create-specs FEAT-001   → SPEC-001
       ↓
/create-plan SPEC-001    → plan.md
       ↓
/create-tasks SPEC-001   → tasks.md (## TASK-001, ## TASK-002, ...)
       ↓
/implement SPEC-001      → real code change, one task at a time
       ↓
/analyze SPEC-001        → validation.md, pass or a specific named gap
```

Every arrow above is a Skill invocation; every Skill invocation's own
deterministic work is one or more `misterspec internal <op>` calls the
agent makes on the user's behalf — never a hand-guessed ID or path
(data-model.md's per-Skill operations table).

## 4. A deliberately incomplete implementation

```text
$ # /analyze SPEC-001, with SPEC-001:R3 not yet implemented

Outcome: FAIL

The implementation does not satisfy SPEC-001:R3.
The Spec remains valid; the implementation is incomplete.

Recommended next step:
/implement SPEC-001
```

(§53's branching rule: a failed `/analyze` recommends the *responsible*
artifact layer, not a fixed "run the next Skill in the pipeline.")

## Validation

Validated by: 005/006/008's full existing suites re-run unmodified
(the recursive-install regression gate); a new structural test parsing
all nine installed `SKILL.md` files against data-model.md's contract
table; and an end-to-end installation test confirming all nine land at
`.claude/skills/<name>/SKILL.md` via the real `claude.Install` adapter
— the same regression discipline every prior feature has used.
