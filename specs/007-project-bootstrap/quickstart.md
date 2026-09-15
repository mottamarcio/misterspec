# Quickstart: Project Bootstrap

Still no CLI — this shows how a *future* caller (eventually
`misterspec init`, after its own interactive agent-selection step) uses
`internal/bootstrap` on top of `internal/project`, `internal/installer`,
and `internal/agents`.

## 1. Inspect before doing anything

```go
result, err := bootstrap.Inspect(targetDir)
// A fresh, empty directory:
// result.Initialized == false

// A directory that's already a project, agent installed:
// result.Initialized    == true
// result.AgentInstalled == true
// result.InstalledAgent == "claude-code"

// A directory that's already a project, no agent installed yet:
// result.Initialized    == true
// result.AgentInstalled == false
// result.InstalledAgent == ""
```

## 2. Bootstrap an uninitialized directory for a chosen agent

```go
registry := builtin.Default()

outcome, err := bootstrap.Bootstrap(targetDir, "claude-code", registry, skillsFixtureFS)
// err == nil
// outcome.ProjectRoot      == targetDir
// outcome.ConfigWritten    == true
// outcome.TemplateOutcomes == one installer.Outcome per kit resource, all Installed
// outcome.AgentInstall.AdapterID       == "claude-code"
// outcome.AgentInstall.IntegrationPath == ".claude/skills"
```

## 3. Rejections happen before any write

```go
// Already-initialized target:
_, err = bootstrap.Bootstrap(targetDir, "claude-code", registry, skillsFixtureFS)
// errors.Is(err, bootstrap.ErrAlreadyInitialized) == true
// nothing in targetDir was touched by this second call

// Unregistered agent ID, against a fresh target:
_, err = bootstrap.Bootstrap(freshDir, "no-such-agent", registry, skillsFixtureFS)
// errors.Is(err, bootstrap.ErrUnknownAgent) == true
// freshDir remains empty — no config, no templates, no Skills written
```

## 4. Verify the result

```go
verify, err := bootstrap.Verify(targetDir, "claude-code")
// verify.Detected       == true
// verify.InstalledAgent == "claude-code"
// verify.AgentMatches   == true

// Verifying against a different expected agent:
verify, err = bootstrap.Verify(targetDir, "codex")
// verify.Detected       == true   (the project itself is still valid)
// verify.InstalledAgent == "claude-code"
// verify.AgentMatches   == false  (mismatch, not an error)
```

## Validation

Validated by this feature's filesystem-integration tests (a fresh
temporary directory, a fixture Skills `fs.FS`, a fixture `Registry`)
plus the full existing regression suite (`internal/project`,
`internal/installer`, `internal/agents`, `-race`) re-run unmodified —
the same regression discipline every prior feature has used.
