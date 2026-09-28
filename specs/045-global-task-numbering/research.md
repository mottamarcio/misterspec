# Research: Numeração Global de Tasks

## Decision 1: Onde a nova função vive

**Decision**: Adicionar `NextTaskID(root string, cfg project.Configuration) (ids.EntityID, error)` em um novo arquivo `internal/ids/next_task_id.go`, composto inteiramente a partir de `Scan(root, cfg, Task)` + `NextID(result.IDs, Task, cfg.IDWidth)` — ambos já exportados e já usados exatamente dessa forma por `operations.Create` para Program/Feature/Spec.

**Rationale**: `ids.Scan(root, cfg, Task)` já enumera Task headings (`## TASK-NNN`) em **todo** o projeto — `scanTaskHeadings` (scan.go:287-298) mescla `taskClaim`s de todos os Specs num único mapa `número -> paths`, sem nenhum escopo por Spec. `ids.NextID` já calcula o maior número existente de um tipo + 1, projeto inteiro, exatamente a mesma função que `operations.Create` já usa para Program/Feature/Spec. Ou seja: **a primitiva de baixo nível já faz exatamente o que a Spec pede** — não falta nenhuma lógica de varredura nova, só uma função pequena que compõe as duas chamadas e retorna o resultado tipado, para não obrigar cada chamador (CLI, futuros testes) a repetir `Scan`+`NextID` na mão.

**Alternatives considered**:
- Colocar a composição direto no comando CLI (`internal/cli/internalcmd/next_task_id.go`), sem uma função dedicada em `internal/ids`. Rejeitado: Principle VI (DRY) — a mesma composição `Scan(Task)+NextID` seria útil a qualquer futuro chamador não-CLI (ex.: um teste, ou uma futura verificação em `internal validate`); uma função exportada em `ids` é o lugar natural, mesmo pacote que já expõe `NextID`.
- Adicionar um método a `ScanTasks`/`TaskScanResult` em vez de uma função livre. Rejeitado: `TaskScanResult` já serve um propósito diferente e mais rico (duplicatas, colisões, vistas por Spec) — misturar "próximo número" ali obrigaria todo chamador de `ScanTasks` (que hoje nunca precisa do próximo número) a pagar esse cálculo. Uma função livre e pequena, ao lado de `NextID`, é mais simples (Principle IV).

## Decision 2: Forma do novo comando CLI

**Decision**: `misterspec internal next-task-id` — sem argumentos além de `--dir` (padrão `.`, mesma convenção de todo comando já existente). Resposta: `{"ok": true, "next_task_id": "TASK-048"}`.

**Rationale**: Mesma forma exata de `migration-check-tasks` (comando somente-leitura, sem argumentos, `--dir` only) — reaproveita o padrão já estabelecido em vez de inventar um novo. Um único campo de resultado (`next_task_id`, já formatado via `EntityID.String()`) é consistente com como `internal resolve`/`internal inspect` retornam IDs já formatados como string, não como número + prefixo separados.

**Alternatives considered**:
- Aceitar um Spec como argumento (`internal next-task-id SPEC-045`) para também informar se aquele Spec específico já tem Tasks. Rejeitado: o valor pedido pela Spec é puramente o próximo número **global** — que Spec vai usá-lo é irrelevante para o cálculo (ao contrário de Program/Feature/Spec, que precisam do pai para resolver o caminho canônico do novo artefato). Adicionar um argumento sem uso real violaria Principle IV (YAGNI).
- Devolver também o maior número já visto e a lista de Specs que o usam (como `migration-check-tasks` faz para colisões). Rejeitado: informação que ninguém pediu — o Skill só precisa do próximo número para escrever o cabeçalho; presente somente para "ser mais informativo" é escopo não solicitado (Principle IV).

## Decision 3: Integração no Skill `mister-tasks`

