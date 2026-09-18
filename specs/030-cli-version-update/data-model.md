# Phase 1 Data Model: `misterspec --version` and `misterspec --update`

No new persisted entity is introduced — the binary's own version is build-time-embedded (not stored anywhere at runtime beyond the compiled binary itself), and `--update` reads (never writes) the project's existing `.misterspec/install.json`. Documented here for the concepts involved.

## Embedded Version (build-time constant, not a stored entity)

| Field | Description |
|---|---|
| Value | Set via `-ldflags "-X .../selfupdate.Version=$TAG"` at build time; empty string (`""`) for a local/development build |
| Reported as | The literal tag string (e.g. `v1.2.0`) when set; a clear "development build" message when empty (FR-002) |

## Latest Release (fetched live from GitHub, never cached/persisted)

| Field | Description |
|---|---|
| TagName | The latest published release's own tag (e.g. `v1.2.0`) |
| Assets | The list of binary files published with that release, each with its own filename and download URL |
| ChecksumsAsset | The `SHA256SUMS` file published alongside the binaries (research.md's own decision) |

## Selected Release Asset (derived, not stored)

| Field | Derivation |
|---|---|
| Name | `misterspec-<GOOS>-<GOARCH>[.exe]` — matching `release.yml`'s own existing naming convention exactly, for the current process's own `runtime.GOOS`/`runtime.GOARCH` |
| Checksum | The matching line for that filename, parsed out of the downloaded `SHA256SUMS` file |

## Update Decision (derived at runtime, not stored)

| State | Meaning |
|---|---|
| Up to date | The running `Version` is not older than the latest release's `TagName` (per `golang.org/x/mod/semver` comparison) — `--update` reports this and exits, no download, no prompt (FR-004) |
| Update available | The latest release is newer — `--update` reports both versions and requires confirmation (FR-005) |
| Check failed | The GitHub API call itself failed (network, outage, rate limit) — reported plainly, nothing changed (FR-012) |
| No compatible asset | The latest release has no asset matching the current OS/arch — reported plainly, nothing changed (FR-012) |

## Project Agent Integration Record (existing, read-only)

Not new — this is `.misterspec/install.json`, already written by `misterspec init` (`internal/bootstrap`). `--update` only reads its existing `agent_id` field to know which Skill-installation path to re-run (research.md's own decision) — this feature adds no new field to it and never writes to it.

## State Transitions (the binary file itself, at its own installed path)

```text
current binary at path P ──(update confirmed, checksum verified)──> new binary at path P
                                                                       (Unix: single atomic rename;
                                                                        Windows: rename-current-aside,
                                                                        then rename-new-into-place)
current binary at path P ──(checksum mismatch, decline, or any earlier failure)──> unchanged at path P
```

No intermediate state is ever externally observable — the path at `P` is a valid, executable binary before, during (via the temp file, never at `P` itself), and after every `--update` run, success or failure.
