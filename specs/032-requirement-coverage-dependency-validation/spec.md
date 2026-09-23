# Feature Specification: Validação de cobertura de requisitos e dependências do SDD

**Feature Branch**: `032-requirement-coverage-dependency-validation`
**Created**: 2026-09-18
**Status**: Draft
**Input**: User description: "PROP-02 — Validação de cobertura e dependências do SDD: transformar as verificações estruturais prometidas pelas Skills em operações determinísticas, incluindo a relação entre requisitos, tarefas e dependências."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Detectar cobertura de requisitos por tarefas (Priority: P1)

Quem executa a validação estrutural do projeto precisa saber, de forma determinística, se todo requisito declarado em uma Spec tem ao menos uma tarefa que o atende, se toda tarefa referencia um requisito existente, e se nenhuma referência aponta para um requisito inexistente ou de outra Spec. Hoje, as Skills (`mister-tasks`, `mister-plan`, `mister-analyze`) já documentam a convenção `R#` para requisitos e `Serves: SPEC-###:R#` para a cobertura de uma tarefa, e `mister-tasks` chega a prometer que `internal validate SPEC-###` confirma "nenhuma Task referenciando um requisito inexistente" — mas esse comportamento não existe hoje no binário; a promessa da Skill não é cumprida.

**Why this priority**: É a lacuna concreta entre o que as Skills prometem e o que o binário determinístico realmente verifica. Sem isso, um requisito pode ficar sem nenhuma tarefa associada, ou uma tarefa pode referenciar um requisito que nunca existiu, e nada detecta isso — quebrando a rastreabilidade que o SDD existe para garantir.

**Independent Test**: Criar uma Spec com dois requisitos (`R1`, `R2`) e um `tasks.md` onde uma tarefa declara `Serves: SPEC-001:R1` e nenhuma tarefa referencia `R2`; rodar a validação e confirmar que `R2` é reportado como requisito sem cobertura. Depois, adicionar uma tarefa com `Serves: SPEC-001:R9` (inexistente) e confirmar que essa referência é reportada como quebrada.

**Acceptance Scenarios**:

1. **Given** uma Spec com requisitos `R1` e `R2` e um `tasks.md` cuja única tarefa declara `Serves: SPEC-001:R1`, **When** a validação estrutural é executada, **Then** o sistema reporta `R2` como requisito sem nenhuma tarefa que o cubra.
2. **Given** a mesma Spec, **When** uma tarefa declara `Serves: SPEC-001:R9` e `R9` não existe na Spec `SPEC-001`, **Then** o sistema reporta essa referência como apontando para um requisito inexistente, identificando a tarefa e o requisito referenciado.
3. **Given** uma tarefa cujo `tasks.md` pertence à `SPEC-001` mas declara `Serves: SPEC-002:R1` (um requisito de outra Spec), **When** a validação é executada, **Then** essa referência é reportada como inválida — cobertura cruzada entre Specs não é aceita como cobertura válida.
4. **Given** uma tarefa sem nenhuma declaração `Serves`, **When** a validação é executada, **Then** o sistema reporta essa tarefa como não vinculada a nenhum requisito.
5. **Given** uma Spec onde todo requisito tem ao menos uma tarefa cobrindo-o e toda tarefa referencia um requisito existente da própria Spec, **When** a validação é executada, **Then** nenhum finding de cobertura é reportado.

---

### User Story 2 - Detectar ciclos no grafo de dependências entre Specs (Priority: P1)

Quem executa a validação precisa saber se as dependências declaradas entre Specs (`depends_on`) formam um ciclo — por exemplo, `SPEC-001` depende de `SPEC-002`, que depende de `SPEC-003`, que depende novamente de `SPEC-001`. Hoje a validação apenas confirma que cada entrada de `depends_on` resolve a uma Spec existente; nenhuma verificação constrói o grafo completo nem detecta ciclos.

**Why this priority**: Um ciclo de dependências é uma contradição estrutural — nenhuma das Specs envolvidas pode legitimamente ser implementada primeiro. Detectar isso é tão fundamental quanto detectar um requisito inexistente, e a ausência dessa verificação é uma lacuna conhecida do validador atual.

**Independent Test**: Criar três Specs em ciclo (`A depends_on B`, `B depends_on C`, `C depends_on A`); rodar a validação e confirmar que o ciclo é reportado com o caminho completo (A → B → C → A).

**Acceptance Scenarios**:

