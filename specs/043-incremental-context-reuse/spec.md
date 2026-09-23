# Feature Specification: Reutilização Incremental de Context Packs

**Feature Branch**: `043-incremental-context-reuse`
**Created**: 2026-09-22
**Status**: Draft
**Input**: User description: "PROP-12 — Reutilização incremental de Context Packs: reduzir repetição de conteúdo entre tarefas ou retomadas, identificando pacotes por hash dos insumos/tarefa/política de ranking/formato/estimador, comparando com um pacote-base conhecido para retornar apenas diferenças (adicionadas, modificadas e removidas), enviando apenas diferenças quando o consumidor confirmar posse do pacote-base, recuperando o pacote completo com segurança quando a base estiver ausente/desatualizada, e invalidando o cache por mudança de conteúdo, configuração ou versão de contrato — sem confundir esse cache local com o cache de prompt do provedor de LLM."

## User Scenarios & Testing *(mandatory)*

<!--
  A Spec 033 já entrega ao agente o conteúdo completo, localizável e
  versionado de um Context Pack, em vez de apenas metadados. A Spec 035
  já garante que o orçamento reportado reflete o que foi de fato
  entregue. O que ainda falta: quando um agente pede um Context Pack
  para o mesmo alvo mais de uma vez — entre Tarefas da mesma Spec, ou
  ao retomar trabalho depois de perder ou compactar seu próprio
  contexto — ele recebe o pacote inteiro de novo, mesmo que a maior
  parte do conteúdo não tenha mudado desde a última vez. Esta
  funcionalidade deixa o agente (ou quem o chama) pedir apenas o que
  mudou desde um pacote que ele já tem, com um mecanismo de segurança
  claro para quando essa suposição de posse não se confirma.
-->

### User Story 1 - Pedir só o que mudou desde o último pacote (Priority: P1)

Como agente que já recebeu um Context Pack para um alvo (por exemplo,
ao começar a Tarefa 1 de uma Spec) e volta a pedir contexto para o
mesmo alvo pouco depois (por exemplo, ao começar a Tarefa 2 da mesma
Spec), quero poder informar que já possuo aquele pacote e receber
apenas o que mudou desde então, para não pagar de novo o custo de
receber conteúdo que já tenho e não mudou.

**Why this priority**: É o valor central da proposta — sem isso, cada
pedido de contexto para o mesmo alvo é tratado como se fosse o
primeiro, mesmo quando a maior parte do conteúdo é idêntica ao pedido
anterior.

**Independent Test**: Pode ser testado pedindo um Context Pack para um
alvo, depois pedindo novamente para o mesmo alvo informando o pacote
anterior como base, sem que o conteúdo de origem tenha mudado entre os
dois pedidos, e confirmando que a segunda resposta é significativamente
menor que a primeira, mas ainda permite reconstruir o pacote completo
original.

**Acceptance Scenarios**:

1. **Given** um pacote já recebido para um alvo, **When** o mesmo alvo
   é pedido novamente informando esse pacote como base, e nada mudou
   no conteúdo de origem, **Then** a resposta retorna um conjunto vazio
   de diferenças, não o pacote inteiro repetido.
2. **Given** um pacote já recebido para um alvo, **When** apenas um
   trecho do conteúdo de origem muda e o mesmo alvo é pedido novamente
   informando o pacote anterior como base, **Then** a resposta identifica
   apenas os trechos adicionados, modificados e removidos desde a base
   — o restante do conteúdo não é reenviado.
3. **Given** uma resposta de diferenças recebida, **When** essa
   diferença é aplicada sobre o pacote-base original, **Then** o
   resultado reproduz exatamente o pacote completo que um pedido direto
   (sem base) teria retornado para o mesmo alvo e configuração.

---

### User Story 2 - Recuperar o pacote completo com segurança quando a base não existe mais (Priority: P1)

Como agente que perdeu ou teve seu próprio contexto compactado (e por
isso não tem mais certeza de que ainda possui o pacote-base que
pretendia informar, ou nunca recebeu esse pacote de fato), quero que o
sistema nunca confie cegamente na minha alegação de posse — e que, ao
não conseguir confirmar a base, eu receba o pacote completo em vez de
uma diferença incompleta que eu não conseguiria interpretar sozinho.

**Why this priority**: Sem essa garantia, a funcionalidade central
(User Story 1) é perigosa: um agente que alega possuir uma base que na
verdade não tem mais recebe uma resposta incompleta e a trata como se
fosse o conteúdo inteiro, silenciosamente perdendo informação.

**Independent Test**: Pode ser testado pedindo uma diferença contra um
identificador de pacote-base que nunca existiu, ou que já foi
invalidado, e confirmando que a resposta é o pacote completo — nunca
um erro que trava o fluxo, nem uma diferença parcial tratada como
resposta completa.

**Acceptance Scenarios**:

1. **Given** um identificador de pacote-base que o sistema não
   reconhece, **When** uma diferença é pedida contra ele, **Then** a
   resposta é o pacote completo para o alvo pedido, claramente
   identificada como recuperação completa, não como diferença.
