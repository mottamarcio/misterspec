# MisterSpec — Implementation-Ready Architecture Specification

**Status:** Architecture Freeze Candidate
**Target Release:** MVP / v0.1
**Language:** Go
**CLI Framework:** Cobra
**TUI Framework:** Bubble Tea
**Primary User Command:** `misterspec init`
**Primary Runtime UX:** Agent slash commands
**Project Model:** Local-first, filesystem-authoritative, Git-friendly

---

# 1. Purpose of This Specification

This document converts the MisterSpec Product Requirements Document into an implementation-ready technical specification.

It freezes:

* the product lifecycle;
* the repository filesystem layout;
* the Go package architecture;
* initialization behavior;
* Bubble Tea interaction flow;
* agent adapter responsibilities;
* artifact schemas;
* ID rules;
* deterministic internal operations;
* machine-readable operation contracts;
* Skill structure;
* context contracts;
* mutation boundaries;
* completion summaries;
* recommended next-step behavior;
* structural validation requirements;
* MVP implementation order.

This document does not replace individual `SKILL.md` files.

Instead, it defines the contract those Skills must follow.

---

# 2. Architecture Summary

MisterSpec consists of three operational layers:

```text
┌─────────────────────────────────────────────────────────────┐
│                     Coding Agent                            │
│                                                             │
│ Claude Code / Codex / Gemini / Antigravity / ...           │
│                                                             │
│ /mister-knowledge-base                                      │
│ /mister-constitution                                        │
│ /mister-program                                             │
│ /mister-features                                             │
│ /mister-specify                                               │
│ /mister-plan                                                │
│ /mister-tasks                                               │
│ /mister-implement                                                  │
│ /mister-analyze                                                    │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            │ semantic reasoning
                            │ + deterministic operations
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  MisterSpec Framework                        │
│                                                             │
│ Canonical SKILL.md files                                    │
│ Templates                                                   │
│ Agent adapters                                              │
│ Internal deterministic operation API                        │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                       Repository                            │
│                                                             │
│ ai/             semantic project state                      │
│ .misterspec/     framework installation                      │
│ source code      implementation reality                     │
│ Git              historical record                          │
└─────────────────────────────────────────────────────────────┘
```

The user normally interacts with the Go executable exactly once:

```bash
misterspec init
```

After initialization:

```text
User
  ↓
Coding Agent
  ↓
misterspec Skills
  ↓
Repository
```

The user does not manually invoke deterministic misterspec operations.

---

# 3. Fundamental Responsibility Boundary

misterspec distinguishes between two forms of work.

## 3.1 Semantic Operations

Semantic operations require understanding meaning.

They belong to the coding agent following a misterspec Skill.

Examples:

```text
determine whether two requirements conflict

decide whether a Feature is cohesive

infer which Knowledge artifacts are relevant

design an implementation strategy

decide whether a requirement is ambiguous

determine whether code satisfies a Spec

decide whether an observation deserves to become a Learning
```

These cannot be reliably replaced by deterministic Go code.

---

## 3.2 Deterministic Operations

Deterministic operations produce the same correct result without needing to understand project meaning.

They belong to the misterspec executable.

Examples:

```text
allocate the next SPEC ID

resolve SPEC-014 to its canonical path

calculate SHA-256 for a source document

create a canonical Feature directory

read YAML frontmatter

validate ID syntax

check parent existence

find child entities

detect duplicate IDs

detect broken references

inspect artifact metadata

validate allowed lifecycle states
```

The coding agent must call these operations instead of reproducing their logic through LLM reasoning.

---

# 4. Public vs Internal Command Surface

The distinction between public and internal commands is mandatory.

## 4.1 Public User Command

The supported public interface for MVP is:

```bash
misterspec init
```

Running:

```bash
misterspec --help
```

should primarily expose:

```text
Usage:
  misterspec init
```

The product must not present users with a large CLI command vocabulary.

---

# 5. Internal Agent Command Namespace

Deterministic operations shall exist under:

```bash
misterspec internal ...
```

The `internal` command tree shall be hidden from normal Cobra help.

For example:

```bash
misterspec internal resolve SPEC-014
```

is valid.

However:

```bash
misterspec --help
```

must not advertise it.

These commands are:

* machine-facing;
* stable within a misterspec major version;
* documented inside canonical Skills;
* callable by supported coding agents;
* unsuitable as normal user UX.

The word `internal` means:

> Not part of the normal human-facing interface.

It does **not** mean:

> Unstable implementation detail.

Skills depend on these contracts, therefore compatibility matters.

---

# 6. Internal Operation Output Contract

Internal commands shall default to machine-readable JSON.

They must not print decorative output.

Example:

```bash
misterspec internal resolve SPEC-014
```

Output:

```json
{
  "ok": true,
  "entity": {
    "id": "SPEC-014",
    "type": "spec",
    "path": "ai/programs/PRG-001/features/FEAT-003/specs/SPEC-014/spec.md"
  }
}
```

Human-oriented ANSI styling must never appear in internal command output.

---

# 7. Internal Error Contract

Failures must produce:

* non-zero exit code;
* structured JSON on stdout or stderr according to implementation convention;
* stable error code;
* concise explanation.

Example:

```json
{
  "ok": false,
  "error": {
    "code": "entity_not_found",
    "message": "No entity with ID SPEC-014 exists."
  }
}
```

