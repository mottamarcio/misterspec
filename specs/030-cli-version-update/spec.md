# Feature Specification: `misterspec --version` and `misterspec --update`

**Feature Branch**: `030-cli-version-update`
**Created**: 2026-09-18
**Status**: Draft
**Input**: User description: "quero criar mais dois comandos CLI: um 'misterspec --version' para checar a versão do binario; e um 'misterspec --update' para, se possivel, verificar a ultima versão que está no github e atualizar localmente (binario e skills). (Preceded by a constitution amendment, v1.1.0: the 'Public command surface' Architecture Constraint was expanded from 'misterspec init only' to explicitly include these two flags, since they otherwise conflicted with that frozen constraint. Clarified with the user: --update requires explicit confirmation before replacing the running binary; the downloaded binary is verified against a published SHA-256 checksum before use; and --update also reinstalls Skills for whatever agent integration is already recorded in the current project's .misterspec/install.json.)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Check the installed version (Priority: P1)

A developer wants to know exactly which version of `misterspec` they have installed — for troubleshooting, for confirming an update actually took effect, or before reporting a bug.

**Why this priority**: The simpler of the two commands, and a prerequisite building block `--update` itself needs (to know the current version before deciding whether a newer one exists).

**Independent Test**: Run `misterspec --version` and confirm it prints a specific version identifier matching the release the binary was actually built from, then exits — no other side effect.

**Acceptance Scenarios**:

1. **Given** a `misterspec` binary built from a tagged release, **When** a developer runs `misterspec --version`, **Then** the exact release version is printed and the command exits successfully, with no network access and no filesystem changes.
2. **Given** a `misterspec` binary built without an embedded version (e.g. a local development build), **When** a developer runs `misterspec --version`, **Then** it prints a clear indication that no release version is embedded (e.g. "development build"), rather than a blank, misleading, or crashing output.

---

### User Story 2 - Check for and apply an available update (Priority: P1)

A developer wants to bring their locally installed `misterspec` binary up to date with the latest published release, without manually visiting GitHub, downloading a file, and replacing it by hand.

**Why this priority**: This is the other half of the entire feature — without it, `--version` alone only tells a developer they're behind, with no built-in way to act on that.

**Independent Test**: With an older binary installed and a newer release actually published, run `misterspec --update`; confirm it reports the current and available versions, asks for confirmation, and — once confirmed — the binary at the same path now reports the newer version via `--version`.

**Acceptance Scenarios**:

1. **Given** a newer release is published on GitHub than the one currently installed, **When** a developer runs `misterspec --update`, **Then** it reports the current version, the newest available version, and asks for explicit confirmation before changing anything.
2. **Given** the developer confirms, **When** the update proceeds, **Then** the correct binary asset for the current operating system and architecture is downloaded, its SHA-256 checksum is verified against the one published alongside it, and only after that verification succeeds is the currently installed binary replaced.
3. **Given** the developer declines the confirmation prompt, **When** `--update` exits, **Then** nothing on disk has changed — the binary and any Skill files are untouched.
4. **Given** the currently installed version is already the newest available, **When** a developer runs `misterspec --update`, **Then** it reports that no update is needed and exits, without downloading anything or asking for confirmation.
5. **Given** the update is applied inside a project that already has `misterspec` initialized (a `.misterspec/install.json` recording which agent integration was installed), **When** the update completes, **Then** that project's own installed Skills are also refreshed to match the newly installed version, without the developer having to separately re-run initialization.
6. **Given** `--update` is run outside any initialized `misterspec` project (no `.misterspec/install.json` present), **When** the binary update completes, **Then** it succeeds and simply reports that no project-level Skills were found to refresh, rather than failing.

### Edge Cases

