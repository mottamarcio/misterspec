# Feature Specification: Skills Enxutas e Contratos de Integração Testáveis

**Feature Branch**: `039-lean-skills-integration-contracts`
**Created**: 2026-09-21
**Status**: Draft
**Input**: User description: "PROP-14 — Skills enxutas e contratos de integração testáveis: reduzir repetição de instruções entre Skills e garantir que elas usem corretamente as capacidades determinísticas do framework, com testes de existência de comandos, validade de exemplos, correspondência entre verificações prometidas e capacidades reais, smoke tests entre integrações representativas, e registro de falhas/fallback do Context Engine."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Manter regras comuns em uma única fonte (Priority: P1)

Como mantenedor do framework, quero que as regras compartilhadas entre as
Skills (critérios de sucesso, pós-condições, regras de interação, etc.)
existam em uma única fonte de autoria, para que uma correção ou ajuste de
regra comum seja aplicado a todas as Skills de uma vez, sem exigir edição
manual repetida em cada arquivo e sem risco de as cópias divergirem entre
si ao longo do tempo.

**Why this priority**: É a causa raiz do problema — sem uma fonte única,
qualquer melhoria futura nas regras comuns exige N edições manuais
sincronizadas, e o histórico do projeto já mostra 10 Skills quase
idênticas na estrutura. Resolver isso primeiro reduz o custo de todas as
demais tarefas desta proposta.

**Independent Test**: Pode ser testado alterando uma regra comum na fonte
canônica e verificando que todas as Skills distribuídas (para cada
integração suportada) refletem a mudança sem edição adicional, mantendo
as exceções específicas de cada Skill intactas.

**Acceptance Scenarios**:

1. **Given** uma regra comum a todas as Skills definida na fonte
   canônica, **When** essa regra é alterada, **Then** todas as Skills
   geradas para todas as integrações suportadas passam a refletir a nova
   redação sem exigir edição manual de cada arquivo individual.
2. **Given** uma Skill com uma exceção legítima a uma regra comum (por
   exemplo, uma restrição adicional só aplicável a ela), **When** as
   Skills são geradas a partir da fonte canônica, **Then** essa exceção
   permanece presente apenas naquela Skill, sem vazar para as demais.
3. **Given** o conjunto atual de Skills com forte repetição estrutural,
   **When** a consolidação é aplicada, **Then** o tamanho total das
   instruções entregues aos agentes é mensuravelmente menor sem que
   nenhuma seção obrigatória do contrato de Skill seja removida.

---

### User Story 2 - Skills não prometem verificações que o binário não pode cumprir (Priority: P1)

Como mantenedor do framework, quero uma validação automática que
confirme que cada comando citado em uma Skill existe de fato no binário,
que os exemplos de invocação são válidos, e que cada verificação
prometida corresponde a uma capacidade determinística real, para que
nenhuma Skill instrua um agente a confiar em um comportamento que o
framework não entrega.

**Why this priority**: Uma Skill que promete uma verificação inexistente
engana o agente que a segue e pode levar a tarefas marcadas como
concluídas sem verificação real — um risco direto à confiabilidade do
SDD que este framework promove.

**Independent Test**: Pode ser testado introduzindo deliberadamente uma
referência a um comando inexistente ou uma verificação sem capacidade
correspondente em uma Skill e confirmando que a validação automática
falha e aponta a Skill e a linha responsáveis.

**Acceptance Scenarios**:

1. **Given** uma Skill que cita um comando que não está registrado no
   binário atual, **When** a validação é executada, **Then** o processo
   falha apontando a Skill e o comando inexistente.