Possible stable error codes include:

```text
project_not_initialized
entity_not_found
entity_ambiguous
invalid_id
duplicate_id
invalid_parent
invalid_metadata
invalid_reference
invalid_target
path_outside_project
already_exists
unsupported_type
validation_failed
invalid_argument
```

Skills should reason about error codes rather than parse prose.

---

# 8. Exit Code Convention

Recommended MVP convention:

```text
0   success

1   unexpected failure

2   invalid invocation

3   target not found

4   structural validation failure

5   mutation rejected

6   misterspec project not initialized
```

Stable JSON error codes remain more important than exact numeric exit codes.

---

# 9. Internal Deterministic Operations

The MVP shall implement the following operations.

---

## 9.1 `project`

Returns information about the current misterspec project.

```bash
misterspec internal project
```

Example:

```json
{
  "ok": true,
  "project": {
    "root": "/repo/acme",
    "misterspec_dir": ".misterspec",
    "artifacts_dir": "ai",
    "agent": "claude-code",
    "schema_version": 1
  }
}
```

Responsibility:

* locate repository root;
* verify initialization;
* load configuration;
* normalize project-relative paths.

It performs no semantic reasoning.

---

## 9.2 `resolve`

Resolve an entity ID to its canonical artifact.

```bash
misterspec internal resolve SPEC-014
```

Supported IDs:

```text
PRG-*
FEAT-*
SPEC-*
TASK-*
KNOW-*
LRN-*
```

Example:

```json
{
  "ok": true,
  "entity": {
    "id": "SPEC-014",
    "type": "spec",
    "path": "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
    "directory": "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014"
  }
}
```

The agent must not search manually for globally identified artifacts when `resolve` can perform the operation.

---

## 9.3 `inspect`

Return deterministic metadata for an artifact.

```bash
misterspec internal inspect SPEC-014
```

Example:

```json
{
  "ok": true,
  "entity": {
    "id": "SPEC-014",
    "type": "spec",
    "status": "ready",
    "parent": "FEAT-004",
    "depends_on": [
      "SPEC-011"
    ],
    "supersedes": []
  }
}
```

`inspect` parses metadata.

It does not interpret body semantics.

---

## 9.4 `parent`

Return the structural parent.

```bash
misterspec internal parent SPEC-014
```

Example:

```json
{
  "ok": true,
  "parent": {
    "id": "FEAT-004",
    "type": "feature"
  }
}
```

---

## 9.5 `children`

Return structural children.

```bash
misterspec internal children PRG-001
```

Optional filter:

```bash
misterspec internal children PRG-001 --type feature
```

Example:

```json
{
  "ok": true,
  "children": [
    {
      "id": "FEAT-001",
      "type": "feature"
    },
    {
      "id": "FEAT-002",
      "type": "feature"
    }
  ]
}
```

This operation discovers structural children.

It does not determine semantic coverage.

---

# 10. Atomic Entity Allocation

A critical design requirement is that ID allocation and initial artifact creation should not be split into two unsafe steps when avoidable.

Instead of:

```text
next-id
↓
scaffold
```

the preferred operation is:

```text
create
```

---

## 10.1 `create`

Example:

```bash
misterspec internal create feature \
  --parent PRG-001 \
  --slug session-management
```

The operation shall:

1. lock or otherwise protect local ID allocation for the duration of the operation;
2. discover existing IDs;
3. allocate the next ID;
4. determine the canonical path;
5. create the canonical directory;
6. create the initial artifact from the embedded template;
7. write required frontmatter;
8. return the resulting ID and paths.

Example:

```json
{
  "ok": true,
  "created": {
    "id": "FEAT-007",
    "type": "feature",
    "path": "ai/programs/PRG-001/features/FEAT-007/feature.md"
  }
}
```

This avoids asking an LLM to implement ID allocation.

---

# 11. Supported `create` Types

MVP:

```text
knowledge
program
feature
spec
learning
```

Plan, Tasks, and Validation are subordinate artifacts inside a Spec and therefore use a separate artifact operation.

Example:

```bash
misterspec internal create-artifact plan --for SPEC-014
```

---

# 12. `create-artifact`

Creates canonical subordinate artifacts.

Examples:

```bash
misterspec internal create-artifact plan --for SPEC-014

misterspec internal create-artifact tasks --for SPEC-014

misterspec internal create-artifact validation --for SPEC-014
```

The command determines the target path.

The agent must not invent it.

---

# 13. `fingerprint`

Calculate source fingerprints.

```bash
misterspec internal fingerprint ai/raw/architecture.pdf
```

Output:

```json
{
  "ok": true,
  "source": {
    "path": "ai/raw/architecture.pdf",
    "algorithm": "sha256",
    "fingerprint": "sha256:..."
  }
}
```

The operation supports any file type.

It does not parse or understand the file.

---

# 14. `inventory`

Return deterministic file inventories.

Examples:

```bash
misterspec internal inventory raw
```

Output:

```json
{
  "ok": true,
  "files": [
    {
      "path": "ai/raw/prd.pdf",
      "extension": ".pdf",
      "size": 384229
    },
    {
      "path": "ai/raw/architecture.md",
      "extension": ".md",
      "size": 14029
    }
  ]
}
```

Other inventory scopes may include:

