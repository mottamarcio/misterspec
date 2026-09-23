# Feature Specification: Regras de Arquitetura e Contexto de Código

**Feature Branch**: `044-architecture-code-context-rules`
**Created**: 2026-09-23
**Status**: Draft
**Input**: User description: "PROP-13 — Regras de arquitetura e contexto de código: aproximar o SDD do código efetivo, verificando restrições arquiteturais configuráveis (módulos, caminhos, imports — dependências proibidas, fronteiras de camadas, contratos exigidos) via adaptadores por linguagem com resultados pass/fail/not_evaluated, iniciando por Go para validar a abordagem no próprio MisterSpec; e recuperando código relevante para a tarefa (arquivos declarados no escopo, assinaturas, testes associados, expandindo para corpos completos quando necessário), indexando código como dado derivado com fingerprints, respeitando exclusões do projeto e o orçamento de contexto já existente."

## User Scenarios & Testing *(mandatory)*

<!--
  O SDD hoje termina no texto: uma Spec declara requisitos e uma Tarefa
  declara um Scope (034-task-oriented-context-preparation), mas nada
  verifica se o código realmente respeita as restrições arquiteturais
  que o projeto declara, e nada ajuda um agente a recuperar apenas o
  código relevante para uma Tarefa sem ler o projeto inteiro à mão.
  Esta funcionalidade tem duas frentes deliberadamente relacionadas: (1)
  verificar restrições arquiteturais configuráveis contra o código
  efetivo, e (2) recuperar código relevante para uma Tarefa do mesmo
  jeito que o Context Engine já recupera Markdown — com orçamento,
  fingerprint e reconstrução a partir do projeto atual. O documento de
  backlog que originou esta proposta já sugere dividi-la em duas Specs
  oficiais se o escopo crescer durante o planejamento; esta
  especificação cobre as duas frentes juntas como um primeiro escopo
  delimitado.
-->

### User Story 1 - Detectar uma dependência arquitetural proibida (Priority: P1)

Como mantenedor, quero declarar restrições de módulo/caminho/import
(dependências proibidas entre camadas, fronteiras de camadas,
contratos exigidos) e ter o sistema apontar uma violação real com
localização reproduzível, para impedir que uma camada interna passe a
depender de algo que não deveria sem depender de revisão manual de
imports a cada mudança.

**Why this priority**: É o valor central da frente de regras de
arquitetura — sem isso, uma restrição declarada é só documentação, sem
nenhum mecanismo que a torne verificável.

**Independent Test**: Pode ser testado declarando uma regra proibindo
um módulo A de importar um módulo B, introduzindo deliberadamente um
import de B dentro de A, e confirmando que a verificação aponta essa
violação com o arquivo e a linha exatos — de forma repetível na mesma
execução ou em execuções diferentes sobre o mesmo código.

**Acceptance Scenarios**:

1. **Given** uma regra declarada proibindo o módulo A de importar o
   módulo B, **When** o código de A contém um import de B, **Then** a
   verificação produz um resultado `fail` identificando o arquivo e a
   linha exatos do import proibido.
2. **Given** a mesma regra, **When** nenhum arquivo de A importa B,
   **Then** a verificação produz um resultado `pass` para essa regra.
3. **Given** uma violação já detectada, **When** a mesma verificação é
   executada novamente sobre o mesmo código, sem nenhuma mudança,
   **Then** o resultado é idêntico — mesma localização, mesmo código de
   identificação.

---

### User Story 2 - Nunca aprovar falsamente uma linguagem ou regra sem suporte (Priority: P1)

Como mantenedor, quando o projeto usa uma linguagem para a qual não
existe adaptador, ou declara uma regra que o adaptador usado não sabe
avaliar, quero que o resultado seja explicitamente "não avaliado" — não
um `pass` silencioso — para nunca confundir ausência de verificação com
conformidade real.

**Why this priority**: Sem essa distinção, a funcionalidade central
(User Story 1) é perigosa: um projeto em uma linguagem não suportada
pareceria "aprovado" nas mesmas regras que um projeto realmente
verificado, escondendo exatamente a lacuna que a proposta existe para
tornar visível.

**Independent Test**: Pode ser testado apontando a verificação para um
projeto (ou uma regra) sem adaptador correspondente e confirmando que o
resultado retornado é `not_evaluated`, distinto tanto de `pass` quanto
de `fail`, e nunca confundido com um `pass` em nenhuma leitura da
resposta.

**Acceptance Scenarios**:

1. **Given** um projeto em uma linguagem sem adaptador implementado,
   **When** a verificação de regras de arquitetura é executada,
   **Then** cada regra aplicável retorna `not_evaluated`, nunca `pass`.
2. **Given** um adaptador existente que não sabe avaliar uma regra
   específica declarada, **When** essa regra é verificada, **Then** o
   resultado para essa regra é `not_evaluated`, mesmo que outras regras
   no mesmo projeto retornem `pass`/`fail` normalmente.
