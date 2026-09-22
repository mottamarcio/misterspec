# Feature Specification: Análise de Impacto e Revisão Incremental

**Feature Branch**: `042-impact-analysis-review`
**Created**: 2026-09-22
**Status**: Draft
**Input**: User description: "PROP-11 — Análise de impacto e revisão incremental: identificar os artefatos e verificações que precisam de revisão após uma alteração em requisito, plano, conhecimento ou código mapeado, comparando fingerprints, percorrendo relações reversas com regras por tipo (dependência formal, cobertura, evidência, referência contextual), distinguindo invalidação determinística de sugestão de revisão, e retornando caminhos de propagação, motivo, severidade e verificações candidatas a reexecução."

## User Scenarios & Testing *(mandatory)*

<!--
  A Spec 041 já vincula a conclusão de uma Tarefa a evidência
  verificável e a um fingerprint dos insumos contra os quais essa
  evidência foi capturada, detectando quando uma evidência individual
  ficou desatualizada. A Spec 038 já registra a proveniência de cada
  referência wikilink (artefato de origem, seção, localização). O que
  falta é conectar os dois: quando um requisito, plano, nota de
  conhecimento ou código mapeado muda, quem precisa saber disso? Os
  usuários desta funcionalidade são mantenedores e agentes de código
  que acabaram de editar um artefato e precisam responder "o que isso
  afeta, e por quê" sem varrer o repositório manualmente atrás de
  menções e dependências.
-->

### User Story 1 - Saber o que revisar depois de mudar um artefato (Priority: P1)

Como mantenedor ou agente de código que acabou de alterar um
requisito, uma seção do plano, uma nota de conhecimento ou um trecho
de código mapeado a um requisito, quero uma análise que aponte quais
Tarefas, evidências e outros artefatos têm relação conhecida com o que
mudou, para que eu não precise adivinhar ou buscar manualmente por
tudo que pode ter ficado desatualizado.

**Why this priority**: É o valor central da proposta — sem isso, uma
alteração fica isolada do resto do SDD e o mantenedor só descobre o
impacto quando algo já quebrou ou foi implementado sobre uma base
desatualizada.

**Independent Test**: Pode ser testado alterando o conteúdo de um
requisito referenciado por uma Tarefa concluída (com evidência
associada, Spec 041) e confirmando que a análise de impacto aponta
essa Tarefa como afetada, com o motivo e o caminho até o requisito
alterado.

**Acceptance Scenarios**:

1. **Given** um requisito cujo conteúdo muda depois que uma Tarefa foi
   verificada contra ele, **When** a análise de impacto é executada
   para essa mudança, **Then** a Tarefa aparece como afetada, com o
   motivo (fingerprint de evidência desatualizado) e o caminho até o
   requisito alterado.
2. **Given** uma Spec cujo `depends_on` aponta para outra Spec alterada,
   **When** a análise de impacto é executada para a Spec alterada,
   **Then** a Spec dependente aparece como afetada por relação de
   dependência formal.
3. **Given** um artefato sem nenhuma relação conhecida com o elemento
   alterado, **When** a análise de impacto é executada, **Then** esse
   artefato não aparece na lista de itens afetados.

---

### User Story 2 - Distinguir o que foi invalidado do que só merece revisão (Priority: P1)

Como mantenedor avaliando um relatório de impacto, quero que a
análise diferencie claramente entre "isto foi invalidado de forma
determinística e precisa ser reexecutado" e "isto apenas menciona o
elemento alterado e pode merecer uma olhada", para que eu não trate
uma menção incidental em wikilink como se fosse uma prova de que a
implementação está quebrada, nem ignore uma dependência formal
realmente rompida.

**Why this priority**: Sem essa distinção, o relatório de impacto ou
gera alarme falso demais para ser útil (toda menção vira "quebrado")
ou é ignorado por parecer ruído — os dois resultados anulam o valor da
funcionalidade.

**Independent Test**: Pode ser testado alterando um artefato que é
apenas mencionado via wikilink por outro (sem relação formal de
dependência, cobertura ou evidência) e confirmando que esse outro
artefato aparece como "sugestão de revisão", nunca como "invalidado".
Em seguida, alterando um artefato do qual uma Tarefa depende
formalmente e confirmando que essa Tarefa aparece como "invalidada".

**Acceptance Scenarios**:

1. **Given** um artefato A que apenas referencia um artefato B via
   wikilink (sem dependência formal, cobertura ou evidência
   associada), **When** B é alterado e a análise de impacto é
   executada, **Then** A aparece classificado como sugestão de
   revisão, nunca como invalidação determinística.
2. **Given** uma Tarefa cuja evidência foi capturada contra um
   requisito, **When** esse requisito é alterado e a análise de
   impacto é executada, **Then** a Tarefa aparece classificada como
   invalidação determinística, com a verificação candidata a
   reexecução identificada.

---

### User Story 3 - Caminho de propagação, motivo e severidade legíveis (Priority: P2)

Como mantenedor recebendo um relatório de impacto com vários itens
afetados, quero ver, para cada um, o caminho de relações que leva do
elemento alterado até ele, o motivo da inclusão e uma severidade, para
que eu possa priorizar a revisão sem reconstruir manualmente a cadeia
de relações.

**Why this priority**: Uma lista de itens afetados sem explicação nem
priorização obriga o mantenedor a investigar cada um do zero, o que
recria o mesmo custo que a funcionalidade deveria eliminar.

**Independent Test**: Pode ser testado provocando uma cadeia de duas
relações (por exemplo, A depende de B, e uma Tarefa tem evidência
contra B) e confirmando que o item afetado por A traz o caminho
completo até a mudança original, junto com motivo e severidade.

**Acceptance Scenarios**:

1. **Given** uma cadeia de relações com mais de um salto entre o
   elemento alterado e um item afetado, **When** a análise de impacto
   é executada, **Then** o caminho completo de propagação é retornado
   para esse item, não apenas o salto mais próximo.
2. **Given** dois itens afetados pelo mesmo elemento alterado por tipos
   de relação diferentes (um por dependência formal, outro por
   referência contextual), **When** a análise de impacto é executada,
   **Then** os dois recebem severidades coerentes com sua
   classificação (User Story 2), permitindo priorizar o determinístico
   sobre o sugerido.

---

### Edge Cases

- O que acontece quando o elemento alterado não tem nenhuma relação
  reversa conhecida — o relatório apresenta "nenhum impacto conhecido"
  de forma distinguível de "impacto verificado como inexistente"?
- Como o sistema processa um ciclo de referências ou dependências (A
  relaciona-se com B, B relaciona-se de volta com A) sem entrar em
  laço infinito?
- O que acontece quando o vínculo entre código e requisito está
  incompleto ou ausente (a análise de regras de arquitetura e contexto
  de código, PROP-13, ainda não existe) — a análise assume
  conservadoramente que pode haver impacto não capturado, ou omite
  código do relatório?
- Como o sistema trata uma mudança que afeta simultaneamente dezenas
  de itens por meio de um artefato com muitas relações reversas
  (por exemplo, uma nota de conhecimento amplamente referenciada)?
- O que acontece quando dois elementos mudam na mesma execução da
  análise (por exemplo, um requisito e o plano que o detalha) — o
  relatório consolida os caminhos de propagação ou trata cada mudança
  isoladamente?
- Como o sistema trata um artefato removido inteiramente (não apenas
  alterado) em relação a quem dependia, cobria ou referenciava seu
  conteúdo?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST comparar o fingerprint registrado de um
  artefato ou elemento (requisito, seção de plano, nota de
  conhecimento, código mapeado) entre um estado anterior e o estado
  atual, produzindo conjuntos de elementos alterados, adicionados e
  removidos.
- **FR-002**: A partir de cada elemento alterado, adicionado ou
  removido, o sistema MUST percorrer as relações reversas conhecidas
  que apontam para ele, cobrindo ao menos: dependência formal
  (`depends_on`), cobertura de requisito por Tarefa (Spec 032),
  vínculo de evidência por fingerprint de insumos (Spec 041), e
  referência contextual via wikilink (Spec 038).
- **FR-003**: O sistema MUST classificar cada item afetado encontrado
  em uma de duas categorias: invalidação determinística (dependência
  formal rompida, cobertura sem correspondência, ou evidência cujo
  fingerprint de insumos não confere mais) ou sugestão de revisão
  (referência contextual via wikilink, sem relação formal
  adicional).
- **FR-004**: O sistema MUST NOT classificar uma referência contextual
  via wikilink, isoladamente, como invalidação determinística de
  qualquer artefato ou Tarefa que a contenha.
- **FR-005**: Para cada item afetado, o sistema MUST retornar o
  caminho de propagação completo (a sequência de relações do elemento
  alterado até o item), o motivo da inclusão, uma severidade, e,
  quando aplicável, a verificação ou evidência candidata a
  reexecução.
