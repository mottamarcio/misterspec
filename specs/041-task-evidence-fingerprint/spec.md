# Feature Specification: Evidências de Execução e Validade por Fingerprint

**Feature Branch**: `041-task-evidence-fingerprint`
**Created**: 2026-09-22
**Status**: Draft
**Input**: User description: "PROP-10 — Evidências de execução e validade por fingerprint: vincular a conclusão de uma tarefa a evidências verificáveis e aos insumos contra os quais a verificação foi realizada, distinguindo evidência automática de declarada, registrando revisão Git e alterações locais, guardando logs fora do Context Pack, e tratando mudanças dos insumos como evidência desatualizada — para que uma tarefa não seja considerada verificada apenas por um checkbox marcado."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Concluir uma tarefa exige evidência, não só um checkbox (Priority: P1)

Como implementador seguindo o fluxo de tarefas, quando termino uma
Tarefa quero registrar uma evidência verificável — o que foi executado
ou checado, o resultado, quando e contra qual conteúdo — para que a
alegação de "concluída" seja algo que outra pessoa (ou eu mesmo, mais
tarde) possa checar, em vez de confiar apenas em uma caixa marcada.

**Why this priority**: É o problema central da proposta — hoje uma
Tarefa é tratada como concluída somente por ter sua caixa de seleção
marcada, sem qualquer registro de que uma verificação realmente
ocorreu ou do que ela realmente cobriu.

**Independent Test**: Pode ser testado marcando uma Tarefa como
concluída sem nenhuma evidência associada e confirmando que ela não é
tratada como verificada; em seguida, associando uma evidência que
descreve o método de verificação e seu resultado, e confirmando que a
Tarefa passa a ser tratada como verificada.

**Acceptance Scenarios**:

1. **Given** uma Tarefa cuja caixa de seleção está marcada mas sem
   nenhum registro de evidência associado, **When** o estado de
   conclusão dessa Tarefa é consultado, **Then** ela não é apresentada
   como verificada apenas por causa da caixa marcada.
2. **Given** uma Tarefa com um registro de evidência associado que
   descreve método, resultado e o momento da verificação, **When** o
   estado de conclusão é consultado, **Then** a Tarefa é apresentada
   como verificada, com a evidência acessível a partir dela.
3. **Given** uma tentativa de verificação que falha, **When** essa
   falha é registrada, **Then** a Tarefa não é concluída — a falha fica
   registrada como tal, distinguível de uma verificação nunca tentada.

---

### User Story 2 - Dependências entre tarefas respeitam evidência real, não só a caixa marcada (Priority: P1)

Como pessoa que depende da prontidão calculada de uma Tarefa (por
exemplo, ao preparar a próxima Tarefa a executar), quero que uma
Tarefa dependente só seja considerada satisfeita quando sua própria
evidência de verificação for válida, para que eu nunca construa
trabalho sobre uma dependência marcada como feita mas nunca realmente
verificada.

**Why this priority**: Sem isso, o próprio mecanismo de prontidão do
framework herda a fragilidade do checkbox isolado — uma Tarefa
upstream marcada sem verificação real ainda liberaria o trabalho
seguinte como se estivesse pronta.

**Independent Test**: Pode ser testado criando uma Tarefa upstream
marcada como concluída sem evidência válida e confirmando que uma
Tarefa que depende dela não é apresentada como pronta para execução;
em seguida, associando evidência válida à Tarefa upstream e confirmando
que a dependente passa a ser apresentada como pronta.

**Acceptance Scenarios**:

1. **Given** uma Tarefa upstream marcada como concluída mas sem
   evidência válida, **When** a prontidão de uma Tarefa que depende
   dela é calculada, **Then** a dependente não é apresentada como
   pronta.
2. **Given** a mesma Tarefa upstream agora com evidência válida
   associada, **When** a prontidão é recalculada, **Then** a
   dependente passa a ser apresentada como pronta.

---

### User Story 3 - Mudar os insumos verificados torna a evidência antiga visivelmente desatualizada (Priority: P2)

Como mantenedor, quando o conteúdo contra o qual uma Tarefa foi
verificada muda depois da verificação (por exemplo, o escopo ou o
requisito da própria Tarefa é editado), quero que a evidência antiga
seja sinalizada como desatualizada, para que uma mudança não fique
silenciosamente coberta por uma verificação que já não corresponde ao
que existe agora.

**Why this priority**: Sem essa sinalização, uma edição legítima em uma
Tarefa já "concluída" continuaria parecendo totalmente verificada,
escondendo exatamente o tipo de divergência que motiva revisão.