3. **Given** uma resposta contendo resultados mistos (`pass`, `fail`,
   `not_evaluated`), **When** um agente ou mantenedor a lê, **Then**
   consegue distinguir claramente as três categorias, sem precisar
   inferir "não avaliado" a partir da ausência de um resultado.

---

### User Story 3 - Recuperar código relevante para uma Tarefa sem ler o projeto inteiro (Priority: P2)

Como agente preparando a implementação de uma Tarefa que já declara um
Scope (034-task-oriented-context-preparation), quero receber primeiro
os arquivos desse escopo, suas assinaturas e os testes já associados a
eles — e só o corpo completo de um arquivo quando isso realmente for
necessário para entender o que preciso mudar — para não pagar o custo
de ler manualmente cada arquivo do projeto só para descobrir se ele é
relevante.

**Why this priority**: Reduz exatamente o tipo de leitura mecânica
repetida que a preparação orientada à tarefa (034) já eliminou para
Markdown — prioridade menor que as User Stories 1/2 porque essa lacuna
já tem uma mitigação parcial (o agente pode ler os arquivos do Scope
manualmente hoje), enquanto a verificação de arquitetura (US1/US2) não
tem mitigação nenhuma sem esta funcionalidade.

**Independent Test**: Pode ser testado pedindo contexto de código para
uma Tarefa cujo Scope declara um subconjunto pequeno de arquivos do
projeto, e confirmando que a resposta prioriza esses arquivos (com
assinaturas e testes associados) sem incluir o corpo completo de
arquivos fora do Scope, e sem exigir que o agente leia o projeto
inteiro para obter o mesmo resultado.

**Acceptance Scenarios**:

1. **Given** uma Tarefa cujo campo Scope declara um conjunto específico
   de arquivos, **When** o contexto de código é pedido para essa
   Tarefa, **Then** a resposta inclui, no mínimo, esses arquivos e
   qualquer teste já associado a eles, antes de qualquer conteúdo fora
   do Scope.
2. **Given** a mesma Tarefa, **When** o corpo completo de um arquivo do
   Scope não é necessário para representar o que mudou (por exemplo,
   apenas a assinatura já é suficiente), **Then** a resposta pode
   incluir a assinatura em vez do corpo inteiro, em vez de sempre
   despejar o arquivo inteiro por padrão.
3. **Given** um arquivo referenciado no Scope que não existe mais no
   projeto, **When** o contexto de código é pedido, **Then** o sistema
   reporta essa ausência explicitamente, em vez de omiti-la em
   silêncio ou falhar a resposta inteira.

---

### Edge Cases

- Uma regra é declarada para uma linguagem sem nenhum adaptador
  implementado no projeto inteiro — todas as regras aplicáveis a essa
  linguagem retornam `not_evaluated`, não apenas algumas.
- Um mesmo arquivo viola mais de uma regra de arquitetura ao mesmo
  tempo (por exemplo, um import proibido e um contrato exigido
  ausente) — cada violação é reportada como um resultado
  independente, nunca combinada em um único resultado ambíguo.
- Um caminho está explicitamente excluído pela configuração do
  projeto (por exemplo, código gerado ou vendorizado) — nunca aparece
  nem em violações de arquitetura nem em recuperação de código, mesmo
  que combine com uma regra declarada.
- Um arquivo do Scope de uma Tarefa é grande demais para caber mesmo
  no nível de assinatura dentro do orçamento restante — tratado pelo
  mesmo mecanismo de exclusão/registro que outros conteúdos do Context
  Engine já usam, não como uma falha especial de código.
- Um arquivo é renomeado ou movido entre execuções — a próxima
  verificação/indexação reflete o estado atual; nenhuma verificação
  aponta para uma localização que não existe mais.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST permitir declarar restrições
  arquiteturais configuráveis — dependências de módulo proibidas,
  fronteiras de camadas e contratos exigidos — como dados do projeto,
  não como lógica fixa no binário.
- **FR-002**: O sistema MUST avaliar as regras declaradas através de
  um adaptador específico da linguagem do projeto, normalizando cada
  resultado e reportando a localização exata (arquivo e linha) de cada
  violação encontrada.
- **FR-003**: Uma violação de dependência proibida MUST ser reportada
  como um resultado com identificação estável (código reproduzível),
  distinguível de outros diagnósticos estruturais que o sistema já
  produz.
- **FR-004**: Quando não existe adaptador para a linguagem do projeto,
  ou quando o adaptador usado não sabe avaliar uma regra declarada
  específica, o resultado para essa regra MUST ser `not_evaluated` —
  MUST NOT ser reportado como `pass`.
- **FR-005**: Toda resposta de verificação de arquitetura MUST
  distinguir claramente entre os três resultados possíveis (`pass`,
  `fail`, `not_evaluated`) para cada regra avaliada — nunca apresentar
  a ausência de um resultado como equivalente a `pass`.
- **FR-006**: O sistema MUST incluir um adaptador para Go como
  primeira implementação, validado contra o próprio código-fonte do
  MisterSpec e suas próprias restrições arquiteturais declaradas.