```text
knowledge
programs
specs
```

Inventory performs discovery only.

It does not decide relevance.

---

# 15. `references`

Return explicit artifact references.

```bash
misterspec internal references SPEC-014
```

Possible output:

```json
{
  "ok": true,
  "references": {
    "parent": [
      "FEAT-004"
    ],
    "dependencies": [
      "SPEC-011"
    ],
    "supersedes": [],
    "requirement_references": []
  }
}
```

It must only report references that can be deterministically extracted.

---

# 16. `validate`

Run structural validation.

Examples:

```bash
misterspec internal validate
```

or:

```bash
misterspec internal validate SPEC-014
```

Example success:

```json
{
  "ok": true,
  "valid": true,
  "findings": []
}
```

Example failure:

```json
{
  "ok": true,
  "valid": false,
  "findings": [
    {
      "code": "missing_parent",
      "severity": "error",
      "path": "...",
      "message": "FEAT-004 references missing parent PRG-003."
    }
  ]
}
```

Important:

```text
ok = command executed successfully

valid = project structure is valid
```

These concepts must remain distinct.

---

# 17. Structural Validation Scope

The deterministic validator shall detect at least:

* invalid YAML frontmatter;
* invalid entity type;
* malformed IDs;
* duplicate IDs;
* wrong canonical directory;
* missing parent;
* invalid parent type;
* invalid lifecycle state;
* missing required fields;
* invalid dependency IDs;
* missing dependencies;
* invalid supersedes references;
* broken artifact links where deterministically detectable;
* duplicate requirement markers inside one Spec;
* duplicate Task IDs within the same Spec's Tasks artifact (Task numbering is scoped per Spec — the same `TASK-NNN` number legitimately appears in more than one Spec, see 031-canonical-task-identity);
* Task references to nonexistent requirements;
* missing Plan where Spec state requires a Plan;
* missing Tasks where Spec state requires Tasks;
* missing Validation artifact where required;
* path traversal outside project root;
* malformed misterspec configuration.

The validator must never report:

```text
"Feature is semantically complete"
```

or:

```text
"Requirement is correctly implemented"
```

Those are semantic judgments.

---

# 18. `status`

A machine-facing structural status operation may exist:

```bash
misterspec internal status
```

Output may include:

```json
{
  "ok": true,
  "counts": {
    "programs": 2,
    "features": 8,
    "specs": 21
  },
  "specs": {
    "draft": 4,
    "ready": 6,
    "in_progress": 2,
    "validated": 9
  },
  "structural_errors": 0
}
```

This command is for Skills.

It is not intended as user-facing CLI UX.

---

# 19. Operations That Must NOT Be Deterministic Commands

misterspec must resist adding machine commands for semantic decisions.

For example, there must not be:

```bash
misterspec internal find-relevant-knowledge SPEC-001

misterspec internal decide-feature-boundaries PRG-001

misterspec internal validate-requirement-meaning SPEC-001

misterspec internal recommend-architecture SPEC-001
```

These operations require semantic reasoning and belong to Skills.

---

# 20. Filesystem Layout

Canonical structure:

```text
/
├── ai/
│   ├── raw/
│   │
│   ├── knowledge/
│   │   └── KNOW-*.md
│   │
│   ├── memory/
│   │   ├── constitution.md
│   │   └── learnings/
│   │       └── LRN-*.md
│   │
│   └── programs/
│       └── PRG-*/
│           ├── program.md
│           └── features/
│               └── FEAT-*/
│                   ├── feature.md
│                   └── specs/
│                       └── SPEC-*/
│                           ├── spec.md
│                           ├── plan.md
│                           ├── tasks.md
│                           └── validation.md
│
├── .misterspec/
│   ├── config.yaml
│   ├── install.json
│   ├── templates/
│   └── skills/
│
└── <agent integration directories>
```

---

# 21. `.misterspec/config.yaml`

Initial schema:

```yaml
schema_version: 1

agent:
  id: claude-code

project:
  artifacts_dir: ai

knowledge:
  raw_dir: ai/raw
  knowledge_dir: ai/knowledge

memory:
  constitution: ai/memory/constitution.md
  learnings_dir: ai/memory/learnings

programs:
  root: ai/programs

ids:
  width: 3
```

The MVP should minimize configuration.

Values that never need customization should remain implementation defaults instead of configuration fields.

---

# 22. `.misterspec/install.json`

Example:

```json
{
  "schema_version": 1,
  "misterspec_version": "0.1.0",
  "agent": {
    "id": "claude-code",
    "integration_path": ".claude/skills"
  }
}
```

This file describes framework installation state.

It is not authoritative product state.

---

# 23. Knowledge Artifact Schema

File example:

```text
ai/knowledge/KNOW-001-authentication.md
```

Required frontmatter:

```yaml
---
id: KNOW-001
type: knowledge
status: active
sources:
  - path: ai/raw/security.pdf
    fingerprint: sha256:...
  - path: ai/raw/authentication.md
    fingerprint: sha256:...
---
```

Required body structure:

```markdown
# Authentication

## Summary

## Known Facts

## Constraints

## Unknowns

## Conflicts

## Provenance
```

Optional:

```markdown
## Related Topics
```

Knowledge topics are semantic and therefore chosen by the agent.

IDs, paths, and fingerprints are deterministic.

---

# 24. Constitution Schema

