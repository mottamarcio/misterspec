# Contract: `init` Outcome and Release Assets

Reconciled against the actual implementation (021-init-scaffold-distribution).

## `misterspec init --agent <id>` JSON success shape

```json
{
  "ok": true,
  "bootstrap": {
    "project_root": "<dir>",
    "config_written": true,
    "directories": ["ai", "ai/raw", "ai/knowledge", "ai/memory", "ai/memory/learnings", "ai/programs"],
    "agent": {
      "adapter_id": "<id>",
      "integration_path": "<agent target path>",
      "outcomes": [{"name": "<skill>", "status": "installed|skipped|failed", "path": "..."}]
    }
  }
}
```

No `"templates"` key (data-model.md — removed, not emptied).

## Interactive `init` TUI (unchanged transition table, changed content)

- **Preview screen**: lists `configPath`, the six scaffolded
  directories (not "kit templates"), and the chosen agent's own Skill
  resources — unchanged shape otherwise.
- **Success screen**: reports "N directories scaffolded" (not "N kit
  resources installed") alongside the existing config/Skills lines.

## Release assets (per tagged version)

| Asset | Platform |
|---|---|
| `misterspec-linux-amd64` | Linux x86-64 |
| `misterspec-linux-arm64` | Linux ARM64 |
| `misterspec-darwin-amd64` | macOS Intel |
| `misterspec-darwin-arm64` | macOS Apple Silicon |
| `misterspec-windows-amd64.exe` | Windows x86-64 |

Built by `.github/workflows/release.yml`, triggered by pushing a `v*` tag or via
`workflow_dispatch` with an existing tag — one job per matrix entry builds with
`GOOS`/`GOARCH`/`CGO_ENABLED=0`, uploads as a build artifact, then a final job
downloads all of them and attaches them to the matching GitHub Release via
`softprops/action-gh-release`.

Reachable via `https://github.com/mottamarcio/misterspec/releases/latest` (research.md #7) — stable across future releases.

## Behavioral guarantees carried over (not re-tested here, only extended)

- `Bootstrap` still refuses an already-initialized target and an
  unregistered agent ID before writing anything (007's own FR-004/
  FR-007, unchanged).
- Every directory `scaffoldDirectories` creates is one of
  `project.Configuration`'s own already-validated default paths — no
  new path-safety surface (Constitution Principle VIII).
- A downloaded release binary's `--help`/`init`/`internal ...` behavior
  is identical to a source build at the same commit (FR-008) —
  verified by the release workflow building from the exact tagged
  commit with no build-flag differences beyond `GOOS`/`GOARCH`.