1. **Given** três Specs formando um ciclo de dependências, **When** a validação estrutural do projeto é executada, **Then** o sistema reporta um ciclo, identificando cada Spec envolvida e a ordem em que o ciclo se fecha.
2. **Given** uma Spec cujo `depends_on` a referencia a si mesma, **When** a validação é executada, **Then** o sistema reporta essa autorreferência como um ciclo degenerado, distinto de uma dependência ausente.
3. **Given** um grafo de dependências sem nenhum ciclo, por maior e mais ramificado que seja, **When** a validação é executada, **Then** nenhum finding de ciclo é reportado.
4. **Given** um ciclo existente em qualquer parte do grafo do projeto, **When** apenas uma das Specs fora do ciclo é validada isoladamente (`validate SPEC-###`), **Then** essa Spec isolada não é afetada, mas validar qualquer Spec que participa do ciclo reporta o mesmo ciclo que a validação do projeto inteiro reportaria.

---

### User Story 3 - Aplicar regras de validação sensíveis à fase da Spec (Priority: P2)

Quem executa a validação precisa que a ausência de tarefas ou de cobertura de requisitos seja tolerada enquanto a Spec ainda está em rascunho, mas bloqueie a Spec de ser considerada pronta para implementação a partir do momento em que ela avança de fase. Hoje o campo `status` de uma Spec é apenas checado contra uma lista fixa de valores permitidos — nenhuma regra liga o estado declarado à completude real dos artefatos.

**Why this priority**: Sem essa distinção de fase, ou a validação é permissiva demais (deixando uma Spec avançar sem cobertura) ou rígida demais (impedindo o rascunho inicial de uma Spec antes que ela tenha tarefas). É uma melhoria de precisão sobre as User Stories 1 e 2, por isso depende delas existirem primeiro.

**Independent Test**: Criar uma Spec em `draft` sem nenhuma tarefa e confirmar que a validação não bloqueia por ausência de cobertura; avançar essa mesma Spec para `status: ready` sem tarefas e confirmar que a validação agora reporta um problema bloqueante.

**Acceptance Scenarios**:

1. **Given** uma Spec com `status: draft` e nenhuma tarefa ainda criada, **When** a validação é executada, **Then** a ausência de tarefas e de cobertura de requisitos não é reportada como um problema bloqueante.
2. **Given** a mesma Spec com `status: ready` (ou qualquer estado posterior a `draft`) ainda sem tarefas cobrindo seus requisitos, **When** a validação é executada, **Then** o sistema reporta isso como um problema que impede a Spec de ser considerada pronta para implementação.
3. **Given** uma Spec com `status: ready` e todos os seus requisitos cobertos por tarefas, **When** a validação é executada, **Then** nenhum finding de fase é reportado.

---

### Edge Cases

- Um requisito declarado duas vezes com o mesmo número (`R1` repetido) dentro da mesma Spec.
- Uma tarefa com múltiplas declarações `Serves`, cobrindo mais de um requisito ao mesmo tempo.
- Um requisito coberto por mais de uma tarefa (não é um erro — cobertura redundante é válida).
- Uma declaração `Serves` malformada (sintaxe inválida, faltando o número do requisito, ou faltando o ID da Spec).
- Uma Spec sem nenhum `tasks.md` ainda criado, em qualquer fase — deve ser tratado de forma consistente com a ausência de tarefas, não como um erro de leitura de arquivo.
- Um ciclo de dependências que envolve uma Spec com `status: cancelled` ou `superseded` — o ciclo ainda deve ser reportado, já que a estrutura declarada continua contraditória independentemente do estado.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST reconhecer uma sintaxe formal para identificadores de requisito (`R#`, por exemplo `R1`, `R2`) declarados no corpo de uma Spec, únicos dentro dessa Spec.
- **FR-002**: O sistema MUST detectar dois requisitos com o mesmo número declarados dentro da mesma Spec como uma duplicidade.
- **FR-003**: O sistema MUST reconhecer uma sintaxe formal para a referência de cobertura de uma tarefa (`Serves: SPEC-###:R#`, podendo haver mais de uma referência por tarefa) declarada no corpo de uma tarefa em `tasks.md`.
- **FR-004**: O sistema MUST detectar uma referência `Serves` que aponta para um requisito inexistente (número que não existe na Spec referenciada), identificando a tarefa e a referência quebrada.
- **FR-005**: O sistema MUST rejeitar como inválida uma referência `Serves` que aponta para uma Spec diferente da Spec proprietária da tarefa — cobertura cruzada entre Specs não é uma cobertura válida.
- **FR-006**: O sistema MUST detectar uma tarefa sem nenhuma declaração `Serves`, reportando-a como não vinculada a nenhum requisito.
- **FR-007**: O sistema MUST detectar um requisito declarado em uma Spec que não é referenciado por nenhuma tarefa `Serves` daquela Spec, reportando-o como requisito sem cobertura.
- **FR-008**: O sistema MUST construir o grafo de dependências entre Specs a partir do campo `depends_on` de todas as Specs do projeto.
- **FR-009**: O sistema MUST detectar todo ciclo existente nesse grafo, incluindo autorreferência (uma Spec que depende de si mesma), reportando o caminho completo das Specs envolvidas no ciclo.
- **FR-010**: O sistema MUST NOT interromper a análise do restante do projeto ao encontrar um ciclo — outras Specs e outros problemas continuam sendo reportados normalmente.
- **FR-011**: O sistema MUST tolerar a ausência de tarefas e de cobertura de requisitos como não bloqueante enquanto o `status` declarado da Spec for `draft`.
- **FR-012**: O sistema MUST reportar a ausência de tarefas ou de cobertura de requisitos como um problema bloqueante para a prontidão de implementação quando o `status` declarado da Spec for qualquer estado posterior a `draft`.
- **FR-013**: A validação de uma única Spec isolada (`validate SPEC-###`) MUST incluir os mesmos resultados de cobertura de requisitos, dependências e regras de fase para essa Spec e para seus artefatos subordinados (`tasks.md`, `plan.md`) que a validação do projeto inteiro produziria para ela.
- **FR-014**: Todo novo problema detectado por esta funcionalidade MUST ser reportado com um código estável, severidade e localização (arquivo e, quando aplicável, a tarefa ou requisito específico), seguindo o mesmo formato dos demais achados estruturais já existentes.
- **FR-015**: A verificação de existência e cobertura declarada de um requisito MUST NOT ser apresentada como prova de que o requisito foi satisfeito semanticamente pela implementação — apenas que existe uma referência declarada entre eles.