Path:

```text
ai/memory/constitution.md
```

Frontmatter:

```yaml
---
type: constitution
schema_version: 1
---
```

Recommended structure:

```markdown
# Project Constitution

## Product Invariants

## Architecture Invariants

## Security Invariants

## Data Invariants

## Integration Invariants

## Quality Requirements

## Compatibility Requirements
```

Only relevant sections should be kept.

Empty bureaucracy must be avoided.

The Constitution should remain concise enough to load frequently.

`Quality Requirements` is the one section every project's Constitution
always carries, regardless of what the project's own Knowledge base
does or doesn't document: a fixed baseline requiring SOLID, DRY, KISS,
and YAGNI, and test-first practice (TDD/BDD). `/mister-constitution`
writes this baseline unconditionally, not distilled from Knowledge like
every other section — see `kit/skills/mister-constitution/SKILL.md`'s
own Outputs for its exact text.

---

# 25. Program Schema

Path:

```text
ai/programs/PRG-001/program.md
```

Frontmatter:

```yaml
---
id: PRG-001
type: program
status: draft
---
```

Body:

```markdown
# <Program Name>

## Problem

## Users and Stakeholders

## Desired Outcome

## Scope

## Non-Goals

## Constraints

## Success Criteria

## Relevant Knowledge

## Open Questions
```

Program statuses:

```text
draft
active
done
cancelled
```

---

# 26. Feature Schema

Path:

```text
.../features/FEAT-001/feature.md
```

Frontmatter:

```yaml
---
id: FEAT-001
type: feature
status: draft
parent: PRG-001
---
```

Body:

```markdown
# <Feature Name>

## Capability

## User Value

## Scope

## Non-Goals

## Constraints

## Relevant Knowledge

## Open Questions
```

Statuses:

```text
draft
active
done
cancelled
```

---

# 27. Spec Schema

Path:

```text
.../specs/SPEC-001/spec.md
```

Frontmatter:

```yaml
---
id: SPEC-001
type: spec
status: draft
parent: FEAT-001
depends_on: []
supersedes: []
---
```

Body:

```markdown
# <Spec Name>

## Intent

## Requirements

### R1 — <Requirement>

### R2 — <Requirement>

## Acceptance Scenarios

## Edge Cases

## Constraints

## Non-Goals

## Unresolved Questions

## Sources
```

Statuses:

```text
draft
ready
in_progress
validated
blocked
superseded
cancelled
```

---

# 28. Plan Schema

Path:

```text
.../SPEC-001/plan.md
```

Frontmatter:

```yaml
---
type: plan
for: SPEC-001
status: draft
---
```

Body:

```markdown
# Implementation Plan

## Summary

## Repository Context

## Requirement Coverage

## Architecture

## Components Affected

## Data Changes

## API Changes

## Integration Changes

## Implementation Sequence

## Test Strategy

## Risks

## Assumptions
```

Requirement coverage should explicitly map:

```text
R1 → planned implementation

R2 → planned implementation
```

---

# 29. Tasks Schema

Path:

```text
.../SPEC-001/tasks.md
```

Frontmatter:

```yaml
---
type: tasks
for: SPEC-001
---
```

Recommended task form:

```markdown
# Tasks

## TASK-001 — Add session persistence

- [ ] Complete
- Requirements: SPEC-001:R1, SPEC-001:R3
- Depends on: none
- Scope:
  - internal/session/store.go
- Verification:
  - go test ./internal/session/...

### Evidence

Pending.
```

Task status should primarily use Markdown checkboxes rather than duplicate lifecycle metadata.

---

# 30. Validation Schema

Path:

```text
.../SPEC-001/validation.md
```

Frontmatter:

```yaml
---
type: validation
for: SPEC-001
result: pending
---
```

Possible results:

```text
pending
passed
failed
```

Body:

```markdown
# Validation

## Summary

## Requirement Validation

### R1

- Plan coverage:
- Task coverage:
- Code evidence:
- Test evidence:
- Result:

## Unplanned Implementation

## Findings

## Recommended Corrections
```

---

# 31. Learning Schema

Path:

```text
ai/memory/learnings/LRN-001-<slug>.md
```

Frontmatter:

```yaml
---
id: LRN-001
type: learning
status: candidate
---
```

Statuses:

```text
candidate
promoted
dismissed
```

Body:

```markdown
# <Learning>

## Observation

## Evidence

## Why It May Be Reusable

## Potential Destination
```

A Skill must not promote a Learning automatically.

---

# 32. Go Repository Architecture

Recommended initial structure:

```text
misterspec/
├── cmd/
│   └── misterspec/
│       └── main.go
│
├── internal/
│   ├── cli/
│   │   ├── root.go
│   │   ├── init.go
│   │   └── internal.go
│   │
│   ├── tui/
│   │   ├── model.go
│   │   ├── update.go
│   │   ├── view.go
│   │   ├── styles.go
│   │   └── screens/
│   │
│   ├── project/
│   │   ├── root.go
│   │   ├── config.go
│   │   └── install.go
│   │
│   ├── artifacts/
│   │   ├── types.go
│   │   ├── metadata.go
│   │   ├── parser.go
│   │   └── paths.go
│   │
│   ├── ids/
│   │   ├── ids.go
│   │   └── allocator.go
│   │
│   ├── validation/
│   │   ├── validator.go
│   │   ├── findings.go
│   │   └── rules/
│   │
│   ├── operations/
│   │   ├── resolve.go
│   │   ├── inspect.go
│   │   ├── create.go
│   │   ├── fingerprint.go
│   │   ├── inventory.go
│   │   ├── references.go
│   │   └── status.go
│   │
│   ├── installer/
│   │   ├── installer.go
│   │   └── filesystem.go
│   │
│   └── agents/
│       ├── registry.go
│       ├── adapter.go
│       ├── claude/
│       ├── codex/
│       └── ...
│
├── kit/
│   ├── templates/
│   ├── skills/
│   └── integrations/
│
├── go.mod
└── go.sum
```