**Decision**: Atualizar `kit/skills/mister-tasks/SKILL.md` em três pontos:
1. **Deterministic Operations** — adicionar `internal next-task-id` à lista de operações obrigatórias, com a mesma redação de "chamar antes de numerar cada nova Task".
2. **Procedure**, passo 5 — trocar "Number each `## TASK-NNN` sequentially, checking existing Tasks first so numbers are never reused" por uma instrução que chama `internal next-task-id` antes de escrever cada cabeçalho, e incrementa a partir daquele valor para Tasks adicionais na mesma execução (evitando N chamadas para N Tasks quando o Skill já sabe que vai escrever várias em sequência).
3. A frase "No allocator operation exists for Task IDs" deixa de ser verdade e precisa ser corrigida para refletir que agora existe uma operação — mas ainda somente-leitura/consultiva, não uma operação de criação; a autoria do cabeçalho continua sendo edição direta do Skill em `tasks.md` (nenhuma mudança na Fronteira de Allowed Modifications).

**Rationale**: `kit/skills/mister-tasks/SKILL.md` é o único lugar hoje que instrui como numerar uma Task nova — é onde a confusão relatada pelo usuário realmente se origina. Sem essa edição, o novo comando existiria mas nenhum Skill o chamaria, e o comportamento observado pelo usuário não mudaria.

**Alternatives considered**:
- Fazer `internal validate` chamar `next-task-id` automaticamente e sugerir a renumeração de uma Task recém-criada com número colidente. Rejeitado: mudaria o comportamento hoje aceito de `internal validate` (spec FR-004/FR-005 exigem que o comportamento estrutural existente não mude) e reintroduziria a possibilidade de "sugestão automática de renumeração" — fora do escopo desta Spec e potencialmente perigoso (renumerar afeta referências `Depends on:` já escritas).

**Correção encontrada durante a implementação**: `kit/skills/mister-tasks/SKILL.md` não é editado diretamente — é gerado a partir de `internal/skillgen/manifests_data.go` (`go generate ./internal/skillgen/...`), verificado por `TestSkillgenDrift`. As três edições acima foram aplicadas ao manifesto Go, não ao Markdown gerado. Um segundo teste, `TestSkillsContent_PlanTasksImplementAnalyze` (`internal/example/skills_content_test.go`), mantém um allowlist estático de comandos `internal` reais — `next-task-id` precisou ser adicionado a `knownInternalCommands` e a `skillOperationsAllowlist["mister-tasks"]`. Um comentário desatualizado em `kit/skills/mister-implement`'s manifesto ("`TASK-NNN` is numbered per-Spec (no global allocator...)") também foi corrigido, já que esta funcionalidade o tornava impreciso.

## Decision 4: Concorrência

**Decision**: Nenhum mecanismo de trava/reserva é adicionado. O comando é uma leitura pura, exatamente como a already-existing preview implícita em `operations.Create`'s uso de `NextID` antes de gravar (que é protegido por `lock.Acquire` **apenas** durante a criação real de Program/Feature/Spec/Knowledge/Learning — Task nunca passou por esse `lock`, porque nunca teve uma operação de criação).

**Rationale**: Documentado explicitamente no spec.md (Edge Cases, Assumptions) como uma limitação aceita — dois agentes concorrentes decompondo dois Specs diferentes podem, em teoria, receber o mesmo número sugerido antes de qualquer um gravar. Isso não é pior do que o status quo (Task nunca teve proteção de concorrência), e o diagnóstico de colisão (`internal migration-check-tasks`) continua sendo a rede de segurança que detecta o resultado depois do fato — exatamente como já funciona hoje.

**Alternatives considered**:
- Adicionar `lock.Acquire` ao redor da nova operação, mesmo sendo somente-leitura. Rejeitado: Principle III descreve o lock como existindo para "proteger uma única operação de mutação" — esta operação não muta nada; adicionar um lock a uma leitura pura seria complexidade não solicitada (Principle IV) que não resolve a corrida real (a corrida está entre a leitura e a escrita manual subsequente do Skill, que acontece fora do processo `misterspec` e não pode ser travada por ele).
