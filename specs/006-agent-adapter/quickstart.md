# Quickstart: Agent Adapter Layer

Still no CLI — this shows how a *future* caller (eventually
`misterspec init`) uses `internal/agents` on top of the relocated/
generalized `internal/installer`.

## 1. Discover available adapters

```go
registry := builtin.Default()
for _, a := range registry.List() {
    fmt.Println(a.ID(), a.Name(), a.TargetPath())
}
// claude-code  Claude Code  .claude/skills
```

## 2. Select one by ID

```go
adapter, ok := registry.Get("claude-code")
// ok == true

_, ok = registry.Get("codex") // not built yet
// ok == false — distinct, not an error, not a default fallback
```

## 3. Install Skills for the selected adapter

```go
result, err := adapter.Install(ctx, agents.InstallRequest{
    ProjectRoot: proj.Root,
    Skills:      skillsFixtureFS, // kit.SkillsFS once Phase 6 exists
    Overwrite:   false,
})
// result.AdapterID       == "claude-code"
// result.IntegrationPath == ".claude/skills"
// result.Outcomes        == one installer.Outcome per Skill resource,
//                            every one Installed, byte-identical
```

## 4. Read back what's currently installed

```go
record, installed, err := agents.CurrentInstall(proj.Root)
// installed == true
// record.Agent.ID              == "claude-code"
// record.Agent.IntegrationPath == ".claude/skills"

// A project that was never installed:
_, installed, err = agents.CurrentInstall(freshProjectRoot)
// installed == false, err == nil — not an error
```

## Validation

Validated by this feature's filesystem-integration tests (a fixture
Skills `fs.FS`, a fresh target directory) plus 005-embedded-kit's own
test suite re-run unmodified after `internal/installer`'s generalization
— the same regression discipline every prior feature has used.
