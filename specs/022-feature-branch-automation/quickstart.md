# Quickstart: Feature-Level Git Branch Automation

## 1. Creating a Feature starts a dedicated branch

```sh
cd /tmp/quickstart-project   # an existing, git-tracked misterspec project, on branch "dev"
git status --short           # clean

misterspec internal create feature --parent PRG-001 --slug "User Auth"
# {"ok":true,"created":{"id":"FEAT-007","type":"feature","path":"...","git":{"branch":"feat/FEAT-007-user-auth","created":true}}}

git branch --show-current
# feat/FEAT-007-user-auth  — not "dev", and readable at a glance

git log dev --oneline -1
# FEAT-007's own artifact commit does NOT appear here
```

`--slug` is optional — omitting it falls back to the bare `feat/FEAT-007`.

## 2. A Spec created afterward stays on the same branch

```sh
# still on feat/FEAT-007 from step 1
misterspec internal create spec --parent FEAT-007
# {"ok":true,"created":{"id":"SPEC-014","...","git":{"branch":"feat/FEAT-007"}}}
# no new branch — no "created" key, no "warning"

git branch --show-current
# still feat/FEAT-007
```

## 3. Creating a Spec from the wrong branch warns, doesn't block

```sh
git checkout dev
misterspec internal create spec --parent FEAT-007
# {"ok":true,"created":{"...","git":{
#   "branch":"feat/FEAT-007",
#   "warning":"current branch \"dev\" does not match parent Feature FEAT-007's own branch \"feat/FEAT-007\""
# }}}
# creation still succeeds; branch is NOT switched automatically
```

## 4. Re-running Feature creation after an interruption resumes cleanly

```sh
git branch feat/FEAT-008        # simulate a branch that already exists
misterspec internal create feature --parent PRG-001
# {"ok":true,"created":{"id":"FEAT-008","...","git":{"branch":"feat/FEAT-008","created":false}}}
# no error — the existing branch is checked out, not recreated
```

## 5. Non-Git project — zero behavior change

```sh
cd /tmp/quickstart-no-git       # misterspec project, no .git at all
misterspec internal create feature --parent PRG-001
# {"ok":true,"created":{"...","git":{"skipped_reason":"not_a_git_repo"}}}
# succeeds exactly as before this feature existed
```

## 6. Opting out entirely

```sh
# .misterspec/config.yaml
# git_branch_automation: false

misterspec internal create feature --parent PRG-001
# {"ok":true,"created":{"...","git":{"skipped_reason":"disabled"}}}
git branch --show-current
# unchanged — whatever branch was already checked out
```