The package structure may evolve based on demonstrated complexity.

Do not introduce speculative layers such as:

```text
domain/
application/
repositories/
services/
controllers/
```

unless actual implementation pressure requires them.

---

# 33. Embedded Kit

Go `embed` should package:

```text
kit/templates/**
kit/skills/**
kit/integrations/**
```

Conceptually:

```go
//go:embed kit/**
var kitFS embed.FS
```

This makes the binary self-contained.

`misterspec init` must not require a network connection merely to retrieve standard Skills or templates.

---

# 34. Agent Adapter Interface

Conceptual Go interface:

```go
type Adapter interface {
    ID() string
    Name() string
    TargetPath() string
    Install(ctx context.Context, req InstallRequest) error
}
```

Example IDs:

```text
claude-code
codex
gemini-cli
opencode
antigravity
```

Exact adapters supported in v0.1 remain a release decision.

---

# 35. Adapter Responsibilities

An adapter may:

* identify the agent integration directory;
* transform canonical Skills where necessary;
* copy Skill files;
* create command definitions;
* create agent-specific metadata.

An adapter may not:

* change artifact semantics;
* modify project requirements;
* implement its own misterspec lifecycle;
* maintain authoritative project state.

Canonical behavior always originates from:

```text
.misterspec/skills/
```

---

# 36. Initialization Transaction Model

`misterspec init` should construct an installation plan before modifying files.

Conceptually:

```text
Inspect
  ↓
Collect user decisions
  ↓
Build installation plan
  ↓
Display preview
  ↓
User confirms
  ↓
Apply plan
  ↓
Verify
```

Where practical, failed initialization should avoid leaving a partially valid installation.

At minimum, partial failure must be detectable and safely rerunnable.

---

# 37. Bubble Tea State Model

Suggested states:

```go
type Screen int

const (
    ScreenInspect Screen = iota
    ScreenNonEmptyWarning
    ScreenAgentSelection
    ScreenPreview
    ScreenInstalling
    ScreenSuccess
    ScreenError
)
```

The Bubble Tea model should contain UI state.

It should not contain filesystem business rules.

---

# 38. Canonical Skill Set

MVP:

```text
mister-knowledge-base
mister-constitution
mister-program
mister-features
mister-specify
mister-plan
mister-tasks
mister-implement
mister-analyze
```

Optional, downstream documentation (029-spec-wrap-up-docs) — not part
of the required MVP pipeline above, usable at any point once a Spec
has a Plan:

```text
mister-wrap-up
```

Canonical source:

```text
.misterspec/skills/
```

---

# 39. Required `SKILL.md` Structure

Every Skill must contain:

```markdown
# Name

## Purpose

## Invocation

## Responsibility

## Inputs

## Outputs

## Preconditions

## Required Context

## Optional Context

## Conditional Context

## Unnecessary Context

## Authority

## Allowed Reads

## Allowed Creates

## Allowed Modifications

## Forbidden Mutations

## Deterministic Operations

## Procedure

## Decision Rules

## Interaction Rules

## Validation Rules

## Failure Conditions

## Stop Conditions

## Success Criteria

## Postconditions

## Idempotency

## Resume Behavior

## Completion Contract

## Recommended Next Step

## Related Skills
```

---

# 40. Deterministic Operations Section in Skills

Every Skill must explicitly state which internal commands it is allowed or expected to call.

Example:

```markdown
## Deterministic Operations

Use misterspec operations for mechanical repository work.

Required operations may include:

- `misterspec internal inventory raw`
- `misterspec internal fingerprint <path>`
- `misterspec internal create knowledge --slug <slug>`
- `misterspec internal validate`

Do not manually generate entity IDs.

Do not manually infer canonical artifact paths.
```

This is an important architecture rule.

The Skill should describe *when* to call the operation.

The binary defines *how* it is performed.

---

# 41. `/mister-knowledge-base` Operation Contract

Expected deterministic operations:

```text
internal project
internal inventory raw
internal fingerprint <source>
internal create knowledge
internal inspect <knowledge-id>
internal validate
```

Semantic responsibilities remain with the agent:

* understand Raw documents;
* extract facts;
* identify conflicts;
* identify unknowns;
* determine topic boundaries;
* merge overlapping knowledge;
* determine whether existing Knowledge should be updated.

---

# 42. `/mister-constitution` Operation Contract

Likely deterministic operations:

```text
internal inventory knowledge
internal resolve KNOW-###
internal validate
```

Semantic responsibilities:

* determine durable invariants;
* distinguish facts from rules;
* identify non-negotiable constraints;
* keep Constitution concise;
* surface contradictions.

