# Feature Specification: Recuperação orientada à tarefa e preparação de execução

**Feature Branch**: `034-task-oriented-context-preparation`
**Created**: 2026-09-18
**Status**: Draft
**Input**: User description: "PROP-04 — Recuperação orientada à tarefa e preparação de execução: selecionar contexto a partir da tarefa executável atual, reunindo informações mecânicas que hoje exigem várias chamadas e interpretação da Skill."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Preparar a execução de uma tarefa em uma única resposta (Priority: P1)

Um agente que vai implementar uma Tarefa específica precisa, hoje, encadear várias chamadas mecânicas e interpretar prosa livre para reunir tudo o que precisa: `internal resolve`, `internal inspect`, `internal context`, e depois ler manualmente o corpo da Tarefa para descobrir de que ela depende, qual seu escopo de arquivos/componentes e como será verificada — exatamente o que `mister-implement` hoje documenta como três passos manuais separados. Uma operação de preparação deve reunir tudo isso — a Tarefa, seu contexto (título, requisitos que atende, escopo, e as seções do Plano associadas a ela), suas restrições, e seu método de verificação — em uma única resposta determinística, sem executar nenhum comando.

**Why this priority**: É o valor central da proposta — sem isso, a preparação de uma tarefa continua exigindo a mesma sequência de chamadas e a mesma interpretação de prosa que já existe hoje, e a proposta não entrega nada de novo.

**Independent Test**: Pedir a preparação de uma Tarefa específica (`SPEC-###:TASK-###`) e confirmar que a resposta já contém a Tarefa, seu contexto, suas restrições e seu método de verificação — sem precisar de nenhuma chamada adicional a `resolve`, `inspect` ou `context`.

**Acceptance Scenarios**:

1. **Given** uma Tarefa existente e sem dependências pendentes, **When** um agente pede sua preparação, **Then** a resposta inclui a Tarefa identificada, o contexto reunido para ela (título, requisitos que atende, escopo e seções do Plano associadas), suas restrições e seu método de verificação, em uma única chamada.
2. **Given** duas Tarefas diferentes da mesma Spec, **When** cada uma é preparada separadamente, **Then** cada resposta contém contexto genuinamente específico daquela Tarefa (requisitos, escopo e seções do Plano diferentes quando as Tarefas de fato diferem), nunca o mesmo pacote de contexto da Spec inteira repetido para as duas.
3. **Given** uma preparação bem-sucedida, **When** a resposta é inspecionada, **Then** nenhum comando foi executado como parte da preparação — ela é estritamente de leitura.

---

### User Story 2 - Nunca apresentar uma tarefa bloqueada como pronta (Priority: P1)

Hoje, se uma Tarefa depende de outra Tarefa ainda não concluída, essa dependência existe apenas como prosa livre no corpo da Tarefa (sem sintaxe formal, sem verificação automática) e `mister-implement` precisa interpretá-la manualmente para decidir se pode prosseguir. A preparação de uma tarefa deve verificar, de forma determinística, se todas as Tarefas das quais ela depende (dentro da mesma Spec) já estão concluídas, e recusar-se a apresentá-la como pronta quando não estiverem — listando explicitamente quais dependências ainda faltam.

**Why this priority**: Apresentar uma tarefa bloqueada como pronta leva a trabalho fora de ordem e retrabalho — é tão fundamental quanto a User Story 1, por isso tem a mesma prioridade.

**Independent Test**: Criar uma Tarefa que declara depender de outra Tarefa ainda incompleta; pedir sua preparação e confirmar que a resposta identifica a Tarefa como bloqueada, nomeando exatamente qual dependência falta, em vez de devolver um contexto de execução como se ela estivesse pronta.

**Acceptance Scenarios**:

1. **Given** uma Tarefa que declara depender de outra Tarefa ainda não concluída, **When** sua preparação é pedida, **Then** a resposta identifica a Tarefa como bloqueada e nomeia especificamente qual dependência está pendente, sem apresentar contexto de execução como se ela estivesse pronta.
2. **Given** a mesma Tarefa, **When** a Tarefa da qual ela depende é marcada como concluída e a preparação é pedida novamente, **Then** a Tarefa agora é apresentada como pronta, com seu contexto completo.
3. **Given** uma Tarefa sem nenhuma dependência declarada, **When** sua preparação é pedida, **Then** ela é tratada como pronta por padrão — a ausência de uma declaração de dependência não é um bloqueio.
4. **Given** duas Tarefas que declaram depender uma da outra (um ciclo), **When** a preparação de qualquer uma delas é pedida, **Then** o sistema reporta o ciclo como um diagnóstico claro, em vez de travar ou entrar em laço infinito.

