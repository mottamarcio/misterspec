# Quickstart: Dogfooding and Evaluation

## 1. Build the fixture (once)

```sh
mkdir -p specs/019-dogfooding-evaluation/fixture
# .misterspec/config.yaml + ai/... tree per data-model.md, encoding
# 011-018's own real dependency graph (research.md #1).
```

## 2. Run User Story 1 against every fixture Spec

```sh
for id in SPEC-006 SPEC-011 SPEC-012 SPEC-013 SPEC-014 SPEC-015 SPEC-016 SPEC-017 SPEC-018; do
  misterspec internal context "$id" --intent planning \
    --dir specs/019-dogfooding-evaluation/fixture
done
```

For `SPEC-015` (real `depends_on`: 011, 012, 013, 014), the returned
`context.items` should include structural entries for all four — any
missing is an `omission` Finding, recorded in `report.md`.

## 3. Install both Skills into the fixture, once

```go
claudeAdapter, _ := builtin.Default().Get("claude-code")
agyAdapter, _ := builtin.Default().Get("agy")
claudeAdapter.Install(ctx, agents.InstallRequest{ProjectRoot: fixtureRoot, Skills: kit.SkillsFS})
agyAdapter.Install(ctx, agents.InstallRequest{ProjectRoot: fixtureRoot, Skills: kit.SkillsFS})
```

## 4. Live-dogfood through Claude Code

Open the fixture in Claude Code; run:

```text
/create-plan SPEC-018
```

Observe: does it call `internal context SPEC-018 --intent planning`
before reading anything else? Is the pack (which already contains
`SPEC-006`, `SPEC-017`, and `KNOW-002` per the real dependency graph)
sufficient on its own? How long did the whole invocation take?

## 5. Repeat through Antigravity

Same Spec, same question set, through Antigravity. Compare against
step 4's own observations.

## 6. Write the Tuning Decision

Review every Finding from steps 2, 4, and 5. If every one is
explained by something other than ranking/budgeting behavior (e.g. a
fixture wikilink that should have been added but wasn't), the outcome
is "no change warranted" — recorded plainly, not treated as a
disappointing result.