The Constitution itself does not need a generated entity ID.

---

# 43. `/mister-program` Operation Contract

Deterministic operations:

```text
internal project
internal status
internal create program
internal validate PRG-###
```

Semantic responsibilities:

* define problem;
* define outcome;
* define scope;
* identify relevant Knowledge;
* determine success criteria;
* detect semantic Program overlap.

---

# 44. `/mister-features` Operation Contract

Deterministic operations:

```text
internal resolve PRG-###
internal children PRG-### --type feature
internal create feature --parent PRG-###
internal validate FEAT-###
```

Semantic responsibilities:

* identify capability boundaries;
* determine cohesion;
* detect semantic overlap;
* determine Program coverage.

---

# 45. `/mister-specify` Operation Contract

Deterministic operations:

```text
internal resolve FEAT-###
internal children FEAT-### --type spec
internal create spec --parent FEAT-###
internal validate SPEC-###
```

Semantic responsibilities:

* identify required behaviors;
* choose Spec boundaries;
* write requirements;
* define acceptance scenarios;
* identify dependencies;
* identify ambiguity;
* determine Feature coverage.

---

# 46. `/mister-plan` Operation Contract

Deterministic operations:

```text
internal resolve SPEC-###
internal inspect SPEC-###
internal references SPEC-###
internal create-artifact plan --for SPEC-###
internal validate SPEC-###
```

Semantic responsibilities:

* inspect relevant code;
* design architecture;
* map requirements;
* choose technical strategy;
* identify risks;
* define testing strategy.

---

# 47. `/mister-tasks` Operation Contract

Deterministic operations:

```text
internal resolve SPEC-###
internal inspect SPEC-###
internal create-artifact tasks --for SPEC-###
internal validate SPEC-###
```

An additional internal operation may allocate Task IDs:

```bash
misterspec internal allocate task --count 5
```

However, an even safer implementation may allow:

```bash
misterspec internal task-id --for SPEC-001 --count 5
```

The final API should avoid agents generating Task identifiers themselves.

Semantic responsibilities:

* decompose work;
* determine ordering;
* map requirements;
* define verification;
* identify dependencies.

---

# 48. `/mister-implement` Operation Contract

Deterministic operations:

```text
internal resolve SPEC-###
internal inspect SPEC-###
internal references SPEC-###
internal validate SPEC-###
```

The agent operates directly on repository source code using its normal coding capabilities.

misterspec does not mediate code edits.

Semantic responsibilities:

* select executable Task;
* load relevant context;
* implement;
* test;
* determine whether evidence is sufficient;
* identify upstream invalidation.

---

# 49. `/mister-analyze` Operation Contract

Deterministic operations:

```text
internal resolve SPEC-###
internal inspect SPEC-###
internal references SPEC-###
internal create-artifact validation --for SPEC-###
internal validate SPEC-###
```

Semantic responsibilities:

* compare required behavior to actual behavior;
* evaluate evidence;
* detect partial implementation;
* detect unplanned behavior;
* identify responsible artifact layer;
* determine pass/fail semantically.

---

# 50. Completion Contract

Every Skill invocation must finish with a concise operational summary.

Required concepts:

```text
Outcome

Artifacts

Important findings

Attention / unresolved issues

Recommended next step

Exact slash command
```

The sections may be formatted according to agent capabilities, but the semantic contract is mandatory.

---

# 51. Example Successful Completion

```text
◆ Knowledge Base created

Sources analyzed       14
Knowledge artifacts     8
Unknowns identified     3
Conflicts identified    1

Created

KNOW-001  Authentication
KNOW-002  Customer domain
KNOW-003  Payment processing
KNOW-004  Order lifecycle
KNOW-005  API contracts
KNOW-006  Data model
KNOW-007  Infrastructure
KNOW-008  Security constraints

Attention

1 conflict remains unresolved and has been preserved in
the relevant Knowledge artifact.

Next step

Create the project's primary memory and invariants:

/mister-constitution
```

---

# 52. Canonical Next-Step Graph

```text
Populate ai/raw/
       ↓
/mister-knowledge-base
       ↓
/mister-constitution
       ↓
/mister-program
       ↓
/mister-features PRG-###
       ↓
/mister-specify FEAT-###
       ↓
/mister-plan SPEC-###
       ↓
/mister-tasks SPEC-###
       ↓
/mister-implement SPEC-###
       ↓
/mister-analyze SPEC-###
```

This is the normal flow, not a mandatory linear state machine.

Skills must inspect current state.

---

# 53. Next-Step Behavior for Branching Cases

After `/mister-features`:

Primary recommendation:

```text
/mister-specify FEAT-001
```

Possible secondary note:

```text
You can continue decomposing PRG-001 with /mister-features PRG-001
before specifying this Feature.
```

After failed `/mister-analyze`:

Do not recommend a mechanically fixed command.

Instead recommend the artifact layer responsible for the gap.

Example:

```text
Next step

The implementation does not satisfy SPEC-014:R3.
The Spec remains valid, but the implementation is incomplete.

Run:

/mister-implement SPEC-014
```

Or:

```text
The current Plan cannot satisfy R4 without changing the public API.

Review the implementation strategy:

/mister-plan SPEC-014
```

---

# 54. Context Loading Rules

Every Skill starts with the smallest sufficient context.

