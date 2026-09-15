# misterspec

An agent-native framework for Spec-Driven Development, from project
knowledge to validated code.

**misterspec** is a single, static Go binary that gives a coding agent
a deterministic backbone: entity IDs, canonical paths, content
fingerprints, structural validation, and a local, deterministic
**Context Engine** for ranked/budgeted retrieval — so the agent's own
reasoning is spent on what actually requires judgment, and nothing
else. Every one of these is exposed as a stable, machine-readable JSON
command; the agent's own understanding, design, and code changes are
never touched.

## Why

A coding agent working from Markdown specs alone has to re-derive
mechanical answers every time — the next free ID, whether an
artifact's frontmatter is well-formed, what a Spec actually depends
on. None of that requires judgment. misterspec computes it once,
deterministically, and lets the agent call that instead of reasoning
its way to the same answer.

## Install

**Download a pre-built binary** (no Go toolchain, no repository clone needed) from the
[latest release](https://github.com/mottamarcio/misterspec/releases/latest) — pick the
archive for your OS/architecture, then:

```sh
chmod +x misterspec
./misterspec --help
```

**Or build from source** (requires Go 1.23.4 or newer):

```sh
git clone https://github.com/mottamarcio/misterspec.git
cd misterspec
go build -o misterspec ./cmd/misterspec
./misterspec --help
```

## Quickstart

```sh
# Inside the repository you want to manage with misterspec:
misterspec init --agent claude-code

# Confirm the project is healthy at any point:
misterspec internal status
```

Once initialized, the rest of the workflow happens through your coding
agent's own installed Skills (`/create-constitution`, `/create-specs`,
`/create-plan`, `/create-tasks`, `/implement`, `/analyze`, …) —
misterspec's own deterministic commands run underneath them
automatically.

## Documentation

The full reference — the project's core concept, every shipped
feature explained, and a complete command reference with examples —
is published at:

**[https://mottamarcio.github.io/misterspec/](https://mottamarcio.github.io/misterspec/)**

## License

[MIT](./LICENSE)
