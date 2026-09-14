# Canonical Skills — placeholder

Canonical Skill content
(`docs/architecture-specification.md` §38-39: `create-knowledge-base`,
`create-constitution`, `create-program`, `create-feature`,
`create-specs`, `create-plan`, `create-tasks`, `implement`, `analyze`)
is a later feature's work (Phase 6), not this one.

This file exists only so `kit.SkillsFS` — used for real by
`misterspec init` (008-cli-cobra) — has at least one embeddable file;
Go's `go:embed` cannot compile against an otherwise-empty directory
(specs/008-cli-cobra/research.md). It is materialized into a
bootstrapped project's `.claude/skills/` directory like any other Skill
file until it is replaced by real ones.