Canonical conceptual tiers:

```text
Tier 0 — Constitution

Tier 1 — target artifact

Tier 2 — structural parents, children, dependencies

Tier 3 — semantically relevant Knowledge

Tier 4 — relevant implementation evidence

Tier 5 — wider repository
```

Important distinction:

The Go binary may discover structural relationships.

It must not choose semantically relevant Knowledge.

The agent performs that decision.

---

# 55. Mutation Rules

A Skill must not casually modify upstream artifacts.

Examples:

`/mister-plan`:

```text
may modify plan.md

must not rewrite spec.md
```

`/mister-tasks`:

```text
may modify tasks.md

must not rewrite plan.md unless the user explicitly requests refinement
```

`/mister-implement`:

```text
may modify code
may modify tests
may update task completion/evidence

must not rewrite Spec requirements
```

`/mister-analyze`:

```text
may modify validation.md
may create candidate Learnings

must not change implementation to make validation pass
must not rewrite the Spec to match the code
```

---

# 56. Path Safety

Every deterministic mutating operation must guarantee that resolved targets remain inside the repository root.

Inputs such as:

```text
../../etc/passwd
```

must be rejected.

Symlink traversal should be handled defensively where applicable.

---

# 57. Atomic Writes

Framework-managed artifact creation should use safe writes.

Preferred sequence:

```text
create temporary file
↓
write complete content
↓
fsync where appropriate
↓
atomic rename
```

User-owned source code remains edited by the coding agent, not by the misterspec operation layer.

---

# 58. Concurrency

MVP primarily targets one active repository workflow.

However, deterministic ID allocation should protect against accidental concurrent local allocation.

A short-lived project lock may be used only during mutating deterministic operations.

Example:

```text
.misterspec/.lock
```

The lock is operational and non-authoritative.

It must not contain semantic project state.

Stale locks must be safely recoverable.

---

# 59. No Persistent ID Counter

Do not store:

```json
{
  "next_spec_id": 15
}
```

as authoritative state.

Next IDs shall be derived from existing artifacts.

This preserves:

```text
filesystem = authoritative state
```

A temporary allocation lock is permitted.

A persistent authoritative counter is not.

---

# 60. ID Allocation Algorithm

For an entity type:

1. scan authoritative artifacts of that type;
2. parse valid IDs;
3. determine maximum numeric suffix;
4. allocate `max + 1`;
5. create artifact atomically while allocation lock is held.

Example:

```text
SPEC-001
SPEC-002
SPEC-004
```

Next ID:

```text
SPEC-005
```

Do not fill historical gaps by default.

---

# 61. Slug Rules

Directory identity must depend on ID, not mutable names.

Preferred:

```text
FEAT-004/
SPEC-014/
```

rather than:

```text
FEAT-004-session-management/
```

This avoids renaming directories when titles evolve.

Knowledge artifacts may include a slug because they are flat:

```text
KNOW-004-session-management.md
```

The ID remains authoritative.

---

# 62. PDF Handling

misterspec deterministic tooling does not perform semantic PDF extraction.

The selected coding agent is responsible for understanding supported Raw source documents.

The misterspec binary may deterministically:

* list PDFs;
* calculate hashes;
* inspect metadata such as path and size.

If an agent cannot read a PDF, `/mister-knowledge-base` must surface the source as unreadable rather than pretending it was processed.

---

# 63. Knowledge Source Fingerprints

Each processed Raw file should have a SHA-256 fingerprint stored in the Knowledge artifacts that cite it.

Example:

```yaml
sources:
  - path: ai/raw/architecture.pdf
    fingerprint: sha256:abc...
```

Future Skills can call:

```bash
misterspec internal fingerprint ai/raw/architecture.pdf
```

and compare the result to recorded fingerprints.

Semantic stale impact remains an agent decision.

---

# 64. Framework Policy vs Project Constitution

Agent operating rules belong to the misterspec framework and individual Skills.

Project invariants belong to:

```text
ai/memory/constitution.md
```

Therefore the MVP does not require:

```text
ai/memory/agent-policy.md
```

Generic rules such as:

```text
do not silently change requirements
use deterministic operations for ID allocation
load minimum sufficient context
```

belong to canonical misterspec Skill policy.

Project-specific rules belong to the Constitution.

---

# 65. Testing Strategy

The Go implementation should have four major test layers.

## Unit Tests

Examples:

```text
ID parsing
ID allocation
path resolution
frontmatter parsing
fingerprinting
validation rules
```

## Filesystem Integration Tests

Use temporary repositories.

Verify:

```text
create Program
create Feature
resolve ID
detect duplicates
reject invalid parents
```

## Adapter Tests

Verify generated agent integration output against fixtures.

## TUI Model Tests

Test Bubble Tea state transitions without relying on visual snapshots alone.

---

# 66. Golden Tests

Generated resources are good candidates for golden-file tests.

Examples:

```text
Program template
Spec template
Claude Skill adapter
config.yaml
install.json
```

Golden changes should require deliberate review.

---

# 67. Dogfooding Bootstrap

Once the following exist:

```text
misterspec init
canonical Skills
Claude or other first adapter
artifact creation primitives
```

the misterspec repository itself should run:

```text
place PRD in ai/raw/
↓
/mister-knowledge-base
↓
/mister-constitution
↓
/mister-program
↓
...
```

