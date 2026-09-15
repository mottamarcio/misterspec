# Quickstart: Multi-Agent Skill Integration

## 1. Discover the six now-supported agents

```go
registry := builtin.Default()
for _, a := range registry.List() {
    fmt.Println(a.ID(), a.Name(), a.TargetPath())
}
```

```text
agy           Antigravity           .agents/skills
claude-code   Claude Code           .claude/skills
codex         Codex CLI             .agents/skills
copilot       GitHub Copilot        .github/skills
cursor-agent  Cursor                .cursor/skills
devin         Devin for Terminal    .devin/skills
```

## 2. Install for a new agent (Cursor, as an example)

```go
adapter, _ := registry.Get("cursor-agent")
result, err := adapter.Install(ctx, agents.InstallRequest{
    ProjectRoot: root,
    Skills:      kit.SkillsFS,
    Overwrite:   false,
})
```

Every canonical Skill now exists at
`<root>/.cursor/skills/<skill-name>/SKILL.md`, byte-identical to what
`.claude/skills/<skill-name>/SKILL.md` already contains.

## 3. Install for two agents that share a target directory

```go
codexAdapter, _ := registry.Get("codex")
agyAdapter, _ := registry.Get("agy")
codexAdapter.Install(ctx, agents.InstallRequest{ProjectRoot: root, Skills: kit.SkillsFS})
agyAdapter.Install(ctx, agents.InstallRequest{ProjectRoot: root, Skills: kit.SkillsFS})
```

Both write to `<root>/.agents/skills/` — the second call's own
`Overwrite: false` default means already-installed files are `Skipped`,
not corrupted; `install.json` records both `codex` and `agy` as
installed, each independently.

## 4. `/implement` now requests a Context Pack first

Before this feature, `/implement SPEC-014`'s own `Procedure` began:

```text
1. Run `internal resolve SPEC-###` and `internal inspect SPEC-###`.
2. Read the Spec's Tasks artifact; select one Task...
```

After this feature:

```text
1. Run `internal resolve SPEC-###` and `internal inspect SPEC-###`.
2. Run `internal context SPEC-### --intent implementation` and begin
   from its returned items. If the request fails, proceed using this
   Skill's own Required Context below instead.
3. Read the Spec's Tasks artifact; select one Task...
```

If step 2's own `internal context` call returns `{"ok": false, ...}`
(for example, the target is momentarily unresolvable), the Skill does
not stop — it proceeds to step 3 exactly as it always did before this
feature.

## 5. The agent can still look further

`/implement`'s own `Required Context`/`Optional Context` sections now
also state, in one added sentence: the Context Pack is a starting
point, not a boundary — the agent remains free to read further
repository files or run further `internal` operations whenever it
determines the pack alone is insufficient (FR-007), exactly as
`docs/context-engine-implementation.md` §23 requires.
