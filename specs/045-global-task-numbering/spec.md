# Feature Specification: Numeração Global de Tasks

**Feature Branch**: `045-global-task-numbering`
**Created**: 2026-09-28
**Status**: Draft
**Input**: User description: "queria implementar algo para corrigir um comportamento que os LLMs estão sempre me questionando: a questão da nomenclatura das features, specs e tasks. O comportamento atual é que ele sempre reiniciam a contagem (por exemplo, FEAT-001/SPEC-002/TASK-001; FEAT-002/SPEC-001/TASK-001, etc). Seria melhor manter sempre incremental (por exemplo, FEAT-001/SPEC-11/TASK-020; FEAT-002/SPEC-020/TASK-100). Isso vale para todos os program, feature, spec e tasks"

## User Scenarios & Testing *(mandatory)*

<!--
  Investigação prévia (durante a especificação) confirmou que Program,
  Feature e Spec já são numerados de forma globalmente incremental
  no código atual (internal/ids: Scan enumera cada tipo em todo o
  projeto via glob, e NextID calcula o maior número existente + 1,
  nunca reiniciando por Program/Feature pai — ver
  internal/ids/scan_test.go "TestScan_SpecsAcrossFeaturesAndPrograms").
  O único tipo que reinicia por design é Task: 031-canonical-task-
  identity estabeleceu deliberadamente que um número de Task só
  precisa ser único dentro do seu próprio Spec, com a forma composta
  "SPEC-###:TASK-###" resolvendo qualquer ambiguidade — colisões de
  número entre Specs diferentes são tratadas como "esperadas, válidas"
  (ver internal/ids/scan.go TaskCollision, e o diagnóstico somente-
  leitura `internal migration-check-tasks`).

  Esta especificação, após confirmação explícita do solicitante,
  escopa-se a: dar ao Skill que decompõe Tasks (kit/skills/mister-
  tasks) uma forma determinística de saber qual o próximo número de
  Task nunca usado em nenhum Spec do projeto — eliminando a incerteza
  que hoje leva o agente a perguntar se deve reiniciar a numeração —
  sem reverter o modelo de identidade composta nem invalidar Specs
  já existentes que hoje legitimamente reaproveitam números de Task
  entre Specs diferentes.
-->

### User Story 1 - Próximo número de Task nunca reinicia (Priority: P1)

Como agente (LLM) executando o Skill de decomposição de Tasks para um
Spec, ao numerar cada novo `## TASK-NNN` preciso saber qual o próximo
número de Task ainda não usado por nenhum Spec do projeto — não apenas
o próximo número dentro do Spec atual — para nunca mais precisar
perguntar ao usuário se a numeração deve reiniciar a cada novo Spec.

**Why this priority**: É a queixa direta que motivou esta
especificação — o comportamento hoje gera reinícios (TASK-001 em
FEAT-002/SPEC-001 tão certo quanto TASK-001 em FEAT-001/SPEC-002), e
isso já foi confirmado como o único ponto realmente inconsistente
(Program/Feature/Spec já são globais). Resolver isso sozinho já entrega
o valor pedido.

**Independent Test**: Num projeto onde o maior número de Task já
existente em qualquer Spec é TASK-047, ao decompor um Spec novo (ainda
sem nenhuma Task própria) o primeiro Task número obtido é TASK-048 —
nunca TASK-001.

**Acceptance Scenarios**:

1. **Given** um projeto com Tasks em múltiplos Specs, sendo o maior
   número existente TASK-047 (em qualquer Spec), **When** o Skill de
   decomposição de Tasks pede o próximo número de Task para um Spec
   novo e vazio, **Then** o número retornado é TASK-048.
2. **Given** um projeto sem nenhuma Task em nenhum Spec ainda, **When**
   o próximo número de Task é solicitado, **Then** o número retornado é
   TASK-001.
3. **Given** um Spec que já tem Tasks TASK-050 e TASK-051 e o projeto
   não tem nenhuma Task com número maior em outro Spec, **When** uma
   nova Task é adicionada a esse mesmo Spec, **Then** o número
   retornado é TASK-052 (continua incremental tanto dentro do Spec
   quanto no projeto inteiro).

---

### User Story 2 - Referência isolada a uma Task permanece inequívoca (Priority: P2)

Como pessoa mantenedora lendo o histórico ou a documentação do projeto,
ao encontrar uma referência isolada `TASK-NNN` (sem o prefixo do Spec),
quero que ela identifique uma única Task em todo o projeto a partir da
adoção desta funcionalidade, sem precisar vasculhar o `tasks.md` de
cada Spec para descobrir a qual Task pertence.

