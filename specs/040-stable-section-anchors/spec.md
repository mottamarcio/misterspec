# Feature Specification: Referências a Seções com Âncoras Estáveis

**Feature Branch**: `040-stable-section-anchors`
**Created**: 2026-09-22
**Status**: Draft
**Input**: User description: "PROP-09 — Referências a seções com âncoras estáveis: permitir apontar para uma política ou requisito específico sem carregar todo o artefato de destino, usando uma extensão de wikilink como `[[KNOW-003#retry-policy|Política de retries]]`, com âncoras explícitas e estáveis independentes de pequenas alterações no título, validação de unicidade/existência/resolução, recuperação da seção alvo com o mínimo de contexto hierárquico necessário, e sem introduzir embeds ou transclusão."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Referenciar apenas a seção relevante (Priority: P1)

Como autor de um artefato, quero apontar um wikilink para uma seção
específica de outro artefato (por exemplo, uma política de retries
dentro de um documento de Conhecimento maior), para que quem segue essa
referência receba apenas o trecho relevante, sem precisar carregar ou
ler o artefato de destino inteiro.

**Why this priority**: É o valor central da proposta — sem isso, uma
referência precisa a uma única seção continua custando o carregamento
do documento inteiro, desperdiçando contexto e obrigando o leitor a
localizar manualmente a parte relevante.

**Independent Test**: Pode ser testado escrevendo um wikilink com
âncora de seção para um artefato que tenha múltiplas seções e
confirmando que a recuperação retorna apenas a seção referenciada (mais
o contexto hierárquico mínimo necessário para preservar seu sentido),
não o artefato completo.

**Acceptance Scenarios**:

1. **Given** um artefato de destino com várias seções, cada uma com uma
   âncora própria, **When** um wikilink referencia uma âncora
   específica, **Then** a recuperação retorna apenas o conteúdo dessa
   seção (mais o mínimo de contexto hierárquico necessário), não o
   artefato inteiro.
2. **Given** um wikilink sem âncora (`[[ID]]` ou `[[ID|Alias]]`),
   **When** a recuperação é executada, **Then** o comportamento
   permanece exatamente o mesmo de antes desta funcionalidade —
   nenhuma referência existente muda de significado.
3. **Given** um wikilink com âncora e alias combinados
   (`[[ID#âncora|Texto de apresentação]]`), **When** o link é
   processado, **Then** a âncora determina a seção recuperada e o alias
   permanece apenas apresentação, sem afetar a resolução.

---

### User Story 2 - Âncora sobrevive a uma pequena edição de título (Priority: P2)

Como autor de um artefato de Conhecimento, quero declarar uma âncora
explícita para uma seção, independente do texto do seu título, para
que eu possa melhorar a redação do título mais tarde sem quebrar todas
as referências existentes àquela seção.

**Why this priority**: Sem uma âncora explícita e estável, qualquer
edição de título — mesmo cosmética — quebraria silenciosamente toda
referência existente àquela seção, desencorajando revisões editoriais
legítimas e tornando as referências frágeis.

**Independent Test**: Pode ser testado declarando uma âncora explícita
em uma seção, referenciando-a por um wikilink, depois editando apenas o
texto do título da seção (mantendo a âncora declarada) e confirmando
que a referência continua resolvendo para a mesma seção sem exigir
qualquer edição no wikilink.

**Acceptance Scenarios**:

1. **Given** uma seção com uma âncora explícita declarada e um wikilink
   que a referencia, **When** o título da seção é reescrito mas a
   âncora declarada permanece a mesma, **Then** o wikilink continua
   resolvendo corretamente para essa seção, sem qualquer edição
   necessária no wikilink.
2. **Given** duas seções diferentes no mesmo artefato, **When** ambas
   declaram a mesma âncora, **Then** isso é sinalizado como um problema
   de unicidade antes que qualquer wikilink apontando para essa âncora
   seja considerado confiável.

---

### User Story 3 - Âncora quebrada nunca vira link para o documento inteiro (Priority: P2)

Como autor ou revisor, quero que uma âncora inexistente ou removida
produza um diagnóstico claro, para que eu nunca seja levado a acreditar
que uma referência precisa continua válida quando, na verdade, ela
aponta para algo que não existe mais.

**Why this priority**: O maior risco de silenciosamente "degradar" uma
âncora quebrada para um link de documento inteiro é esconder uma perda
real de precisão da referência — o leitor recebe conteúdo a mais sem
saber que a referência original havia se tornado inválida.

**Independent Test**: Pode ser testado removendo ou renomeando uma
âncora declarada sem atualizar o wikilink que a referencia e
confirmando que a validação estrutural sinaliza claramente essa âncora
como quebrada, distinguindo-a de um link de documento inteiro quebrado.

**Acceptance Scenarios**:

1. **Given** um wikilink que referencia uma âncora que foi removida do
   artefato de destino, **When** a validação estrutural é executada,
   **Then** o problema é sinalizado explicitamente como uma âncora
   quebrada — nunca tratado silenciosamente como um link válido para o
   documento inteiro.
2. **Given** um wikilink que referencia um artefato inexistente
   combinado com uma âncora, **When** a validação é executada, **Then**
   o diagnóstico deixa claro qual das duas coisas falhou (artefato
   inexistente, ou âncora inexistente dentro de um artefato existente).

### Edge Cases

- O que acontece quando duas seções diferentes, no mesmo artefato,
  declaram a mesma âncora?
