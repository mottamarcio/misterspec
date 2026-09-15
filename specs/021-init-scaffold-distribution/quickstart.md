# Quickstart: Init Scaffolding and Binary Distribution

## 1. Confirm the templates directory is gone

```sh
misterspec init --agent claude-code --dir /tmp/quickstart-project
ls /tmp/quickstart-project
# .claude/  .misterspec/  ai/   — no "templates/"
```

## 2. Confirm the full directory hierarchy exists, committable

```sh
find /tmp/quickstart-project/ai -type d
# ai, ai/raw, ai/knowledge, ai/memory, ai/memory/learnings, ai/programs
find /tmp/quickstart-project/ai -name .gitkeep
cd /tmp/quickstart-project && git init -q && git add -A && git status --short
# every ai/... directory shows as staged, via its own .gitkeep
```

## 3. Confirm creating a real artifact still works normally

```sh
misterspec internal create knowledge --slug auth-model --dir /tmp/quickstart-project
# succeeds; ai/knowledge/ now has both .gitkeep and the new KNOW-001-auth-model.md
```

## 4. Download a release binary without cloning

```sh
curl -LO https://github.com/mottamarcio/misterspec/releases/latest/download/misterspec_linux_amd64
chmod +x misterspec_linux_amd64
./misterspec_linux_amd64 init --agent claude-code
```

## 5. Trigger the release workflow for a new tag

```sh
git tag v1.0.1
git push origin v1.0.1
# .github/workflows/release.yml builds all 5 binaries and attaches
# them to the resulting GitHub Release automatically
```