**Independent Test**: Pode ser testado capturando evidência para uma
Tarefa, editando o conteúdo contra o qual ela foi verificada, e
confirmando que a evidência passa a ser sinalizada como desatualizada
sem exigir nenhuma ação manual de detecção.

**Acceptance Scenarios**:

1. **Given** uma Tarefa com evidência válida capturada contra seu
   conteúdo atual, **When** esse conteúdo é editado depois da captura,
   **Then** a evidência passa a ser sinalizada como desatualizada, não
   mais apresentada como verificação corrente.
2. **Given** uma Tarefa cuja evidência está marcada como desatualizada,
   **When** uma nova verificação é capturada contra o conteúdo atual,
   **Then** a Tarefa volta a ser apresentada como verificada, sem
   registro conflitante da verificação antiga.

---

### User Story 4 - Execução automatizada de verificação é explícita, controlada e registra o estado real do repositório (Priority: P2)

Como pessoa responsável por revisar o que uma Tarefa realmente fez,
quando a evidência é capturada por execução automatizada de um comando
de verificação, quero que essa execução seja explícita (comando,
argumentos, diretório, tempo limite), respeite as permissões já
existentes do agente e do projeto, e registre a revisão real do Git
incluindo alterações locais não commitadas, para que a evidência nunca
misrepresente o que de fato foi executado ou contra qual estado do
código.

**Why this priority**: Sem controles explícitos, a captura automática
de evidência poderia executar texto arbitrário encontrado em Markdown,
ignorar permissões já estabelecidas, ou registrar apenas um hash de
commit que não reflete alterações locais reais — comprometendo
exatamente a confiabilidade que a evidência deveria trazer.

**Independent Test**: Pode ser testado solicitando captura automatizada
de evidência com comando, argumentos, diretório e tempo limite
explícitos, e confirmando que apenas esse comando exato é executado;
em seguida, alterando um arquivo localmente sem commitar e confirmando
que a evidência registra a árvore de trabalho como modificada, não
apenas o SHA do commit.

**Acceptance Scenarios**:

1. **Given** uma solicitação de captura automatizada de evidência com
   comando, argumentos, diretório e tempo limite explícitos, **When**
   a captura é executada, **Then** apenas esse comando exato roda,
   dentro do diretório e do tempo limite declarados — nenhum texto de
   Markdown é executado implicitamente.
2. **Given** um comando de verificação que excede as permissões já
   concedidas ao agente ou ao projeto, **When** a captura automatizada
   é solicitada, **Then** ela é recusada, sem burlar essa permissão.
3. **Given** alterações locais não commitadas relevantes ao escopo
   verificado, **When** a evidência é capturada, **Then** o registro
   inclui que a árvore de trabalho estava modificada, não apenas o SHA
   do commit atual.

### Edge Cases

- O que acontece quando uma Tarefa é desmarcada de concluída para
  pendente — a evidência anterior é descartada ou mantida como registro
  histórico?
- Como o sistema trata evidência declarada/importada por uma pessoa
  (sem execução automatizada real), em termos de nível de confiança
  apresentado?
- O que acontece quando o comando de verificação automatizada excede o
  tempo limite declarado?
- Como o sistema trata uma execução automatizada de verificação bem-
  sucedida, mas cujo comando na verdade não exercita o requisito que a
  Tarefa deveria satisfazer?
- O que acontece quando não é possível calcular o fingerprint dos
  insumos (por exemplo, o arquivo referenciado não existe mais)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST permitir que uma Tarefa concluída carregue
  um registro estruturado de evidência além da caixa de seleção,
  contendo ao menos: identidade da Tarefa, requisito(s) atendido(s),
  método ou comando de verificação usado, resultado, momento da
  captura, e a ferramenta ou pessoa que a capturou.
- **FR-002**: O sistema MUST distinguir evidência capturada por
  execução automatizada de evidência declarada/importada por uma
  pessoa, registrando explicitamente qual das duas categorias cada
  registro representa.
- **FR-003**: Para execução automatizada de verificação, o sistema
  MUST exigir comando, argumentos, diretório de trabalho e tempo
  limite explícitos — MUST NOT executar texto de Markdown
  implicitamente como comando.
- **FR-004**: A execução automatizada de verificação MUST respeitar as
  permissões já estabelecidas para o agente e para o projeto — MUST
  NOT contornar essas permissões para capturar evidência.
- **FR-005**: Toda evidência MUST registrar a revisão do Git contra a
  qual foi capturada, além de se a árvore de trabalho tinha alterações
  locais relevantes ao escopo verificado naquele momento — o SHA do
  commit isoladamente MUST NOT ser tratado como suficiente quando havia
  alterações locais relevantes.
- **FR-006**: Toda evidência MUST registrar um fingerprint do conteúdo
  contra o qual a verificação foi realizada, de modo que uma mudança
  posterior nesse conteúdo seja detectável.
