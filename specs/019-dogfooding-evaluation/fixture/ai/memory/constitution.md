---
type: constitution
---

## Principles

Condensed from misterspec's own real `.specify/memory/constitution.md`
— only the three principles actually cited by name across the real
011-018 Constitution Check tables.

### III. Filesystem Is the Single Source of Truth

Project state MUST be fully reconstructable from repository files at
any time. There MUST be no persistent authoritative counter, no
database, and no vector store as authoritative state.

### IV. Simplicity First — YAGNI & Minimal Configuration

New package layers, new machine commands for decisions that require
semantic reasoning, and new configuration fields MUST NOT be
introduced speculatively; they require demonstrated implementation
pressure, not anticipated need.

### IX. Transparent, Machine-Readable Contracts

Internal commands MUST default to structured JSON with no
decorative/ANSI output, and MUST distinguish `ok` (command executed)
from `valid`/`result` (semantic or structural outcome) as separate
concepts.
