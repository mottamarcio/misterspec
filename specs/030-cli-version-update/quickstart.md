# Quickstart: Validate `misterspec --version` and `misterspec --update`

## Prerequisites

- A local checkout of this repo on branch `030-cli-version-update` (or later, once merged).
- Go 1.23.4+ installed (`go version`).

## Automated validation

```sh
go test ./internal/selfupdate/...
go test ./internal/cli/...
go test ./internal/installer/...
```

Expected: all pass, including:
- Semver comparison tests covering this project's own real mixed-suffix tag history (`v1.1.5-alpha` correctly ranked below `v1.2.0`).
- Release-lookup and checksum-verification tests using an injected fake `http.RoundTripper` — no real network call.
- Atomic-replace tests for both the Unix direct-rename path and the Windows rename-aside path (exercised via an injectable "current OS" seam, not a real Windows runner).
- `internal/installer.WriteAtomicFile`'s existing callers still pass unchanged after gaining the new mode parameter.

## Manual validation (end-to-end scenario)

1. Build a binary with an embedded version: `go build -ldflags "-X github.com/mottamarcio/misterspec/internal/selfupdate.Version=v0.0.0-test" -o /tmp/misterspec-test ./cmd/misterspec`.
2. Run `/tmp/misterspec-test --version`. Confirm it prints `v0.0.0-test` via the JSON envelope, and that this made no network call (e.g. run with network disabled/offline and confirm it still works).
3. Build a binary with no `-ldflags` at all: `go build -o /tmp/misterspec-dev ./cmd/misterspec`. Run `/tmp/misterspec-dev --version`. Confirm it reports "development build" plainly, not a blank or misleading value.
4. Run `/tmp/misterspec-test --update` (a version deliberately older than the latest real published release). Confirm it reports both versions and prompts for confirmation before doing anything else.
5. Decline the prompt. Confirm `/tmp/misterspec-test` is byte-identical to what it was before (e.g. compare a checksum before/after).
6. Re-run and confirm this time. Confirm the binary at that same path now reports the newer version via `--version`, and that the download was verified against the published `SHA256SUMS` before being used (inspect the reported outcome/logs for the verification step).
7. Repeat step 6 inside a project directory that already has `misterspec init` run (a `.misterspec/install.json` present). Confirm the project's own installed Skills are refreshed to match the newly installed version — compare `.claude/skills/*/SKILL.md` (or whichever integration is recorded) against the new binary's own embedded `kit.SkillsFS` content, byte-for-byte.
8. Repeat step 6 outside any initialized project. Confirm it still succeeds and plainly reports that no project-level Skills were found to refresh — not a failure.
9. Manually corrupt a downloaded binary (or otherwise force a checksum mismatch) and confirm `--update` stops before replacing anything, reports the mismatch clearly, and the original binary is untouched.

## Expected outcome

Every acceptance scenario in `spec.md` (User Stories 1–2) passes as described, and `go test ./...` passes with no regressions.