2. **Given** um identificador de pacote-base que o sistema já invalidou
   (por mudança de conteúdo, configuração ou versão de contrato),
   **When** uma diferença é pedida contra ele, **Then** a resposta
   também é o pacote completo, pela mesma razão.
3. **Given** uma resposta de recuperação completa, **When** o agente a
   examina, **Then** consegue distinguir claramente que recebeu o
   pacote inteiro (não uma diferença) sem precisar inferir isso pelo
   tamanho da resposta.

---

### User Story 3 - Nunca reutilizar conteúdo desatualizado por engano (Priority: P2)

Como mantenedor confiando no resultado de um Context Pack incremental,
quero ter certeza de que uma diferença nunca é calculada contra uma
base cujo conteúdo de origem, configuração de pedido (política de
ranking, formato, estimador) ou versão de contrato já mudaram desde
que essa base foi gerada, para que a reutilização de cache nunca
introduza uma resposta tecnicamente correta em formato mas
semanticamente desatualizada.

**Why this priority**: Sem essa disciplina de invalidação, a
funcionalidade central (User Story 1) poderia parecer funcionar
perfeitamente e ainda assim entregar, silenciosamente, uma diferença
calculada sobre uma premissa que não é mais verdadeira.

**Independent Test**: Pode ser testado gerando um pacote-base, depois
mudando a configuração do pedido (por exemplo, pedindo com uma
política de ranking diferente) ou a versão do contrato de saída, e
confirmando que uma diferença pedida contra a base antiga não é aceita
como válida — o sistema recorre à recuperação completa (User Story 2)
em vez disso.

**Acceptance Scenarios**:

1. **Given** um pacote-base gerado sob uma configuração de pedido,
   **When** uma diferença é pedida contra essa base usando uma
   configuração diferente (política de ranking, formato ou estimador),
   **Then** o sistema não tenta calcular uma diferença entre
   configurações incompatíveis — recorre à recuperação completa.
2. **Given** um pacote-base gerado antes de uma mudança na versão do
   contrato de saída, **When** uma diferença é pedida contra essa base
   depois da mudança, **Then** o sistema trata a base como inválida e
   recorre à recuperação completa.
3. **Given** um pacote-base ainda válido em todos os outros aspectos,
   **When** o conteúdo de origem de um dos trechos selecionados muda,
   **Then** apenas esse trecho aparece como modificado na diferença —
   o restante do pacote-base continua sendo tratado como válido.

---

### Edge Cases

- O que acontece quando o pacote-base informado corresponde a um alvo
  diferente do que está sendo pedido agora (por exemplo, o consumidor
  confundiu o identificador de duas Specs diferentes)?
- Como o sistema trata um pedido de diferença cujo pacote-base é tão
  antigo que a diferença resultante seria praticamente do mesmo
  tamanho do pacote completo — ainda assim retorna a diferença, ou
  decide sozinho recuperar o pacote completo?
- O que acontece quando duas partes do conteúdo de origem trocam de
  posição na ordenação (por exemplo, por causa de uma mudança de
  relevância), sem que o conteúdo em si tenha mudado — isso aparece
  como uma remoção e uma adição, ou como algo estável?
- Como o sistema se comporta quando o armazenamento local usado para
  guardar pacotes-base anteriores é apagado ou corrompido entre um
  pedido e outro?
- O que acontece quando o mesmo pacote-base é informado por dois
  pedidos concorrentes (por exemplo, dois agentes retomando trabalho
  ao mesmo tempo) — cada um recebe sua própria diferença
  independentemente?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST identificar um Context Pack de forma
  determinística a partir de um conjunto fixo de insumos: o conteúdo de
  origem selecionado, o alvo/tarefa pedido, a política de ranking em
  uso, o formato de saída, e o estimador de tamanho usado — dois
  pedidos com os mesmos insumos MUST produzir o mesmo identificador.
- **FR-002**: O sistema MUST permitir que quem chama informe um
  identificador de pacote-base já recebido anteriormente e peça, em
  vez do pacote completo, apenas as diferenças (trechos adicionados,
  modificados e removidos) desde essa base.
- **FR-003**: Aplicar a diferença retornada sobre o pacote-base exato
  a que ela se refere MUST reproduzir, sem perda de informação, o
  pacote completo que um pedido direto (sem base) teria retornado para
  o mesmo alvo e configuração.
- **FR-004**: Quando o identificador de pacote-base informado não é
  reconhecido, está invalidado, ou não pode ser confirmado pelo
  sistema, o sistema MUST retornar o pacote completo em vez de uma
  diferença — a alegação de posse de quem chama MUST NOT ser aceita
  sem essa verificação.
- **FR-005**: Uma resposta de recuperação completa (FR-004) MUST se
  identificar explicitamente como tal, de forma que quem a recebe não
  precise inferir pelo tamanho ou pela ausência de uma base se recebeu
  uma diferença ou o pacote inteiro.
- **FR-006**: O sistema MUST invalidar um pacote-base sempre que
  qualquer um dos insumos que o identificam (FR-001) mudar — conteúdo
  de origem, configuração do pedido, ou versão do contrato de saída —
  mesmo que apenas parte do conteúdo tenha efetivamente mudado.
