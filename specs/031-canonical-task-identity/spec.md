# Feature Specification: Modelo canônico de tarefas e identidade composta

**Feature Branch**: `031-canonical-task-identity`
**Created**: 2026-09-18
**Status**: Draft
**Input**: User description: "PROP-01 — Modelo canônico de tarefas e identidade composta: tornar inequívoca a identidade das tarefas e centralizar sua interpretação. As Skills atuais descrevem numeração local por Spec, enquanto o scanner e a validação global agregam números entre Specs."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Resolver uma tarefa sem ambiguidade entre Specs (Priority: P1)

Um agente (via Skill ou comando) precisa localizar uma tarefa específica para inspecioná-la, validá-la ou prepará-la para execução. Hoje, se duas Specs diferentes contêm cada uma um `TASK-001`, o sistema não sabe a qual das duas o agente se refere.

**Why this priority**: Sem isso, qualquer operação que dependa de identificar "a" tarefa (inspeção, validação, preparação de execução) pode silenciosamente pegar a tarefa errada ou falhar de forma pouco clara. É a base sobre a qual as demais propostas (PROP-02, PROP-04) dependem.

**Independent Test**: Criar duas Specs, cada uma com um `tasks.md` contendo `TASK-001`. Pedir a resolução de `SPEC-A:TASK-001` e confirmar que apenas a tarefa da Spec A é retornada, e o mesmo para a Spec B.

**Acceptance Scenarios**:

1. **Given** duas Specs distintas, cada uma com uma tarefa `TASK-001` em seu próprio `tasks.md`, **When** o sistema resolve `SPEC-A:TASK-001`, **Then** apenas a tarefa pertencente à Spec A é retornada, sem menção à tarefa da Spec B.
2. **Given** o mesmo cenário, **When** o sistema resolve `SPEC-B:TASK-001`, **Then** apenas a tarefa pertencente à Spec B é retornada.
3. **Given** uma referência a uma tarefa sem indicar a Spec de contexto, **When** mais de uma Spec no projeto possui o mesmo número de tarefa, **Then** o sistema recusa a resolução automática entre Specs e retorna um erro acionável pedindo a Spec de contexto, em vez de escolher uma tarefa arbitrariamente ou tratar os números como a mesma identidade.

---

### User Story 2 - Detectar duplicatas apenas dentro da mesma Spec (Priority: P1)

Quem executa a validação estrutural do projeto precisa que duplicatas de `TASK-NNN` sejam sinalizadas quando ocorrem dentro do mesmo `tasks.md`, mas não quando o mesmo número aparece em Specs diferentes — que é o uso normal e esperado da numeração local por Spec.

**Why this priority**: A validação atual (`ids.Scan` + verificação de duplicatas) trata todos os `TASK-NNN` do projeto como um único espaço de números, produzindo falsos positivos de duplicidade entre Specs não relacionadas e mascarando o problema real: duas tarefas com o mesmo número dentro da mesma Spec.

**Independent Test**: Rodar a validação estrutural em um projeto com duas Specs que cada uma tem um `TASK-001` legítimo e distinto; confirmar que nenhum erro de duplicidade é reportado entre elas. Depois, inserir duas seções `## TASK-001` na mesma Spec e confirmar que a validação sinaliza a duplicidade, apontando o arquivo e as linhas envolvidas.

**Acceptance Scenarios**:

1. **Given** duas Specs cada uma com um `TASK-001` legítimo, **When** a validação estrutural do projeto é executada, **Then** nenhuma duplicidade é reportada para esse número entre as duas Specs.
2. **Given** uma única Spec cujo `tasks.md` contém duas seções `## TASK-001`, **When** a validação estrutural é executada (seja no projeto inteiro ou apenas naquela Spec), **Then** o sistema reporta uma duplicidade identificando a Spec, o número duplicado e as localizações envolvidas.

---

### User Story 3 - Migrar projetos existentes sem renumeração automática (Priority: P2)

Quem já tem Specs e tarefas criadas antes desta mudança precisa entender se algum projeto seu ficou exposto a resoluções ambíguas sob o comportamento antigo, sem que o sistema renumere tarefas silenciosamente ou quebre referências existentes.

**Why this priority**: Protege trabalho já existente. É posterior às User Stories 1 e 2 porque depende do novo modelo já estar implementado para poder ser comparado ao comportamento anterior.

**Independent Test**: Rodar o diagnóstico de migração em um projeto com Specs pré-existentes que continham números de tarefa repetidos entre Specs; confirmar que o diagnóstico lista os casos afetados e que nenhuma tarefa foi renumerada como efeito colateral.

**Acceptance Scenarios**:

1. **Given** um projeto existente com `TASK-NNN` repetidos entre Specs diferentes (comportamento válido sob o novo modelo, mas que colidia sob o antigo scanner global), **When** o diagnóstico de migração é executado, **Then** cada caso potencialmente afetado pelo comportamento anterior é listado com a Spec e o número envolvidos.
2. **Given** o mesmo projeto, **When** a migração é aplicada, **Then** nenhum `TASK-NNN` existente é renumerado automaticamente.

---

### Edge Cases

