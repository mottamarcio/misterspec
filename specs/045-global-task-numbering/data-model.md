# Data Model: Numeração Global de Tasks

Esta funcionalidade não introduz nenhuma entidade persistida nova —
apenas uma operação determinística derivada de entidades já existentes.

## NextTaskID (operação, não entidade)

Uma função pura, somente-leitura, sem estado próprio:

```go
// internal/ids/next_task_id.go
func NextTaskID(root string, cfg project.Configuration) (EntityID, error)
```

| Campo (retorno `EntityID`) | Tipo | Descrição |
|---|---|---|
| `Type` | `EntityType` | Sempre `Task` |
| `Prefix` | `string` | Sempre `"TASK"` |
| `Number` | `int` | Maior número de Task encontrado em qualquer `tasks.md` do projeto + 1 (ou 1 se nenhuma Task existir ainda) |
| `Width` | `int` | `cfg.IDWidth`, o mesmo usado para formatar qualquer outro `EntityID` |

**Regras de derivação** (mapeando diretamente para spec.md FR-001/FR-003/FR-007):

- Computado como `NextID(Scan(root, cfg, Task).IDs, Task, cfg.IDWidth)` — reutiliza inteiramente as duas primitivas já existentes e já testadas.
- Nunca escreve em disco, nunca reserva o número calculado em nenhum estado (FR-003).
- Escopo é o projeto inteiro (`Scan` já enumera Task headings em todo `tasks.md` sob `cfg.ProgramsRoot`), nunca um único Spec — é essa mudança de escopo, em relação ao que o Skill fazia antes (ler só o `tasks.md` do Spec atual), que resolve FR-001/FR-002.
- Quando não existe nenhuma Task em nenhum Spec, retorna `Number: 1` (FR-007) — mesma regra de `NextID` para qualquer outro tipo.

## Entidades existentes referenciadas (sem alteração)

- **`ids.EntityID`** (`internal/ids/types.go`) — tipo de retorno já existente, sem mudanças.
- **`ids.ScanResult`** (`internal/ids/scan.go`) — usado internamente por `NextTaskID` via `Scan`, sem mudanças de forma.
- **Task heading (`## TASK-NNN — <título>`)** — convenção de texto já existente em `tasks.md`; esta funcionalidade não muda seu formato, apenas informa qual `NNN` usar a seguir.