- **FR-007**: Uma diferença calculada entre um pacote-base e o estado
  atual MUST refletir apenas os trechos cujo conteúdo de origem
  realmente mudou — um trecho cujo conteúdo permanece idêntico MUST NOT
  aparecer como removido e readicionado apenas por ter mudado de
  posição na ordenação, quando essa mudança de posição não afeta a
  relevância do resultado.
- **FR-008**: O sistema MUST manter a ordenação das partes de um
  pacote que não foram afetadas por uma mudança estável entre pedidos,
  sempre que isso não comprometer a ordenação por relevância já
  estabelecida pela funcionalidade de ranking existente.
- **FR-009**: O sistema MUST oferecer uma forma de medir, por
  integração consumidora, se a reutilização incremental está de fato
  reduzindo o conteúdo repetido enviado — a existência do mecanismo de
  cache, por si só, não é suficiente sem essa medição.
- **FR-010**: A documentação e a resposta do sistema MUST deixar claro
  que este mecanismo de reutilização local é distinto do cache de
  prompt de um provedor de LLM — nenhuma alegação de economia
  financeira MUST ser feita a partir apenas da existência deste
  mecanismo, sem medição específica por integração (FR-009).
- **FR-011**: Todo estado usado para calcular diferenças (identificação
  de pacotes-base, seu conteúdo associado) MUST ser tratável como
  descartável — sua perda ou corrupção MUST NUNCA impedir o sistema de
  produzir uma resposta completa e correta (via recuperação completa,
  FR-004), apenas eliminar a possibilidade de entregar uma diferença
  reduzida naquele pedido específico.

### Key Entities

- **Identificador de Pacote-Base**: o identificador determinístico
  derivado dos insumos de um Context Pack (conteúdo, alvo, política de
  ranking, formato, estimador), usado por quem chama para alegar posse
  de um pacote já recebido.
- **Diferença de Context Pack**: o conjunto de trechos adicionados,
  modificados e removidos entre um pacote-base e o estado atual dos
  mesmos insumos — suficiente, junto com o pacote-base, para
  reconstruir o pacote completo.
- **Gatilho de Invalidação**: a condição (mudança de conteúdo de
  origem, de configuração do pedido, ou de versão do contrato de
  saída) que torna um pacote-base inválido para cálculo de diferença,
  forçando recuperação completa.
- **Medição de Reutilização**: o registro, por integração consumidora,
  de quanto conteúdo repetido deixou de ser reenviado graças a este
  mecanismo — a evidência de que a reutilização está de fato reduzindo
  repetição, não apenas uma capacidade não medida.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Para um pedido repetido ao mesmo alvo com o pacote-base
  anterior informado e sem mudança no conteúdo de origem, o tamanho da
  resposta é mensuravelmente menor que o de um pedido completo
  equivalente, em 100% dos cenários testados de reutilização.
- **SC-002**: Aplicar a diferença retornada sobre seu próprio
  pacote-base reproduz exatamente o pacote completo que um pedido
  direto teria retornado, em 100% dos casos testados.
- **SC-003**: Quando o pacote-base informado é desconhecido,
  invalidado, ou não pode ser confirmado, quem chama sempre recebe uma
  resposta completa e corretamente identificada como tal — 0 casos de
  uma resposta incompleta apresentada sem essa identificação explícita.
- **SC-004**: 0 casos, nos cenários testados, em que uma diferença é
  servida usando insumos que não correspondem mais ao estado atual do
  projeto ou à configuração do pedido atual.
- **SC-005**: Para cada integração participante testada, a redução real
  de conteúdo repetido enviado é mensurável e reportada, não apenas
  presumida a partir da existência do mecanismo.

## Assumptions

- Esta funcionalidade constrói sobre o Context Pack completo e
  versionado (Spec 033) e a precisão de orçamento sobre a saída efetiva
  (Spec 035), já entregues — reutiliza o fingerprint de conteúdo por
  item e a versão de contrato que essas Specs já estabelecem, em vez de
  criar um segundo mecanismo de identidade de conteúdo.
- O armazenamento local de pacotes-base (se implementado como um
  índice ou cache em disco) é descartável e reconstruível a qualquer
  momento — nunca uma fonte autoritativa de estado do projeto,
  consistente com o índice já existente do Context Engine.
- Economia financeira de tokens depende de como cada integração usa,
  por conta própria, o cache de prompt do seu provedor de LLM — esta
  funcionalidade garante apenas que o mecanismo local produz
  diferenças corretas e recuperações completas seguras; não garante
  nem mede economia de custo do provedor diretamente (FR-010).
- Fora do escopo: qualquer alteração no comportamento de cache de
  prompt do provedor de LLM em si; a proposta trata os dois mecanismos
  como deliberadamente independentes.
- A ordenação estável (FR-007, FR-008) é uma otimização best-effort
  para reduzir ruído na diferença — não é uma garantia formal de que a
  ordenação nunca muda; quando uma mudança de ordenação é necessária
  para refletir relevância, ela tem prioridade sobre a estabilidade.