- Referência a uma Spec inexistente combinada com um número de tarefa válido em outra Spec: deve falhar com erro claro, não resolver por acaso para a tarefa de outra Spec.
- Duas seções `## TASK-NNN` idênticas dentro do mesmo `tasks.md` (mesmo número, mesmo título): ainda deve ser tratado como duplicidade.
- Um cabeçalho de tarefa malformado (número ausente ou não numérico) não deve ser contado como uma identidade válida nem quebrar a varredura das demais tarefas.
- Validação rodada sobre uma única Spec isolada (`validate SPEC-###`) deve produzir o mesmo diagnóstico de duplicidade/coerência que a validação do projeto inteiro produziria para aquela Spec.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST definir um modelo de identidade de tarefa compartilhado, com Spec proprietária, ID local da tarefa, e uma chave composta derivada (por exemplo, `SPEC-014:TASK-003`) que identifica a tarefa de forma inequívoca em todo o projeto.
- **FR-002**: O sistema MUST aceitar e resolver a sintaxe de identidade composta (`SPEC-###:TASK-###`) para exatamente uma tarefa, em qualquer operação que resolva referências de tarefa (inspeção, validação, preparação de execução).
- **FR-003**: Ao receber uma referência de tarefa sem Spec de contexto explícita (`TASK-###` isolado), o sistema MUST exigir contexto de Spec — vindo do argumento do comando, do escopo de invocação, ou de uma Spec única já implícita no projeto — antes de tentar resolver a tarefa; na ausência desse contexto, quando mais de uma Spec reivindica o mesmo número, o sistema MUST retornar um erro acionável em vez de escolher uma tarefa arbitrariamente ou agregar duplicatas entre Specs.
- **FR-004**: A detecção de duplicidade de ID de tarefa MUST ser escopada por Spec: dois `TASK-NNN` com o mesmo número em Specs diferentes NÃO MUST ser reportados como duplicidade; duas ocorrências do mesmo número dentro do `tasks.md` de uma única Spec MUST ser reportadas como duplicidade.
- **FR-005**: As operações de inspeção, validação e seleção de tarefas MUST compartilhar um único parser/resolvedor de `tasks.md` e do modelo de identidade, em vez de cada uma reimplementar sua própria varredura e resolução.
- **FR-006**: Comandos e integrações (CLI, Skills) que recebem uma tarefa como argumento MUST aceitar e propagar o contexto de Spec necessário para resolução inequívoca (seja via forma composta, seja via argumento de Spec acompanhando o argumento de tarefa).
- **FR-007**: O sistema MUST NOT renumerar automaticamente cabeçalhos `TASK-NNN` existentes como parte da adoção do novo modelo de identidade.
- **FR-008**: O sistema MUST fornecer um diagnóstico de migração que identifique tarefas cuja resolução seria afetada pela mudança do comportamento de varredura global anterior para o novo comportamento escopado por Spec, sem alterar o conteúdo dos artefatos.
- **FR-009**: Dentro do contexto de uma Spec já conhecida (por exemplo, ao listar ou apresentar as tarefas daquela Spec), o sistema MUST continuar aceitando e exibindo a forma local curta (`TASK-003`), sem exigir a forma composta do usuário quando o contexto já está estabelecido.
- **FR-010**: Um cabeçalho de tarefa malformado (número ausente, não numérico, ou fora do padrão esperado) MUST NOT ser tratado como identidade de tarefa válida e MUST NOT interromper a varredura das demais tarefas do mesmo arquivo.
- **FR-011**: A validação de uma única Spec isolada (`validate SPEC-###`) MUST produzir, para as tarefas daquela Spec, os mesmos diagnósticos de duplicidade e identidade que a validação do projeto inteiro produziria para essa mesma Spec.

### Key Entities

- **Task (Tarefa)**: unidade de trabalho executável descrita em `tasks.md`, com Spec proprietária, ID local (`TASK-NNN`), estado, requisitos atendidos, dependências e método de verificação. Sua identidade completa só existe em conjunto com a Spec proprietária.
- **Spec**: artefato que agrupa um conjunto de tarefas em seu próprio `tasks.md`; é o escopo de numeração local das tarefas que contém.
- **Composite Task Identity (Identidade composta de tarefa)**: chave derivada da combinação Spec + ID local (por exemplo, `SPEC-014:TASK-003`) usada para referenciar uma tarefa de forma inequívoca fora do contexto de sua própria Spec.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Duas Specs diferentes podem cada uma conter uma tarefa `TASK-001` sem que nenhuma das duas seja reportada como inválida ou duplicada pela validação estrutural.
- **SC-002**: Toda referência de tarefa na forma composta resolve a exatamente uma tarefa, com 0% de ambiguidade observada ao testar contra o conjunto de Specs existentes no projeto.
- **SC-003**: Em 100% dos casos em que uma referência de tarefa sem contexto de Spec é ambígua entre Specs, o sistema retorna uma mensagem de erro clara e acionável, em vez de uma resolução incorreta ou arbitrária.
- **SC-004**: Após a adoção do novo modelo, todas as Specs existentes no projeto passam na validação estrutural sem qualquer renumeração manual de tarefas.
- **SC-005**: Inspeção e validação de uma mesma tarefa produzem resultados de identidade consistentes entre si (mesma Spec, mesmo ID local, mesma chave composta) em 100% dos casos testados.

## Assumptions

- O formato de autoria de tarefas em Markdown (`## TASK-NNN — Título`) permanece o mesmo; apenas a camada de identidade e resolução é alterada, não a sintaxe visível ao autor humano.
- A forma local curta (`TASK-003`) continua válida como forma de apresentação e de entrada quando o contexto de Spec já está estabelecido; a forma composta é necessária apenas quando esse contexto não está implícito.
- Projetos existentes podem conter números de tarefa repetidos entre Specs diferentes; isso é tratado como uso válido sob o novo modelo, não como erro a corrigir.
- Nenhuma tarefa existente é renumerada automaticamente; qualquer necessidade de renumeração fica a critério do autor humano, informado pelo diagnóstico de migração.
- Esta especificação cobre o modelo de identidade e resolução de tarefas; a formalização completa de cobertura de requisitos e dependências (ciclos, cobertura declarada) é tratada pela proposta subsequente (PROP-02), que depende deste modelo.