---

### User Story 3 - Selecionar automaticamente a próxima tarefa pronta (Priority: P2)

Quando o agente não especifica qual Tarefa quer preparar, o sistema deve selecionar automaticamente a próxima Tarefa pronta (sem dependências pendentes) dentro da Spec indicada, com um critério de desempate estável e determinístico quando mais de uma Tarefa está igualmente pronta. Depois que uma Tarefa é concluída, a próxima preparação deve recalcular a prontidão a partir do estado atual — nunca de um resultado armazenado de uma chamada anterior.

**Why this priority**: É uma melhoria de conveniência sobre a seleção explícita das User Stories 1 e 2 — só faz sentido depois que a lógica de prontidão (US2) já existe. Prioridade menor porque a seleção explícita por ID já cobre o caso de uso essencial.

**Independent Test**: Em uma Spec com três Tarefas, das quais uma está bloqueada e duas estão prontas, pedir a preparação sem especificar uma Tarefa e confirmar que a Tarefa pronta de menor número é selecionada. Concluir essa Tarefa e pedir novamente, sem especificar; confirmar que a próxima Tarefa pronta é selecionada agora, refletindo o novo estado.

**Acceptance Scenarios**:

1. **Given** uma Spec com múltiplas Tarefas prontas, **When** a preparação é pedida sem especificar uma Tarefa, **Then** a Tarefa pronta de menor número (`TASK-NNN`) é selecionada, de forma determinística e repetível.
2. **Given** a mesma Spec, **When** a Tarefa selecionada é concluída e a preparação é pedida novamente sem especificar uma Tarefa, **Then** a seleção automática reflete o novo estado — nunca reutiliza a resposta anterior.
3. **Given** uma Spec em que nenhuma Tarefa está pronta (todas concluídas ou todas bloqueadas), **When** a preparação é pedida sem especificar uma Tarefa, **Then** o sistema informa claramente que não há nenhuma Tarefa pronta, em vez de retornar uma seleção arbitrária ou vazia sem explicação.

---

### Edge Cases

- Uma Tarefa declara depender de um número de Tarefa que não existe naquela Spec.
- Uma Tarefa declara depender de si mesma.
- Duas ou mais Tarefas declaram dependência mútua, formando um ciclo maior que um par.
- Uma Tarefa referencia, no `Depends on`, uma Tarefa de outra Spec — não é um caso válido, já que a identidade de Tarefa é local à sua própria Spec (031-canonical-task-identity).
- Nenhuma seção do Plano está associada a uma Tarefa específica (a Tarefa não serve nenhum requisito coberto pelo Plano) — a preparação ainda deve funcionar, apenas sem seções de Plano na resposta.
- A Spec indicada não tem nenhum `tasks.md` ainda — a preparação deve informar isso claramente, não falhar de forma confusa.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST oferecer uma operação de preparação que, para uma Tarefa identificada (`SPEC-###:TASK-###` ou `TASK-###` já no contexto de uma Spec), retorna em uma única resposta: a identidade da Tarefa, seu contexto reunido, suas restrições e seu método de verificação.
- **FR-002**: A operação de preparação MUST NOT executar nenhum comando, script, ou teste como parte de sua própria execução — ela é estritamente de leitura e montagem de informação já existente.
- **FR-003**: O sistema MUST reconhecer uma sintaxe formal para a declaração de dependência de uma Tarefa em relação a outra Tarefa da mesma Spec (por exemplo, um campo `Depends on:` no corpo da Tarefa, análogo ao já estabelecido para `Serves:`), aceitando também a ausência dessa declaração como "sem dependências".
- **FR-004**: O sistema MUST reconhecer uma sintaxe formal para o método de verificação de uma Tarefa (por exemplo, um campo `Verify:`), permitindo que a preparação o retorne como um campo estruturado, não como prosa a ser interpretada.
- **FR-005**: O sistema MUST reconhecer uma sintaxe formal para o escopo de arquivos/componentes de uma Tarefa (por exemplo, um campo `Scope:`), com o mesmo tratamento.
- **FR-006**: Antes de apresentar uma Tarefa como pronta, o sistema MUST verificar se toda Tarefa da qual ela depende (dentro da mesma Spec) já está concluída.
- **FR-007**: Quando uma ou mais dependências não estiverem concluídas, o sistema MUST reportar a Tarefa como bloqueada e nomear especificamente quais dependências estão pendentes, em vez de retornar um contexto de execução como se ela estivesse pronta.
- **FR-008**: O sistema MUST detectar um ciclo de dependências entre Tarefas da mesma Spec e reportá-lo como um diagnóstico claro, sem travar ou processar indefinidamente.
- **FR-009**: O sistema MUST rejeitar uma declaração de dependência que aponte para uma Tarefa de outra Spec, ou para um número de Tarefa inexistente na mesma Spec, como uma declaração inválida.
- **FR-010**: O contexto reunido para uma Tarefa MUST ser específico daquela Tarefa — reunindo os requisitos que ela serve (já estabelecidos por `Serves:`), seu próprio escopo declarado, e as seções do Plano cujo conteúdo se relaciona aos mesmos requisitos que a Tarefa serve — nunca o mesmo pacote de contexto da Spec inteira, indiferenciado entre Tarefas.
- **FR-011**: Quando nenhuma Tarefa é especificada, o sistema MUST selecionar automaticamente a Tarefa pronta de menor número dentro da Spec indicada, de forma determinística e repetível diante do mesmo estado.
- **FR-012**: Quando nenhuma Tarefa da Spec está pronta (todas concluídas ou todas bloqueadas), o sistema MUST informar isso explicitamente, em vez de retornar uma seleção vazia ou arbitrária sem explicação.
- **FR-013**: A prontidão de uma Tarefa MUST ser recalculada a partir do estado atual do projeto em toda chamada de preparação — nunca reutilizando um resultado de uma chamada anterior.