- **FR-007**: O sistema MUST tratar a evidência de uma Tarefa como
  desatualizada sempre que o conteúdo cujo fingerprint foi registrado
  não corresponder mais ao conteúdo atual — evidência desatualizada
  MUST NOT ser apresentada como verificação corrente e válida.
- **FR-008**: O sistema MUST NOT considerar uma Tarefa concluída apenas
  pela caixa de seleção marcada — a conclusão MUST exigir um registro
  de evidência associado cuja própria verificação não tenha falhado.
- **FR-009**: Uma tentativa de verificação que falha MUST NOT concluir
  a Tarefa — a falha em si MUST ser registrada, de forma distinguível
  de uma verificação nunca tentada.
- **FR-010**: O sistema MUST manter logs extensos de execução fora do
  Context Pack, retornando um resumo verificável com referência ao
  registro completo, em vez de incluir o log bruto diretamente.
- **FR-011**: Toda funcionalidade que hoje trata "caixa marcada" como
  "dependência satisfeita" (por exemplo, cálculo de prontidão de
  Tarefas dependentes) MUST passar a considerar o estado de conclusão
  derivado da evidência, não apenas a caixa de seleção.
- **FR-012**: O sistema MUST oferecer estados de completude mais
  granulares do que um único par concluído/pendente — distinguindo ao
  menos verificado, desatualizado, com falha registrada, e nunca
  verificado.

### Key Entities

- **Registro de Evidência**: um registro estruturado vinculado a uma
  Tarefa, contendo o método/comando de verificação, resultado, momento
  da captura, origem (automática ou declarada), revisão de Git e estado
  da árvore de trabalho, e o fingerprint dos insumos verificados.
- **Estado de Completude**: a classificação real de uma Tarefa, mais
  granular que a caixa de seleção isolada — por exemplo verificado,
  desatualizado, com falha registrada, ou nunca verificado.
- **Fingerprint de Insumos**: a impressão determinística do conteúdo
  contra o qual uma verificação foi realizada, usada para detectar
  quando esse conteúdo mudou depois da captura.
- **Snapshot de Revisão do Repositório**: a revisão de Git e a
  indicação de alterações locais relevantes, registradas no momento da
  captura de uma evidência.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% das Tarefas apresentadas como concluídas após esta
  funcionalidade têm um registro de evidência identificável de forma
  independente da caixa de seleção.
- **SC-002**: 100% das evidências cujos insumos fingerprintados mudam
  depois da captura passam a ser sinalizadas como desatualizadas sem
  qualquer verificação manual adicional.
- **SC-003**: Em 100% dos casos testados, uma Tarefa dependente só é
  apresentada como pronta quando sua dependência upstream tem evidência
  válida associada, nunca apenas pela caixa de seleção da dependência.
- **SC-004**: Em 100% dos cenários de falha testados, uma tentativa de
  verificação malsucedida nunca resulta em uma Tarefa relatada como
  concluída.
- **SC-005**: Usuários conseguem distinguir, sem inspecionar logs
  brutos, se a evidência de uma Tarefa "concluída" foi capturada
  automaticamente ou declarada por uma pessoa, e se ainda é válida
  contra o conteúdo atual, em 100% dos registros exibidos.

## Assumptions

- Esta funcionalidade constrói sobre o modelo canônico de identidade de
  tarefas (Spec 031) e a validação de cobertura/dependências (Spec
  032), já entregues — reutiliza essa identidade e infraestrutura de
  parsing em vez de recriá-las.
- O cálculo de fingerprint reutiliza a capacidade determinística já
  entregue pelo framework, em vez de introduzir um novo algoritmo.
- Execução automatizada de um comando de verificação é opt-in e
  explícita, declarada pela própria Tarefa (por exemplo, em seu próprio
  campo de verificação já existente) — nunca inferida automaticamente a
  partir de texto livre em prosa.
- O armazenamento da evidência permanece consistente com o princípio de
  o sistema de arquivos ser a fonte da verdade do framework — nenhum
  armazenamento externo opaco novo; o formato exato fica a critério do
  planejamento, mas deve continuar inspecionável e reconstruível como
  os demais artefatos.
- Evidência declarada/importada (não capturada ao vivo) é uma categoria
  legítima, de confiança automatizada menor — não é rejeitada, apenas
  rotulada de forma distinta.
- Detectar quando o conteúdo direto de uma Tarefa muda é o gatilho de
  desatualização tratado por esta funcionalidade; uma análise de
  impacto mais ampla, atravessando outros artefatos relacionados, é
  escopo de uma proposta separada (análise de impacto e revisão
  incremental) e não é duplicada aqui.