Development of later misterspec functionality should happen through misterspec itself whenever practical.

---

# 68. Suggested MVP Implementation Sequence

## Phase 1 — Core Repository Model

Implement:

```text
project root detection
config
artifact types
metadata parsing
path rules
ID parser
ID scanner
```

## Phase 2 — Deterministic Operations

Implement:

```text
resolve
inspect
parent
children
inventory
fingerprint
create
create-artifact
validate
status
```

## Phase 3 — Embedded Kit

Implement:

```text
templates
skills
go:embed
resource installer
```

## Phase 4 — First Agent Adapter

Prefer one reference adapter first.

Complete its integration end-to-end before generalizing aggressively.

## Phase 5 — `misterspec init`

Implement:

```text
Cobra command
repository inspection
Bubble Tea flow
agent selection
preview
installation
verification
success screen
```

## Phase 6 — Canonical Skills

Write and test:

```text
/mister-knowledge-base
/mister-constitution
/mister-program
/mister-features
/mister-specify
/mister-plan
/mister-tasks
/mister-implement
/mister-analyze
```

## Phase 7 — Second Brain

Turn the existing artifact model into a lightweight, local-first project
knowledge graph and retrieval layer — inspired by personal-knowledge-
management tools such as Obsidian's wikilink/backlink graph, but built
minimally into misterspec itself: no external app, no editor plugin, no
graphical UI, nothing the community is required to install.

Goal: find and assemble the smallest sufficient project context for the
current agent operation, reducing token usage, repeated repository
exploration, and agent startup cost per operation.

Implement:

```text
artifact wikilinks ([[SPEC-014]], [[SPEC-014|alias]])
formal + semantic reference graph (parent/depends_on/supersedes + wikilinks)
backlinks
section-aware Markdown chunking with provenance
disposable, rebuildable SQLite + FTS5 index (never authoritative)
incremental fingerprint-based synchronization
intent-aware, tiered, budgeted context retrieval
misterspec internal context
retrieval diagnostics (token-reduction metrics)
```

The filesystem remains the sole source of truth; the index is always
disposable and fully reconstructible from it. The coding agent retains
all semantic judgment — misterspec only resolves structure, ranks
retrieval candidates, and enforces a token budget. Existing Skills and
internal operations keep working unmodified until this phase's own
retrieval quality is dogfooded and proven; only then are Skills updated
to consume a Context Pack first.

Full implementation specification: `docs/context-engine-implementation.md`.

---

# 69. MVP Definition of Done

The MVP is done when a user can execute:

```bash
misterspec init
```

select a supported agent, then never need to directly use the misterspec executable again.

The developer can then:

```text
populate ai/raw/

run /mister-knowledge-base

run /mister-constitution

run /mister-program

run /mister-features

run /mister-specify

run /mister-plan

run /mister-tasks

run /mister-implement

run /mister-analyze
```

During this lifecycle:

* agents use internal deterministic operations when appropriate;
* users do not need to know those operations exist;
* IDs are never invented by the LLM;
* canonical paths are never guessed by the LLM;
* source hashes are never fabricated;
* structural validity is never inferred semantically;
* semantic decisions remain with the coding agent;
* every Skill reports what it changed;
* every Skill recommends the next action;
* project state remains reconstructable from repository files;
* misterspec remains independent of a proprietary coding-agent runtime.

---

# 70. Frozen Core Architecture

For MVP, the following decisions should be considered frozen:

```text
Name
    misterspec

Language
    Go

Executable
    misterspec

Public command
    misterspec init, misterspec --version, misterspec --update
    (--version/--update amended 030-cli-version-update; see
    .specify/memory/constitution.md's Sync Impact Report, v1.1.0)

Public terminal UX
    bootstrap only

CLI framework
    Cobra

TUI
    Bubble Tea

Primary product UX
    agent slash commands

Canonical lifecycle
    Raw
    → Knowledge
    → Constitution
    → Program
    → Feature
    → Spec
    → Plan
    → Tasks
    → Implementation
    → Validation

Artifact storage
    filesystem

Semantic format
    Markdown

Raw MVP formats
    Markdown + PDF

History
    Git

Agent runtime
    external coding agent

Multi-agent orchestration
    not MVP

Database
    none

Vector store
    none

Secondary authoritative state
    none

Canonical Skills
    .misterspec/skills/

Project artifacts
    ai/

Project memory
    ai/memory/constitution.md

Deterministic machine API
    misterspec internal ...

Human visibility of machine API
    hidden

Internal operation output
    structured JSON

Semantic reasoning
    coding agent

Mechanical operations
    Go binary

Completion summary
    mandatory after every Skill

Recommended next slash command
    mandatory after every Skill
```

---

# 71. Final Architecture Principle

The central implementation boundary is:

```text
If the answer depends on what project information means:
    the agent decides.

If the answer can be computed from repository structure:
    misterspec computes it.
```

This boundary should be visible throughout the codebase and throughout every `SKILL.md`.

The developer should experience misterspec as a disciplined workflow inside the coding agent.

The coding agent should experience misterspec as:

```text
explicit project artifacts
+
semantic procedures
+
reliable deterministic primitives
```

The deterministic primitives exist to remove mechanical uncertainty from LLM reasoning.

They are infrastructure for the Skills, not commands the user is expected to learn.
