---
id: KNOW-002
type: knowledge
status: active
---

## Summary

Real research from 018-multi-agent-skill-integration: Antigravity,
Codex CLI, GitHub Copilot, Cursor, and Devin for Terminal all scan the
same open "Agent Skills" convention (a directory per skill containing
a `SKILL.md` with `name`/`description` frontmatter) misterspec's own
canonical Skills already use — verified against each agent's own
documentation and against GitHub Spec Kit's own integrations
reference. Only the target root directory differs per agent
(`.agents/skills` for Antigravity and Codex CLI, `.github/skills` for
GitHub Copilot, `.cursor/skills` for Cursor, `.devin/skills` for Devin
for Terminal, `.claude/skills` for Claude Code) — no content
transformation was needed for any of the five new adapters.
