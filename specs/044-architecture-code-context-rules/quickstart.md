# Quickstart: Validating Architecture Rules & Code Context

Prerequisites: a built `misterspec` binary from this branch, a Go
project (has `go.mod` at its root) initialized with `misterspec init`.

## 1. A forbidden dependency is detected with a reproducible location (User Story 1)

Declare a rule in `.misterspec/config.yaml`:

```yaml
architecture_rules:
  - kind: forbidden_dependency
    from: internal/artifacts/**
    to: internal/cli/internalcmd
```

Introduce (or already have) an import of `internal/cli/internalcmd`
somewhere under `internal/artifacts/`, then:

```sh
misterspec internal check-architecture
```

**Expected**: `ok: true`; `architecture.results` includes one entry
with `status: "fail"`, the exact file and line of the forbidden
import. Running the command again with no code change produces the
identical `path`/`line`.

## 2. No adapter never means "pass" (User Story 2)

Run the same command against a project with no `go.mod` at its root
(or temporarily rename it):

```sh
misterspec internal check-architecture
```

**Expected**: `architecture.adapter` is `""`; every declared rule in
`architecture.results` has `status: "not_evaluated"` and
`reason: "no_adapter_for_project"` — never `"pass"`.

## 3. An unsupported rule kind is `not_evaluated`, not silently skipped

Declare a `required_contract` rule (not supported by the Go adapter in
this feature's first version):

```sh
misterspec internal check-architecture
```

**Expected**: that rule's own entry in `architecture.results` has
`status: "not_evaluated"`, `reason: "rule_kind_unsupported"` — still
present in the response, not omitted.

## 4. Code relevant to a Task's Scope arrives without reading the whole project (User Story 3)

With a Task whose body declares `Scope: internal/auth/refresh.go`:

```sh
misterspec internal prepare SPEC-014 --task TASK-003
```

**Expected**: `preparation.code_context` includes at least one entry
for `internal/auth/refresh.go` — its own declaration(s) at signature
level at minimum, each with `path`/`location`/`fingerprint` — without
needing a separate call or manual file read to discover it.

## 5. A Scope path that no longer exists is reported, not silently dropped

Point a Task's `Scope:` at a file that has been deleted or renamed,
then request its preparation.

**Expected**: `preparation.code_scope_not_found` names that path
explicitly; `code_context` still includes whatever other Scope paths
did resolve.

## 6. Fingerprints stay consistent when code hasn't changed

Run `internal prepare` for the same Task twice with no source change
in between.

**Expected**: every `code_context` entry's own `fingerprint` is
byte-identical across both calls (spec SC-004).

## Related commands

- `misterspec internal prepare` (034) — the command `code_context`/
  `code_scope_not_found` extend; every other field is unchanged.
- `misterspec internal context --mode package` (033) — the same
  `PackageItem` shape `code_context` entries reuse.
- `misterspec internal eval-retrieval`/`eval-compare` (037) — the
  protocol spec SC-005 relies on to measure code-context retrieval
  against full-file reads; not re-implemented by this feature.