- Como o sistema trata uma âncora declarada mas nunca referenciada por
  nenhum wikilink (não deve ser tratada como erro)?
- O que acontece quando uma seção referenciada por âncora está aninhada
  sob várias subseções — quanto de contexto hierárquico ascendente é
  necessário para preservar seu sentido, sem recuperar irmãos ou
  seções não relacionadas?
- Como o sistema trata um wikilink cuja âncora aponta para uma seção
  cujo próprio corpo está vazio (título imediatamente seguido de outro
  título)?
- O que acontece quando um artefato de destino não declara nenhuma
  âncora e é referenciado com uma âncora mesmo assim?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST permitir que um wikilink referencie uma
  âncora específica dentro de um artefato de destino, usando uma
  extensão de sintaxe reconhecível (por exemplo
  `[[ID#âncora]]`/`[[ID#âncora|Alias]]`), sem alterar o significado dos
  wikilinks existentes sem âncora.
- **FR-002**: O sistema MUST permitir que o autor de um artefato
  declare uma âncora explícita para qualquer seção, de forma
  independente do texto atual do título dessa seção.
- **FR-003**: A identidade de uma âncora declarada MUST permanecer
  estável através de uma edição do título da seção, desde que a própria
  declaração da âncora não seja alterada.
- **FR-004**: O sistema MUST validar que toda âncora referenciada por
  um wikilink existe no artefato de destino declarado, sinalizando uma
  âncora inexistente como um problema distinto e específico — nunca
  tratando-a implicitamente como um link válido para o documento
  inteiro.
- **FR-005**: O sistema MUST detectar e sinalizar âncoras duplicadas
  declaradas dentro do mesmo artefato, antes que qualquer resolução
  baseada nelas seja considerada confiável.
- **FR-006**: Ao recuperar conteúdo para uma referência com âncora, o
  sistema MUST retornar apenas o conteúdo da seção referenciada mais o
  mínimo de contexto hierárquico ascendente necessário para preservar
  seu sentido — nunca o artefato de destino inteiro nem seções
  irmãs/não relacionadas.
- **FR-007**: Uma referência com âncora MUST permanecer uma referência
  (localização e recuperação sob demanda) — o sistema MUST NOT
  incorporar (embed) ou fundir (transcluir) o conteúdo da seção
  referenciada dentro do artefato que a referencia.
- **FR-008**: O sistema MUST manter total compatibilidade com wikilinks
  de documento inteiro já existentes (`[[ID]]`, `[[ID|Alias]]`) — nenhum
  wikilink existente muda de comportamento por causa desta
  funcionalidade.
- **FR-009**: O sistema MUST distinguir, no diagnóstico de validação,
  entre "artefato de destino inexistente" e "âncora inexistente dentro
  de um artefato existente", para que o autor saiba exatamente o que
  corrigir.

### Key Entities

- **Âncora de Seção**: um identificador explícito e estável, declarado
  pelo autor para uma seção específica de um artefato, independente do
  texto do título dessa seção.
- **Referência com Âncora**: a parte `#âncora` de um wikilink,
  qualificando a referência a uma seção específica do artefato alvo em
  vez do artefato inteiro.
- **Resolução de Âncora**: o resultado de associar uma Referência com
  Âncora à Âncora de Seção real que ela nomeia — bem-sucedida, ou
  sinalizada como quebrada com um motivo específico.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em 100% das resoluções de uma referência com âncora
  válida, o conteúdo retornado corresponde exatamente à seção
  referenciada (mais o contexto hierárquico mínimo), nunca ao artefato
  de destino inteiro.
- **SC-002**: Renomear o título de uma seção, mantendo sua âncora
  declarada, preserva 100% das referências existentes a essa âncora,
  sem exigir qualquer edição nos wikilinks que a referenciam.
- **SC-003**: 100% das âncoras referenciadas mas inexistentes são
  sinalizadas pela validação estrutural com um diagnóstico específico e
  distinto de um link de documento inteiro quebrado.
- **SC-004**: 100% das declarações de âncora duplicadas dentro de um
  mesmo artefato são detectadas pela validação antes de qualquer uso em
  navegação ou recuperação de contexto.
- **SC-005**: Para um artefato de destino com múltiplas seções, o
  conteúdo recuperado por uma referência com âncora é mensuravelmente
  menor do que o artefato completo, comprovando que a precisão da
  referência reduz o volume de contexto entregue.

## Assumptions

- Esta funcionalidade constrói sobre a proveniência de wikilinks por
  trecho já entregue (Spec 038) e sobre a segmentação em seções/Chunks
  já existente no framework; não reimplementa o parser de wikilinks ou
  a segmentação de documentos do zero.
- A sintaxe de declaração de âncora em Markdown é uma extensão
  explícita e visível no próprio título da seção (não inferida a partir
  do texto do título), seguindo um padrão já familiar de extensões de
  Markdown que anexam um identificador estável a um cabeçalho.
- "Mínimo de contexto hierárquico necessário" significa a cadeia de
  títulos ancestrais da seção referenciada (para preservar a que
  política/seção maior ela pertence), não o conteúdo integral das
  seções ancestrais nem de seções irmãs.
- Uma âncora declarada mas nunca referenciada por nenhum wikilink não é
  tratada como erro — declarar uma âncora "adiantada" para uso futuro é
  um caso válido.
- Esta funcionalidade não introduz um mecanismo novo de armazenamento
  persistente: âncoras continuam sendo informação derivável do Markdown
  de origem, reconstruível como o restante do índice de contexto.
