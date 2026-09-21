// Package skillgen composes the canonical kit/skills/*/SKILL.md files
// from a small set of named, reusable Fragments plus a per-Skill
// SkillManifest (specs/039-lean-skills-integration-contracts/data-model.md
// "Fragment"/"SkillManifest"). It is a development-time authoring tool
// with no CLI wiring: generation happens via `go generate
// ./internal/skillgen/...` (see cmd/gen), never at product runtime, and
// its output is committed like any other source file — kit.SkillsFS
// (kit/kit.go) continues to embed only the resulting, already-composed
// SKILL.md content (plan.md "Scale/Scope").
package skillgen