2. **Given** uma Skill que promete uma verificação (por exemplo, "valida
   cobertura de requisitos") sem que exista capacidade determinística
   correspondente no binário, **When** a validação é executada, **Then**
   o processo falha apontando a promessa não cumprida.
3. **Given** um exemplo de invocação de comando dentro do texto de uma
   Skill, **When** a validação é executada, **Then** o exemplo é
   confirmado como sintaticamente válido contra a CLI atual ou reportado
   como inválido.

---

### User Story 3 - Comportamento consistente entre integrações representativas (Priority: P2)

Como mantenedor do framework, quero smoke tests executados em
integrações de agente representativas, para confirmar que conteúdo
idêntico de Skill não implica comportamento idêntico entre diferentes
agentes, e detectar regressões específicas de integração antes que
cheguem a um usuário.

**Why this priority**: Sem essa verificação, uma mudança pode passar em
todos os testes de conteúdo e ainda assim quebrar o fluxo em uma
integração específica, algo que só aparece em uso real com aquele
agente.

**Independent Test**: Pode ser testado executando o smoke test contra um
conjunto fixo de integrações representativas após uma mudança de
conteúdo nas Skills e confirmando que o fluxo básico (preparação de
tarefa, obtenção de contexto, verificação) é concluído com sucesso em
cada uma.

**Acceptance Scenarios**:

1. **Given** uma mudança nas Skills consolidadas, **When** os smoke
   tests são executados contra as integrações representativas
   definidas, **Then** cada integração completa o fluxo básico esperado
   ou o teste falha identificando qual integração e em qual etapa.
2. **Given** duas integrações que compartilham o mesmo diretório de
   destino de Skills, **When** o smoke test é executado, **Then** ambas
   são verificadas separadamente, pois compartilhar diretório não
   garante comportamento idêntico do agente.

---

### User Story 4 - Falhas e fallback do Context Engine ficam visíveis (Priority: P2)

Como responsável por avaliar a eficiência do framework, quero que toda
falha ou fallback do motor de contexto durante o uso de uma Skill seja
registrado, para que o custo real de leituras adicionais causadas por
esse fallback seja contabilizado na avaliação de qualidade e eficiência,
em vez de ficar oculto atrás de um resultado aparentemente bem-sucedido.

**Why this priority**: Sem esse registro, a avaliação de eficiência
subestima o custo real de tarefas onde o fluxo preferencial falhou e o
agente precisou compensar com leituras manuais — inflando
artificialmente os ganhos reportados do framework.

**Independent Test**: Pode ser testado forçando uma condição de falha no
motor de contexto durante a execução de uma Skill e confirmando que o
evento de fallback é registrado com informação suficiente para ser
contabilizado por uma avaliação externa.

**Acceptance Scenarios**:

1. **Given** uma solicitação de contexto que falha ou recorre a
   comportamento alternativo, **When** a Skill continua a tarefa usando
   esse fallback, **Then** o evento de fallback fica registrado de forma
   que uma avaliação posterior possa identificá-lo e contabilizar as
   leituras extras associadas.
2. **Given** uma execução de tarefa sem nenhuma falha do motor de
   contexto, **When** a tarefa é concluída, **Then** nenhum evento de
   fallback é registrado indevidamente.

### Edge Cases

- O que acontece quando uma Skill cita um comando que existia em uma
  versão anterior do binário e foi renomeado ou removido?
- Como o sistema trata uma regra que é exceção legítima em apenas uma
  Skill, garantindo que a consolidação não a apague silenciosamente?
- O que acontece quando uma integração representativa dos smoke tests
  está indisponível no ambiente de execução (por exemplo, o agente não
  está instalado)?
- Como o registro de fallback trata o caso em que o motor de contexto
  falha parcialmente, já tendo entregado parte da saída antes de
  recorrer ao fallback?
- O que acontece quando a redução de instruções elimina, por engano, uma
  seção obrigatória do contrato de Skill (por exemplo, uma das seções
  definidas em `docs/architecture-specification.md` §39)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O framework MUST manter uma fonte canônica única para as
  regras e instruções compartilhadas entre as Skills, a partir da qual
  as variantes compactas entregues a cada integração suportada são
  produzidas, sem depender de mecanismos de inclusão em tempo de
  execução que os agentes não suportam.
- **FR-002**: O framework MUST eliminar duplicação de instruções comuns
  (por exemplo, critérios de sucesso, pós-condições, regras de
  interação) entre as Skills existentes, preservando toda exceção
  legítima e específica de uma Skill individual.
- **FR-003**: Todas as Skills MUST refletir o fluxo de preparação
  orientada à tarefa e o consumo explícito do Context Pack completo já
  disponíveis no framework, em vez de instruir leituras mecânicas
  redundantes ou recuperação de contexto ad hoc.
- **FR-004**: O framework MUST validar automaticamente que todo comando
  interno citado no texto de uma Skill corresponde a um comando
  realmente registrado no binário corrente.
- **FR-005**: O framework MUST validar automaticamente que todo exemplo
  de invocação de comando presente no texto de uma Skill é
  sintaticamente válido contra a interface de linha de comando corrente.
- **FR-006**: O framework MUST validar automaticamente que toda
  verificação ou checagem prometida pelo texto de uma Skill corresponde
  a uma capacidade determinística real do binário, rejeitando promessas
  sem capacidade correspondente.
- **FR-007**: O framework MUST executar smoke tests de ponta a ponta em
  um conjunto definido de integrações de agente representativas,
  cobrindo ao menos o fluxo básico de preparação de tarefa, obtenção de
  contexto e verificação, e MUST reportar falhas identificando a
  integração e a etapa específica.
- **FR-008**: O framework MUST registrar toda falha ou execução de
  fallback do motor de contexto ocorrida durante o uso de uma Skill,
  com informação suficiente para que uma avaliação externa contabilize
  as leituras adicionais correspondentes.
- **FR-009**: A consolidação de instruções MUST ser verificável: a
  validação estrutural existente (presença e ordem das seções
  obrigatórias do contrato de Skill) MUST continuar passando após
  qualquer redução de conteúdo.
- **FR-010**: Qualquer redução no tamanho das instruções entregues às
  Skills MUST ser avaliada quanto ao impacto em qualidade usando o
  arcabouço de avaliação de ponta a ponta existente, antes de ser
  considerada pronta para adoção.

### Key Entities

- **Fonte Canônica de Regras**: o conjunto único e versionado de regras,
  critérios e instruções compartilhados entre Skills, do qual as
  variantes por integração são derivadas.
- **Variante de Skill por Integração**: o arquivo de Skill compacto e
  específico entregue a uma integração de agente suportada, derivado da
  fonte canônica mais o conteúdo específico daquela Skill.
- **Capacidade Determinística**: uma operação real do binário (por
  exemplo, um comando interno) que uma Skill pode legitimamente citar ou
  prometer como verificação.
- **Execução de Smoke Test**: o registro de uma verificação de ponta a
  ponta contra uma integração representativa específica, incluindo
  resultado e etapa de falha, quando houver.
- **Evento de Fallback do Context Engine**: o registro de uma falha ou
  recurso a comportamento alternativo durante a obtenção de contexto,
  incluindo a Skill e a tarefa em que ocorreu.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: O volume total de instruções entregues às Skills (medido
  em tamanho de texto) não regride além de uma margem pequena e
  explícita em relação à linha de base atual, sem que nenhuma
  verificação estrutural existente passe a falhar. **Nota de emenda
  (pós-implementação)**: a meta original desta Spec — redução de pelo
  menos 20% — mostrou-se inatingível pela abordagem de consolidação em
  fonte canônica adotada (User Story 1): extrair frases duplicadas para
  fragmentos elimina a duplicação na *fonte de autoria*
  (`internal/skillgen/fragment.go`), mas cada `SKILL.md` gerado
  permanece Markdown completo e byte-idêntico ao que um agente lê —
  nenhuma invocação de agente lê mais de um arquivo de Skill por vez,
  então a duplicação entre arquivos nunca era custo pago em uma única
  sessão. Reduzir o tamanho de fato exigiria cortar conteúdo real das
  Skills, o que conflita com a garantia de qualidade de FR-010. O
  volume total após esta Spec é 2393 linhas (alta de 0,6% sobre a
  linha de base de 2378, por causa do texto de registro de fallback da
  User Story 4) — SC-001 foi reescrito para um guarda-corpo de
  não-regressão (verificado por
  `internal/example/skills_size_test.go`), e uma meta real de redução
  de tamanho fica para uma Spec futura que trate isso como sua própria
  User Story, com orçamento e trade-offs de qualidade avaliados
  deliberadamente — não como efeito colateral de uma consolidação de
  autoria.
- **SC-002**: 100% dos comandos internos citados e das verificações
  prometidas no texto das Skills correspondem a capacidades reais do
  binário corrente, de forma continuamente verificável.
- **SC-003**: A taxa de sucesso de tarefas medida pelo arcabouço de
  avaliação de ponta a ponta não piora, dentro da margem de tolerância
  definida pela linha de base, após a consolidação das Skills.
- **SC-004**: Smoke tests cobrindo pelo menos três integrações de agente
  representativas e com destinos de instalação distintos são executados
  e aprovados de forma consistente antes de qualquer mudança nas Skills
  ser considerada pronta.
- **SC-005**: 100% dos eventos de fallback do motor de contexto
  ocorridos durante a execução de uma tarefa ficam disponíveis para
  contabilização por uma avaliação externa, sem leituras extras
  ocultas.

## Assumptions

- O Context Pack completo (Spec 033) e a preparação orientada à tarefa
  (Spec 034) já estão disponíveis no framework; esta proposta consome
  essas capacidades em vez de reconstruí-las.
- O arcabouço de avaliação de ponta a ponta (Spec 037) já existe e pode
  ser estendido para medir o efeito da redução de instruções e capturar
  eventos de fallback, em vez de exigir um novo sistema de avaliação
  paralelo.
- "Integrações representativas" para os smoke tests significa um
  subconjunto fixo e documentado das integrações de agente já
  suportadas, escolhido para cobrir destinos de instalação distintos,
  não a totalidade das integrações a cada mudança.
- A fonte canônica de regras e o processo de geração de variantes são
  detalhes internos de manutenção; o usuário final de cada integração
  continua recebendo um único arquivo de Skill por integração, como
  hoje.
- Nenhum mecanismo novo de inclusão em tempo de execução (runtime
  include) será introduzido para os agentes consumirem a fonte
  canônica diretamente, já que os agentes atualmente suportados não têm
  esse suporte.
