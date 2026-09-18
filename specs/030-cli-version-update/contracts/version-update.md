# Contract: `misterspec --version` and `misterspec --update`

## `misterspec --version`

No network access, no filesystem changes. JSON envelope, matching `init`'s own existing "public command" convention:

```json
{"ok": true, "version": "v1.2.0"}
```

Or, for a build with no embedded version:

```json
{"ok": true, "version": "development build"}
```

## `misterspec --update`

### Up to date (no download, no prompt)

```json
{"ok": true, "update": {"current": "v1.2.0", "latest": "v1.2.0", "status": "up_to_date"}}
```

### Update available — confirmation required before anything changes

Prompts interactively (current vs. latest version shown); on decline:

```json
{"ok": true, "update": {"current": "v1.1.5-alpha", "latest": "v1.2.0", "status": "declined"}}
```

On confirmation, once the binary replace and (if applicable) Skill reinstall both succeed:

```json
{
  "ok": true,
  "update": {
    "current": "v1.1.5-alpha",
    "latest": "v1.2.0",
    "status": "updated",
    "skills_refreshed": true
  }
}
```

`skills_refreshed` is `false` (never absent) when no `.misterspec/install.json` was found — a normal, non-error outcome (FR-011):

```json
{
  "ok": true,
  "update": {
    "current": "v1.1.5-alpha",
    "latest": "v1.2.0",
    "status": "updated",
    "skills_refreshed": false
  }
}
```

### Failure shapes (JSON error envelope, matching every other command's own contract)

Checksum mismatch (binary and Skills left untouched):

```json
{"ok": false, "error": {"code": "checksum_mismatch", "message": "..."}}
```

GitHub check failed (network/outage/rate-limit):

```json
{"ok": false, "error": {"code": "release_check_failed", "message": "..."}}
```

No compatible asset for the current OS/architecture:

```json
{"ok": false, "error": {"code": "no_compatible_asset", "message": "..."}}
```

Insufficient permission to replace the installed binary:

```json
{"ok": false, "error": {"code": "permission_denied", "message": "..."}}
```

## `SHA256SUMS` File Format (published by `release.yml` alongside binary assets)

Standard `sha256sum`-compatible format, one line per asset:

```text
<64-hex-char-sha256>  misterspec-linux-amd64
<64-hex-char-sha256>  misterspec-linux-arm64
<64-hex-char-sha256>  misterspec-darwin-amd64
<64-hex-char-sha256>  misterspec-darwin-arm64
<64-hex-char-sha256>  misterspec-windows-amd64.exe
```

## Invariants

- `--version` never touches the network or the filesystem.
- `--update` never modifies anything on disk before both (a) explicit user confirmation and (b) successful checksum verification have both happened, in that order.
- A checksum mismatch, a declined confirmation, a failed release check, a missing compatible asset, or a permission error all leave the installed binary and any project Skills byte-for-byte unchanged.
- `skills_refreshed` reuses `internal/bootstrap`'s own existing install path — its own output is byte-identical to what `misterspec init --agent <id>` would produce for the same `kit.SkillsFS` content, never a separate implementation.
