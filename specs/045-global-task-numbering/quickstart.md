# Quickstart: Numeração Global de Tasks

Cenários manuais para validar a funcionalidade fim a fim, após a
implementação (Phase 2/Implementação).

## Pré-requisitos

- Binário `misterspec` compilado (`go build ./cmd/misterspec`).
- Um diretório de projeto MisterSpec de teste (`misterspec init` ou um
  fixture com `.misterspec/config.yaml`).

## Cenário 1 — Projeto sem nenhuma Task

```sh
misterspec internal next-task-id --dir /tmp/proj-vazio
```

**Esperado**: `{"ok": true, "next_task_id": "TASK-001"}`.

## Cenário 2 — Maior número existe em outro Spec

Crie duas Tasks em Specs diferentes, uma delas com o maior número do
projeto:

```text
ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md
  ## TASK-001 — ...
  ## TASK-047 — ...

ai/programs/PRG-001/features/FEAT-002/specs/SPEC-002/tasks.md
  (vazio ainda)
```

```sh
misterspec internal next-task-id --dir /tmp/proj-teste
```

**Esperado**: `{"ok": true, "next_task_id": "TASK-048"}` — nunca
`TASK-001`, mesmo o Spec "atual" (SPEC-002) não tendo nenhuma Task
própria ainda (spec.md User Story 1, Acceptance Scenario 1).

## Cenário 3 — Somente-leitura

```sh
md5sum ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md
misterspec internal next-task-id --dir /tmp/proj-teste
md5sum ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md
```

**Esperado**: o hash não muda — nenhuma escrita ocorreu (spec FR-003,
contracts/next-task-id-contract.md §3).

## Cenário 4 — Projeto pré-existente com colisão histórica permanece intacto

Usando o próprio repositório do MisterSpec (Specs 001–044, que já têm
`TASK-001` repetido em vários Specs por design):

```sh
misterspec internal migration-check-tasks --dir .
```

**Esperado**: reporta exatamente as mesmas colisões de antes da
implementação desta funcionalidade — nenhuma mudança de comportamento
(spec FR-005, User Story 3).

## Cenário 5 — Skill `mister-tasks` usa o novo comando

Após a atualização de `kit/skills/mister-tasks/SKILL.md` (research.md
Decision 3), inspecionar manualmente o arquivo e confirmar que:

1. `internal next-task-id` aparece na seção **Deterministic
   Operations** como obrigatória.
2. O passo 5 do **Procedure** instrui chamá-lo antes de escrever cada
   novo `## TASK-NNN`.
3. A frase "No allocator operation exists for Task IDs" foi corrigida
   para refletir que a operação existe, mas continua somente-leitura/
   consultiva (a autoria do cabeçalho continua manual/direta no
   `tasks.md`).