**Why this priority**: É uma consequência direta e desejável da User
Story 1, mas depende dela — sem números globais, uma referência
isolada continua ambígua por natureza.

**Independent Test**: Após a adoção desta funcionalidade, resolver uma
referência isolada `TASK-NNN` de qualquer Task criada depois da adoção
retorna exatamente uma correspondência em todo o projeto.

**Acceptance Scenarios**:

1. **Given** duas Tasks criadas após a adoção desta funcionalidade, em
   dois Specs diferentes, **When** cada uma é referenciada apenas pelo
   seu número isolado `TASK-NNN`, **Then** cada referência resolve para
   exatamente uma Task, nunca mais de uma.
2. **Given** a forma composta `SPEC-###:TASK-###` já usada hoje,
   **When** essa forma é usada para referenciar uma Task, **Then** ela
   continua funcionando exatamente como funciona hoje — esta
   funcionalidade não remove nem descontinua a forma composta.

---

### User Story 3 - Projetos existentes continuam válidos sem migração (Priority: P2)

Como pessoa mantenedora de um projeto que já tem Specs com números de
Task legitimamente repetidos entre Specs diferentes (criados antes
desta funcionalidade existir), quero que adotar esta funcionalidade não
exija renumerar ou migrar nenhuma Task existente, para que a adoção
seja segura.

**Why this priority**: Sem esta garantia, a funcionalidade quebraria
todo o histórico dos 44 Specs já implementados neste próprio projeto —
inaceitável como efeito colateral de uma melhoria de conveniência.

**Independent Test**: Rodar o diagnóstico de colisão entre Specs já
existente contra o conteúdo atual do projeto continua reportando
exatamente as mesmas colisões de antes, sem nenhuma mudança — nenhuma
Task é renomeada, renumerada ou removida por esta funcionalidade.

**Acceptance Scenarios**:

1. **Given** o conjunto de Specs já existentes no projeto (com
   colisões de número de Task entre Specs diferentes, hoje válidas),
   **When** esta funcionalidade é adotada, **Then** o diagnóstico de
   colisão entre Specs continua reportando o mesmo conjunto de
   colisões pré-existentes, sem nenhum erro novo.
2. **Given** uma Task já existente cujo número colide com o de outra
   Task em outro Spec, **When** a validação estrutural do Spec é
   executada, **Then** essa colisão pré-existente continua não sendo
   tratada como erro bloqueante.

---

### Edge Cases

- O que acontece se dois agentes, em dois terminais diferentes,
  pedirem o "próximo número de Task" quase ao mesmo tempo para dois
  Specs diferentes, antes que qualquer um dos dois grave sua própria
  Task em disco? Ambos podem receber o mesmo número sugerido — esta
  funcionalidade não introduz nenhum mecanismo de reserva/trava; ela
  apenas troca uma estimativa manual e escopada ao Spec atual por uma
  estimativa determinística e escopada ao projeto inteiro. Essa mesma
  limitação (leitura seguida de escrita, sem trava) já existe hoje para
  Program/Feature/Spec fora do bloqueio de `internal create`, e para
  Task — que nunca teve nenhuma operação de criação — a situação já era
  ainda menos protegida antes desta funcionalidade. Ela não piora nem
  resolve esse cenário; apenas o próprio `internal validate`/o
  diagnóstico de colisão continuam sendo a rede de segurança que
  detecta o resultado depois do fato.
- O que acontece se o Spec que hoje detém o maior número de Task for
  removido do projeto? O próximo número calculado passa a considerar
  apenas o que ainda existe no disco — comportamento idêntico ao já
  aceito hoje para Program/Feature/Spec (uma renumeração após remoção
  nunca foi garantida antes desta funcionalidade).
- O que acontece com uma Task cujo cabeçalho `## TASK-NNN` foi escrito
  à mão, sem passar pela nova operação determinística? Nada muda:
  continua sendo aceita exatamente como hoje — este recurso não se
  torna uma trava de escrita, apenas uma fonte de verdade opcional que
  o Skill de decomposição passa a consultar antes de escolher um
  número.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema DEVE fornecer uma operação somente-leitura que
  calcule o próximo número de Task disponível como o maior número de
  Task encontrado em qualquer `tasks.md` do projeto inteiro, mais um
  (ou 1, se nenhuma Task existir ainda) — nunca escopado a um único
  Spec, espelhando a mesma regra que Program/Feature/Spec já seguem
  hoje.