### Key Entities

- **Task Dependency (Dependência de Tarefa)**: uma declaração formal, dentro do corpo de uma Tarefa, de que ela depende de outra Tarefa da mesma Spec — só é válida quando aponta para uma Tarefa existente na mesma Spec proprietária.
- **Task Readiness (Prontidão de Tarefa)**: o estado computado de uma Tarefa — pronta, quando todas as suas dependências estão concluídas, ou bloqueada, com a lista explícita de dependências pendentes.
- **Task Preparation (Preparação de Tarefa)**: o pacote de resposta único que reúne a identidade da Tarefa, seu contexto (título, requisitos servidos, escopo, seções do Plano associadas), suas restrições e seu método de verificação.
- **Plan Section Association (Associação de seção do Plano)**: a relação, determinística e derivada dos requisitos que uma Tarefa serve, entre uma Tarefa e as seções do Plano cujo conteúdo trata dos mesmos requisitos.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Um agente que prepara uma Tarefa específica obtém tudo o que precisa para começar a implementá-la em uma única chamada, sem nenhuma chamada adicional de resolução, inspeção ou contexto, em 100% dos casos testados.
- **SC-002**: Uma Tarefa bloqueada nunca é apresentada como pronta — verificado em 100% dos casos testados, incluindo ciclos de dependência.
- **SC-003**: Duas Tarefas diferentes da mesma Spec recebem pacotes de contexto genuinamente distintos sempre que seus requisitos, escopo, ou seções de Plano associadas realmente diferem.
- **SC-004**: Depois de concluir uma Tarefa, a próxima preparação sem Tarefa especificada reflete o novo estado de prontidão em 100% dos casos testados, nunca reutilizando uma resposta anterior.
- **SC-005**: Uma Spec sem nenhuma Tarefa pronta produz uma mensagem clara e explícita, nunca uma resposta vazia sem explicação, em 100% dos casos testados.

## Assumptions

- Esta especificação depende diretamente de três funcionalidades já implementadas: a identidade composta de Tarefas (031-canonical-task-identity), a cobertura de requisitos e o campo `Serves:` (032-requirement-coverage-dependency-validation), e o Context Pack com conteúdo completo e localização absoluta (033-context-pack-output-contract).
- A recuperação de contexto hoje rejeita explicitamente uma Tarefa como alvo direto (ela não é um dos cinco tipos de entidade referenciáveis) — esta especificação assume que a preparação de uma Tarefa passa a reunir contexto especificamente para ela, e não que a Tarefa se torna um alvo genérico de toda operação de contexto existente.
- Uma Tarefa autorada antes desta funcionalidade existir, sem nenhuma declaração `Depends on:`, `Verify:` ou `Scope:`, continua válida — a ausência dessas declarações é tratada como "sem dependências" / "sem informação estruturada disponível", nunca como um erro.
- A associação entre uma Tarefa e as seções do Plano é derivada dos requisitos que a Tarefa serve (via `Serves:`) cruzados com o conteúdo do Plano — não exige que o Plano declare formalmente a que Tarefa cada seção pertence.
- A operação de preparação não introduz nenhum estado novo persistido — a prontidão e o contexto são recalculados a cada chamada a partir do sistema de arquivos, preservando o Princípio III (sistema de arquivos como fonte única de verdade).