- What happens when the downloaded binary's checksum does not match the published one? The update MUST stop before replacing anything, report the mismatch clearly, and leave the existing binary and Skills exactly as they were.
- What happens when GitHub is unreachable, or the check for a newer release otherwise fails (network error, rate limiting, GitHub outage)? `--update` MUST report this plainly and exit without changing anything — never silently do nothing and claim success, and never claim there's no update available when the check genuinely could not run.
- What happens when the current platform/architecture has no published binary asset for the latest release (e.g. a new, less common platform)? `--update` MUST report that no compatible asset was found, rather than downloading a mismatched binary or failing with a raw, confusing error.
- What happens if the process lacks permission to replace the binary at its own installed path (e.g. installed to a system directory without write access)? `--update` MUST report the permission problem clearly, and MUST NOT leave a partially-written or corrupted binary in place.
- What happens if a project's recorded agent integration (`.misterspec/install.json`) no longer matches any Skill-installation mechanism the current binary knows about (e.g. the integration was later removed from the tool)? `--update` MUST report this plainly as something it could not refresh, rather than silently skipping it or failing the whole update.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide `misterspec --version`, printing the exact version of the release the running binary was built from, with no network access and no filesystem changes.
- **FR-002**: When the running binary has no embedded release version (e.g. a local development build), `misterspec --version` MUST say so plainly rather than printing a blank or misleading value.
- **FR-003**: The system MUST provide `misterspec --update`, which checks the latest published release on GitHub and compares it to the running binary's own version.
- **FR-004**: If the latest published release is not newer than the currently running version, `--update` MUST report that no update is needed and MUST NOT download anything or prompt for confirmation.
- **FR-005**: If a newer release exists, `--update` MUST report both the current and the newest available version, and MUST require the developer's explicit confirmation before making any change.
- **FR-006**: If the developer declines confirmation, `--update` MUST exit leaving the installed binary and any project Skills completely unchanged.
- **FR-007**: Once confirmed, `--update` MUST download the binary asset matching the current operating system and architecture, and MUST verify it against a published SHA-256 checksum before using it in any way.
- **FR-008**: If checksum verification fails, `--update` MUST stop before replacing the installed binary, report the failure clearly, and leave the existing binary and Skills unchanged.
- **FR-009**: Only after successful checksum verification MUST `--update` replace the currently installed binary with the newly downloaded one.
- **FR-010**: If run inside a project directory that already has a recorded agent integration (`.misterspec/install.json`), a successful binary update MUST also reinstall that project's own Skills to match the newly installed version.
- **FR-011**: If run outside any initialized project (no `.misterspec/install.json` present), `--update` MUST still successfully update the binary and MUST report plainly that no project-level Skills were found to refresh — this MUST NOT be treated as a failure.
- **FR-012**: If the GitHub release check itself fails (network error, outage, rate limiting) or no compatible binary asset exists for the current platform, `--update` MUST report the specific problem plainly and exit without changing anything on disk.
- **FR-013**: If the process lacks permission to replace the installed binary, `--update` MUST report that plainly and MUST NOT leave a partially-written or corrupted binary in place.

### Key Entities

- **Release Version**: The version identifier associated with one published GitHub release (e.g. `v1.2.0`), embedded into a binary at build time; `--version` reports it, `--update` compares it against the latest published one.
- **Release Asset**: The platform/architecture-specific binary file published alongside a release, plus its published SHA-256 checksum — `--update`'s own download-and-verify target.
- **Project Agent Integration Record**: The existing `.misterspec/install.json` record of which agent integration a project already has Skills installed for — `--update` reads this (never writes it) to know what to refresh.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can determine the exact installed version of `misterspec` in one command, without consulting any external source.
- **SC-002**: A developer can bring their installation fully up to date — binary and, where applicable, project Skills — in one command plus one confirmation, without manually downloading or copying any file themselves.
- **SC-003**: 100% of update attempts either fully succeed (binary and, if applicable, Skills both updated, matching the newest published release) or fully fail with nothing changed — never a partially-applied update.
- **SC-004**: 100% of binaries `--update` installs have been verified against their published checksum before being put into use.

## Assumptions

- "The current project," for the purpose of refreshing Skills, is resolved the same way every other `misterspec` command already resolves its target project (the directory `--update` is run from, or its nearest ancestor containing `.misterspec/`) — this feature does not change that resolution logic.
- Publishing SHA-256 checksums alongside release binaries is a build/release-process change this feature requires (today's release process publishes binaries with no checksum file) — in scope for this feature's own implementation, not a separate prerequisite feature.
- Embedding a version string into the binary at build time is likewise a build/release-process change this feature requires (today's build has no version embedded at all).
- `--update` operates only on the single binary at its own currently-running path — it does not manage multiple installed copies, does not modify `PATH`, and does not offer to install `misterspec` somewhere it isn't already present (that remains a manual install step, unrelated to this feature).
- Exact prompt wording, output formatting, and the specific GitHub API endpoint/mechanism used to determine the latest release are planning-level detail, not fixed by this specification.