- **FR-002**: O Skill que decompõe o Plano de um Spec em Tasks DEVE
  obter o número de cada nova Task a partir dessa operação, antes de
  escrever o cabeçalho `## TASK-NNN`, em vez de determinar o número
  lendo apenas o `tasks.md` do Spec atual.
- **FR-003**: A operação do FR-001 NÃO DEVE escrever no sistema de
  arquivos nem reservar/persistir o número calculado em nenhum estado
  — é consistente com o fato de Task nunca ter tido uma operação de
  criação própria (o cabeçalho continua sendo escrito à mão/pelo
  agente diretamente no `tasks.md`).
- **FR-004**: O comportamento estrutural já existente para números de
  Task (detecção de duplicata dentro do mesmo Spec, e o diagnóstico
  somente-leitura de colisão entre Specs diferentes) NÃO DEVE mudar —
  uma colisão entre Specs pré-existente continua sendo histórico
  válido, não bloqueante; esta funcionalidade só muda como um número
  *novo* é escolhido, nunca como um número já existente é julgado.
- **FR-005**: O diagnóstico de colisão entre Specs já existente
  (`internal migration-check-tasks`) DEVE continuar reportando
  exatamente as mesmas colisões históricas de antes, sem alteração de
  comportamento causada por esta funcionalidade.
- **FR-006**: Tanto a referência isolada `TASK-NNN` quanto a forma
  composta `SPEC-###:TASK-NNN` DEVEM continuar sendo aceitas em todo
  lugar que já resolve referências de Task hoje — a forma composta não
  é removida nem descontinuada por esta funcionalidade.
- **FR-007**: Quando o projeto não tiver nenhuma Task em nenhum Spec, a
  operação do FR-001 DEVE retornar TASK-001, espelhando o
  comportamento já existente de `NextID` para os demais tipos de
  entidade quando nenhuma instância existe ainda.

### Key Entities

- **Alocador de Próximo Número de Task**: uma operação determinística,
  somente-leitura, sem estado persistido — uma função pura do que já
  existe em disco no momento da chamada, análoga à já existente
  `NextID` de Program/Feature/Spec, mas aplicada a Task com escopo de
  projeto inteiro em vez de escopo por Spec.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em um projeto onde o maior número de Task já existente
  em qualquer Spec é N, a primeira Task de um Spec recém-decomposto é
  sempre numerada N+1, nunca 1, em 100% dos casos.
- **SC-002**: O Skill de decomposição de Tasks passa a resolver essa
  numeração sozinho — sem depender de o agente perguntar ao usuário se
  a contagem deve reiniciar a cada novo Spec.
- **SC-003**: Adotar esta funcionalidade em um projeto já existente,
  com colisões de número de Task entre Specs pré-existentes, não exige
  nenhum passo manual de renumeração e não produz nenhum erro de
  validação novo sobre esse conteúdo pré-existente.
- **SC-004**: Resolver, pelo número isolado `TASK-NNN`, qualquer Task
  criada após a adoção desta funcionalidade retorna exatamente uma
  correspondência em todo o projeto.

## Assumptions

- Program, Feature e Spec já são numerados de forma globalmente
  incremental no código atual (confirmado em `internal/ids`) — o
  escopo desta funcionalidade se limita a Task e ao Skill que a
  escreve.
- O modelo de identidade composta de Task (031-canonical-task-
  identity) — um número de Task só precisa ser único dentro do
  próprio Spec, com `SPEC-###:TASK-###` como forma inequívoca — não
  está sendo substituído nem revertido por esta funcionalidade; a
  forma composta continua funcionando exatamente como hoje.
- As Tasks já existentes nos Specs 001-044 (algumas das quais já
  reaproveitam legitimamente o mesmo número em Specs diferentes) não
  são renumeradas, migradas nem tocadas por esta funcionalidade.
- Nenhuma operação de criação de Task é introduzida — a autoria de
  Task continua sendo edição direta do `tasks.md` pelo Skill de
  decomposição, respeitando o limite arquitetural já existente do
  projeto (Task nunca teve uma operação de criação própria); apenas a
  pergunta "qual número eu uso agora" passa a ter uma resposta
  determinística e de escopo global, em vez de uma estimativa manual
  escopada ao Spec atual.
- A decomposição concorrente de Tasks para dois Specs diferentes, por
  dois agentes independentes, pode em teoria receber o mesmo "próximo"
  número antes que qualquer um dos dois grave sua própria Task em
  disco — é uma limitação aceita, idêntica à mesma condição de
  leitura-seguida-de-escrita que já existe hoje na própria `NextID` de
  Program/Feature/Spec; esta funcionalidade não introduz nenhuma trava
  além da que já existe.