- **FR-007**: Para uma Tarefa com um campo Scope já declarado (034),
  o sistema MUST recuperar, em ordem de prioridade, os arquivos desse
  Scope, suas assinaturas exportadas, e qualquer arquivo de teste já
  associado a eles, antes de qualquer conteúdo de código fora do
  Scope declarado.
- **FR-008**: O corpo completo de um arquivo MUST ser incluído apenas
  quando o nível de assinatura/teste não for suficiente para
  representar o que a Tarefa precisa — nunca como padrão incondicional
  quando o nível mais leve já cobre o pedido.
- **FR-009**: Código indexado para esta recuperação MUST ser tratado
  como dado derivado, descartável e reconstruível a partir dos
  arquivos atuais do projeto a qualquer momento — nunca como estado
  autoritativo — e cada item retornado MUST carregar um fingerprint do
  seu próprio conteúdo.
- **FR-010**: A indexação e recuperação de código MUST respeitar as
  exclusões já declaradas pelo projeto (por exemplo, diretórios
  gerados ou vendorizados) e MUST contar contra o mesmo orçamento de
  contexto já estabelecido pelo Context Engine — nunca um orçamento
  paralelo específico de código.
- **FR-011**: A eficácia da recuperação de contexto de código (se de
  fato reduz custo de leitura mantendo qualidade, comparado a ler
  arquivos inteiros) MUST ser mensurável através do protocolo de
  avaliação já existente — nunca apenas alegada sem essa medição.

### Key Entities

- **Regra de Arquitetura**: uma restrição declarada sobre módulos,
  caminhos ou imports (dependência proibida, fronteira de camada,
  contrato exigido), associada a um projeto e a uma linguagem.
- **Adaptador de Linguagem**: o avaliador específico de uma linguagem
  que verifica regras declaradas contra o código real, normalizando
  resultado e localização; sua ausência (ou a ausência de suporte a
  uma regra específica dentro dele) produz `not_evaluated`, nunca uma
  aprovação presumida.
- **Resultado de Arquitetura**: um resultado de avaliação de uma regra
  (`pass`, `fail`, ou `not_evaluated`), com localização e identificação
  reproduzível quando aplicável.
- **Item de Contexto de Código**: uma unidade de código recuperada
  relevante ao Scope de uma Tarefa — assinatura, teste associado, ou
  corpo completo — carregando seu próprio fingerprint de conteúdo.
- **Índice de Código**: o registro derivado e descartável de código
  indexado usado para a recuperação, reconstruível a partir dos
  arquivos atuais do projeto a qualquer momento.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Uma dependência proibida introduzida entre dois módulos
  com regra declarada é detectada e reportada com localização
  reproduzível em 100% dos casos testados.
- **SC-002**: Um projeto ou regra sem suporte de adaptador nunca é
  reportado como aprovado — 0 aprovações falsas em 100% das
  combinações de linguagem/regra sem suporte testadas.
- **SC-003**: Para uma Tarefa com Scope declarado, um agente recebe as
  assinaturas e testes associados a esse Scope sem precisar ler o
  projeto inteiro, com uma redução mensurável na quantidade de código
  retornado comparado a ler todos os arquivos do projeto, em 100% dos
  cenários testados com Scope menor que o projeto inteiro.
- **SC-004**: Todo item de código incluído em uma resposta pode ser
  reconstruído/atualizado a partir dos arquivos atuais do projeto a
  qualquer momento, com 100% de correspondência entre o fingerprint
  reportado e o conteúdo real do arquivo quando este não mudou.
- **SC-005**: Uma comparação através do protocolo de avaliação já
  existente mostra que a recuperação de contexto de código, nos casos
  testados, mantém a taxa de conclusão correta de tarefas pelo menos
  no mesmo nível que ler arquivos inteiros — uma comparação registrada,
  nunca apenas uma alegação.

## Assumptions

- Esta funcionalidade constrói sobre o campo Scope já estabelecido
  pela preparação orientada à tarefa (Spec 034) e sobre o protocolo de
  avaliação de qualidade/eficiência já entregue (Spec 037) — reutiliza
  ambos em vez de recriar a declaração de escopo ou a medição
  comparativa.
- Reutiliza o orçamento e o estimador de contexto já existentes
  (Spec 035) para dimensionar código retornado — código conta contra o
  mesmo orçamento de qualquer outro conteúdo selecionado, sem um
  orçamento paralelo específico para código.
- O primeiro Adaptador de Linguagem cobre Go, validado contra o
  próprio código-fonte do MisterSpec (precedente de dogfooding já
  estabelecido pela Constituição do projeto) — suporte a outras
  linguagens fica fora do escopo até que o adaptador de Go valide a
  abordagem.
- Conforme a recomendação do próprio documento de backlog, se uma das
  duas frentes (regras de arquitetura ou contexto de código) crescer
  substancialmente durante o planejamento, ela PODE ser dividida em
  duas Specs oficiais separadas — esta especificação cobre as duas
  frentes juntas como um primeiro escopo delimitado, deixando essa
  decisão de divisão para a fase de planejamento, se necessário.
- Regras de arquitetura são declaradas explicitamente, nunca inferidas
  automaticamente a partir de padrões observados no código.