- **FR-006**: O sistema MUST terminar o processamento mesmo na
  presença de ciclos entre relações percorridas, sem expandir o mesmo
  elemento mais de uma vez na mesma análise.
- **FR-007**: Quando o vínculo entre código e requisito estiver
  incompleto ou não existir no projeto analisado, o sistema MUST
  aplicar uma estratégia conservadora explícita (por exemplo, sinalizar
  a lacuna de cobertura em vez de presumir ausência de impacto) e MUST
  declarar essa limitação no relatório, em vez de omiti-la
  silenciosamente.
- **FR-008**: O sistema MUST distinguir, no relatório, "nenhuma relação
  reversa conhecida foi encontrada para este elemento" de "impacto
  verificado como inexistente" — a ausência de relação registrada
  MUST NOT ser apresentada como garantia de que não há impacto.
- **FR-009**: O sistema MUST permitir disparar a análise de impacto a
  partir de uma comparação explícita entre dois estados de um
  artefato ou elemento (por exemplo, revisão anterior e atual), sem
  exigir que o usuário informe manualmente quais relações existem.
- **FR-010**: A severidade atribuída a um item afetado MUST ser
  determinística para a mesma combinação de elemento alterado, tipo de
  relação e caminho de propagação — a mesma entrada MUST NOT produzir
  severidades diferentes em execuções repetidas.

### Key Entities

- **Conjunto de Mudança**: os elementos identificados como alterados,
  adicionados ou removidos entre dois estados comparados de um ou mais
  artefatos, com base em seus fingerprints.
- **Caminho de Propagação**: a sequência ordenada de relações
  reversas que conecta um elemento do Conjunto de Mudança a um item
  afetado, usada para explicar por que esse item apareceu no
  relatório.
- **Classificação de Impacto**: o rótulo atribuído a um item afetado —
  invalidação determinística ou sugestão de revisão — derivado do tipo
  de relação percorrida no Caminho de Propagação.
- **Candidato a Reexecução**: a verificação ou evidência (Spec 041)
  apontada por um item afetado como precisando ser reexecutada ou
  reconfirmada em função da mudança.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Para 100% das mudanças testadas em um elemento com
  relações reversas conhecidas, o relatório de impacto identifica cada
  item afetado com caminho de propagação, motivo e severidade,
  verificável sem busca manual adicional no repositório.
- **SC-002**: 0% dos itens afetados apenas por referência contextual
  via wikilink são apresentados como invalidação determinística nos
  cenários testados.
- **SC-003**: 100% das análises executadas sobre projetos contendo um
  ciclo de relações terminam e retornam um relatório, sem expansão
  redundante do mesmo elemento.
- **SC-004**: Em 100% dos casos testados sem relação reversa conhecida,
  o relatório distingue explicitamente "nenhuma relação conhecida"
  de uma alegação de impacto verificado como inexistente.
- **SC-005**: Mantenedores conseguem, a partir de um único relatório,
  priorizar itens afetados por severidade sem precisar reconstruir
  manualmente a cadeia de relações entre o elemento alterado e cada
  item, em 100% dos relatórios com mais de um item afetado.

## Assumptions

- Esta funcionalidade constrói sobre a identidade canônica de Tarefas
  (Spec 031), a validação de cobertura e dependências (Spec 032), a
  proveniência de referências por wikilink (Spec 038) e as evidências
  de execução vinculadas a fingerprint (Spec 041), já entregues —
  reutiliza essas capacidades em vez de recriá-las.
- O cálculo de fingerprint reutiliza a capacidade determinística já
  entregue pelo framework, coerente com o uso já feito na Spec 041.
- Regras de arquitetura e contexto de código (PROP-13, ainda não
  especificada) não são pré-requisito desta funcionalidade; onde o
  vínculo entre código e requisito ainda não existir, a análise
  assume a estratégia conservadora descrita em FR-007 em vez de
  esperar por essa funcionalidade futura.
- Severidade é uma classificação relativa derivada do tipo de relação
  e do caminho de propagação (por exemplo, dependência formal rompida
  acima de sugestão de revisão), não uma pontuação numérica calibrada
  externamente.
- A análise de impacto é disparada explicitamente (por exemplo, após
  uma alteração ou antes de retomar trabalho), não como um processo
  contínuo em segundo plano.