### Key Entities

- **Requirement (Requisito)**: unidade de requisito numerada (`R#`) declarada no corpo de uma Spec; sua identidade completa (`SPEC-###:R#`) só existe em conjunto com a Spec proprietária, no mesmo espírito da identidade composta de Tarefa já estabelecida.
- **Coverage Reference (Referência de cobertura / `Serves`)**: declaração dentro de uma tarefa, em seu `tasks.md`, que aponta para um ou mais Requirements que essa tarefa atende. Só é válida quando aponta para um Requirement da mesma Spec proprietária da tarefa.
- **Dependency Graph (Grafo de dependências)**: estrutura derivada dos campos `depends_on` de todas as Specs do projeto, usada para detectar ciclos; não é persistida — reconstruída a cada validação a partir do próprio sistema de arquivos.
- **Spec Phase Gate (Regra de fase)**: regra que associa o `status` declarado de uma Spec à obrigatoriedade de cobertura completa de requisitos — tolerante em `draft`, bloqueante a partir de qualquer estado posterior.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Um ciclo entre Specs é reportado com o caminho completo das Specs envolvidas em 100% dos casos testados, sem interromper a verificação do restante do projeto.
- **SC-002**: Uma referência a um requisito inexistente, ou a um requisito de outra Spec, é detectada em toda execução da validação, tanto no projeto inteiro quanto em uma Spec isolada.
- **SC-003**: Uma Spec em rascunho sem tarefas nunca é bloqueada por ausência de cobertura; toda Spec além do rascunho sem cobertura completa é bloqueada em 100% dos casos testados.
- **SC-004**: Validar uma Spec isoladamente e validar o projeto inteiro produzem exatamente os mesmos achados de cobertura e dependência para essa Spec, em todo caso testado.
- **SC-005**: Specs existentes, já corretamente cobertas e sem ciclos, continuam validando sem nenhuma reescrita manual após a adoção desta funcionalidade.

## Assumptions

- A sintaxe de requisito adotada é a convenção já documentada nas Skills (`R#` no corpo da Spec, `Serves: SPEC-###:R#` no corpo da tarefa) — não a convenção `FR-###` do template genérico de especificação do próprio spec-kit, que é um nível diferente (descreve a especificação em si, não os requisitos de um projeto gerenciado pelo misterspec).
- Uma tarefa só pode declarar cobertura válida de requisitos pertencentes à sua própria Spec proprietária; uma referência para outra Spec é tratada como inválida, não como uma forma de cobertura cruzada entre Specs.
- A obrigatoriedade de cobertura completa se aplica a partir de qualquer `status` posterior a `draft` (ou seja, `ready` em diante); apenas `draft` é tolerante à ausência de tarefas e cobertura.
- Esta funcionalidade não introduz nenhum estado persistido novo — o grafo de dependências e a cobertura de requisitos são recalculados a partir do sistema de arquivos a cada validação, preservando o Princípio III (sistema de arquivos como fonte única de verdade).
- Esta especificação depende do modelo de identidade canônica de tarefas já estabelecido (031-canonical-task-identity): uma referência `Serves` dentro de uma tarefa é resolvida no contexto da Spec proprietária dessa tarefa, do mesmo modo que a resolução de `TASK-NNN` já é escopada por Spec.
