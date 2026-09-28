# Implementation Plan: Numeração Global de Tasks

**Branch**: `045-global-task-numbering` | **Date**: 2026-09-28 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/045-global-task-numbering/spec.md`

## Summary

`Program`, `Feature` e `Spec` já são numerados de forma globalmente
incremental hoje (`internal/ids.Scan` enumera cada tipo em todo o
projeto via glob; `internal/ids.NextID` calcula o maior número
existente + 1, nunca reiniciando por Program/Feature pai). Só `Task`
reinicia por design (031-canonical-task-identity): um número de Task só
precisa ser único dentro do seu próprio Spec, e a forma composta
`SPEC-###:TASK-###` resolve qualquer ambiguidade. Esta funcionalidade
não reverte esse modelo — ela dá ao Skill que decompõe Tasks
(`kit/skills/mister-tasks`) uma operação determinística e somente-
leitura para descobrir o próximo número de Task nunca usado em nenhum
Spec do projeto, reaproveitando as mesmas primitivas (`ids.Scan` +
`ids.NextID`) já usadas por Program/Feature/Spec, mas aplicadas ao tipo
`Task` com escopo de projeto inteiro em vez de escopo por Spec.

## Technical Context

**Language/Version**: Go 1.23 (mesmo módulo/toolchain do repositório)
**Primary Dependencies**: nenhuma nova — reaproveita `internal/ids`
(`Scan`, `NextID`, já exportados) e `github.com/spf13/cobra` (já usado
por todo `internal/cli/internalcmd`)
**Storage**: N/A — nenhum estado persistido; a operação é uma função
pura do que já existe em disco no momento da chamada (Constitution
Principle III)
**Testing**: `go test` — testes unitários para a nova função em
`internal/ids`, mais um teste de integração de sistema de arquivos para
o novo comando CLI, seguindo exatamente o padrão já usado por
`internal/cli/internalcmd/migration_check_tasks_test.go`
**Target Platform**: CLI multiplataforma (Linux/macOS/Windows) — mesmo
binário `misterspec` já distribuído
**Project Type**: CLI / operação determinística interna (projeto Go
único, sem frontend/backend separados)
**Performance Goals**: mesma ordem de grandeza da varredura já existente
para `internal migration-check-tasks` (um glob + parse de todo
`tasks.md` do projeto) — nenhum requisito de desempenho novo
**Constraints**: somente-leitura; não pode alterar o comportamento já
existente de detecção de duplicata dentro do mesmo Spec nem o
diagnóstico de colisão entre Specs (spec FR-004/FR-005); não introduz
nenhuma operação de criação/escrita para Task
**Scale/Scope**: uma função nova exportada em `internal/ids`, um novo
comando CLI somente-leitura em `internal/cli/internalcmd`, e uma
atualização do Skill `kit/skills/mister-tasks/SKILL.md` para consultar
esse comando antes de numerar cada nova Task — nenhum pacote novo

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Semantic/Deterministic Separation** — PASS. "Qual o próximo
  número de Task ainda não usado no projeto" é inteiramente computável
  a partir da estrutura do repositório (nenhum julgamento semântico
  envolvido) — exatamente por isso pertence ao binário Go, não a uma
  estimativa do LLM. É a própria razão de existir desta funcionalidade.
- **II. Deterministic Operations são a única primitiva de mutação (sem
  sequência dividida `next-id` → `scaffold`)** — PASS, com raciocínio
  explícito: Task nunca teve (e esta funcionalidade não introduz) uma
  operação própria de criação/scaffold — sua autoria já é, hoje,
  edição direta de texto em `tasks.md` pelo Skill/agente
  (`kit/skills/mister-tasks/SKILL.md`: "No allocator operation exists
  for Task IDs — headings are authored directly into tasks.md"). A
  restrição da Principle II mira o risco de duas chamadas a um
  `next-id` retornarem o mesmo número e ambas prosseguirem para um
  `scaffold` que grava um arquivo com esse ID — um risco que só existe
  quando existe, de fato, uma etapa de criação separada a ser
  corrompida por essa divisão. Para Task não existe essa etapa: a nova
  operação é uma consulta estrutural somente-leitura, da mesma
  categoria de `internal inspect`/`internal resolve`/`internal
  migration-check-tasks` (que já leem o projeto e relatam fatos, sem
  jamais mutar nada) — não uma nova primitiva de mutação, nem a
  metade que faltava de uma sequência de criação.
- **III. Filesystem como fonte única de verdade** — PASS. A operação
  nova é puramente derivada da varredura de `tasks.md` já existente no
  disco (mesmo padrão de `NextID`); nenhum contador persistido, nenhum
  banco de dados.
- **IV. YAGNI / Configuração mínima** — PASS. Nenhum pacote novo;
  estende `internal/ids` com uma função (composição de `Scan`+`NextID`
  já existentes), adiciona um comando CLI fino seguindo exatamente a
  forma de `migration_check_tasks.go`, e edita um Skill já existente.
  Nenhum campo novo de configuração.
- **V. Test-First Discipline** — a nova função em `internal/ids` e o
  novo comando CLI recebem testes unitários/de integração de sistema
  de arquivos escritos antes/junto da implementação (tasks.md vai
  aplicar TDD task-a-task).
- **VI. Clean Code & SOLID** — PASS. Reaproveita `ids.Scan`+`ids.NextID`
  já exportados em vez de reimplementar varredura de Task; a nova
  função é uma composição fina, não lógica duplicada.
- **VII. Explicit Mutation Boundaries** — PASS trivialmente: a operação
  não muta nenhum artefato, então nenhuma fronteira de propriedade de
  artefato está em risco.
- **VIII. Safety by Construction** — PASS trivialmente: somente-
  leitura, nenhuma escrita em disco, nenhuma resolução de caminho além
  do glob que `ids.Scan` já realiza.
- **IX. Transparent, Machine-Readable Contracts** — PASS. O novo
  comando segue o envelope JSON já existente (`WriteSuccess`/
  `WriteError`), com um único campo de resultado (`next_task_id`).

Nenhuma violação a justificar — Complexity Tracking fica vazio.

## Project Structure

### Documentation (this feature)

```text
specs/045-global-task-numbering/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/ids/
├── ids.go                  # existing: Parse, ParseAny, ParseTaskRef, NextID
├── ids_test.go
├── scan.go                 # existing: Scan, ScanTasks, scanTaskHeadings, ...
├── scan_test.go
└── next_task_id.go         # NEW: NextTaskID(root, cfg) composing Scan+NextID
    next_task_id_test.go     # NEW: unit tests

internal/cli/internalcmd/
├── migration_check_tasks.go       # existing: read-only diagnostic (reference shape)
├── next_task_id.go               # NEW: `misterspec internal next-task-id` command
└── next_task_id_test.go          # NEW: filesystem integration tests

internal/cli/internal.go          # MODIFIED: register the new command

kit/skills/mister-tasks/SKILL.md  # MODIFIED: Procedure/Deterministic Operations
                                   # sections call the new command before
                                   # numbering each new Task; "No allocator
                                   # operation exists" line updated to reflect
                                   # the new (still non-mutating) operation
```

**Structure Decision**: Projeto Go único (sem frontend/backend
separados). A funcionalidade é inteiramente aditiva: uma função nova em
um pacote existente (`internal/ids`), um comando novo em um pacote
existente (`internal/cli/internalcmd`), e a edição de um Skill já
existente — nenhum diretório novo de alto nível.

## Complexity Tracking

> Sem violações a justificar — seção vazia deliberadamente.
