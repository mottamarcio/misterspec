# Feature Specification: Context Pack completo e contrato de saída versionado

**Feature Branch**: `033-context-pack-output-contract`
**Created**: 2026-09-18
**Status**: Draft
**Input**: User description: "PROP-03 — Context Pack completo e contrato de saída versionado: entregar ao agente os trechos selecionados, evitando que ele precise redescobrir seu conteúdo. A saída padrão atual apresenta metadados; o texto aparece apenas com --render."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Receber o conteúdo selecionado sem releitura obrigatória (Priority: P1)

Um agente pede um Context Pack para uma Spec (via `internal context`) e precisa do texto real dos trechos selecionados para trabalhar. Hoje, a saída padrão só traz metadados (caminho, heading, score, tokens) — nunca o conteúdo — e mesmo com `--render` o texto só existe embutido em um único bloco de Markdown, nunca como campo estruturado por item. As Skills (`mister-plan`, `mister-tasks`, `mister-analyze`, `mister-implement`, `mister-wrap-up`) já documentam esse padrão: pedem o Context Pack, tratam-no como "ponto de partida", e depois releem os arquivos originais separadamente para obter o conteúdo de fato.

**Why this priority**: É a lacuna central que a proposta existe para fechar — sem o conteúdo entregue, o Context Pack não cumpre sua promessa de reduzir leituras adicionais; o agente sempre paga o custo de uma segunda rodada de leitura mecânica.

**Independent Test**: Pedir um Context Pack em um modo que inclua conteúdo completo para uma Spec com trechos selecionados; confirmar que cada item retornado já contém o texto selecionado, e que nenhuma leitura adicional do arquivo original é necessária para obter esse texto.

**Acceptance Scenarios**:

1. **Given** uma Spec com trechos elegíveis para seleção, **When** um agente pede o Context Pack no modo de conteúdo completo, **Then** cada item retornado inclui o texto selecionado diretamente, sem exigir uma chamada adicional para lê-lo.
2. **Given** o mesmo pedido, **When** o agente compara o texto retornado com o conteúdo real do arquivo de origem, **Then** o texto retornado corresponde exatamente ao trecho selecionado naquele arquivo.
3. **Given** um pedido no modo leve (somente metadados, sem conteúdo), **When** a resposta é recebida, **Then** o sistema deixa explícito que aquele modo não inclui conteúdo, para que o agente não assuma erroneamente que o texto já foi entregue.

---

### User Story 2 - Localizar cada trecho no arquivo de origem (Priority: P1)

Um agente que recebe um item do Context Pack precisa conseguir localizá-lo exatamente no arquivo original — para citar a fonte, para abrir o arquivo na posição certa, ou para comparar com uma versão mais recente. Hoje, a numeração de linha usada internamente por um dos componentes de seleção é relativa ao corpo do documento (depois do frontmatter), não absoluta ao arquivo — a mesma linha reportada não corresponde à linha real do arquivo quando o consumidor não sabe descontar o frontmatter.

**Why this priority**: Uma localização incorreta é pior do que nenhuma localização — o agente confia nela e erra. Isso é tão fundamental quanto entregar o conteúdo (User Story 1), por isso tem a mesma prioridade.

**Independent Test**: Pedir um Context Pack para uma Spec cujo arquivo de origem tem frontmatter; confirmar que a localização reportada para cada item, quando usada para abrir o arquivo original naquela posição, aponta exatamente para o trecho selecionado — não alguma linha antes, deslocada pelo tamanho do frontmatter.

**Acceptance Scenarios**:

1. **Given** um arquivo de origem com frontmatter seguido de corpo, **When** um item selecionado desse corpo é retornado no Context Pack, **Then** a localização reportada é absoluta ao arquivo inteiro, não relativa ao corpo pós-frontmatter.
2. **Given** um arquivo de origem sem frontmatter, **When** um item é selecionado, **Then** a localização reportada continua correta e consistente com o caso com frontmatter.
3. **Given** qualquer item retornado, **When** o agente pede a origem (caminho), o heading e a razão de seleção desse item, **Then** todos os três acompanham o item, junto com a localização e um fingerprint do conteúdo.

---

### User Story 3 - Consumir um contrato de saída versionado e sem duplicação (Priority: P2)

Um agente (ou uma integração futura) precisa saber, de forma explícita, qual formato de saída está recebendo, e não deve receber o mesmo conteúdo selecionado duplicado em duas representações diferentes na mesma resposta (por exemplo, uma vez como campo estruturado e de novo embutido em um bloco de Markdown), o que infla o tamanho da resposta sem agregar informação nova. Consumidores existentes que já dependem do formato de saída atual (metadados por padrão) não podem quebrar sem aviso.

**Why this priority**: É uma melhoria de precisão e de compatibilidade sobre as User Stories 1 e 2 — só faz sentido definir o contrato formalmente depois que se sabe o que ele precisa carregar (conteúdo e localização). Prioridade menor porque não bloqueia o valor central das duas primeiras histórias.

**Independent Test**: Pedir Context Packs nos diferentes modos de saída explícitos e confirmar que cada resposta identifica de forma legível qual modo/versão de contrato foi usado, que nenhum modo duplica o mesmo conteúdo em duas formas na mesma resposta, e que uma chamada no formato usado antes desta funcionalidade continua funcionando ou aponta claramente para o caminho de migração.

**Acceptance Scenarios**:

1. **Given** um pedido de Context Pack, **When** a resposta é recebida, **Then** ela identifica explicitamente a versão do contrato de saída usada.
2. **Given** um pedido no modo de pacote completo, **When** a resposta é recebida, **Then** o conteúdo de cada item aparece uma única vez na resposta — nunca duplicado como campo estruturado e também embutido em uma renderização de texto corrido na mesma chamada.
3. **Given** uma integração já existente que depende do formato de saída anterior a esta funcionalidade, **When** essa integração continua chamando `internal context` do mesmo modo que antes, **Then** ela continua funcionando sem modificação, ou recebe uma indicação clara e documentada de como migrar.

---

### Edge Cases

- Um item selecionado vem de uma Tarefa (`TASK-NNN`) dentro de um arquivo de tarefas compartilhado por várias tarefas — a localização reportada deve continuar identificando essa tarefa especificamente dentro do arquivo, não apenas o arquivo inteiro.
- Um pedido no modo de pacote completo para uma seleção muito grande — o comportamento de o que fazer quando o conteúdo obrigatório não cabe no orçamento pedido é tratado por uma iniciativa relacionada e está fora do escopo desta especificação; aqui importa apenas que o tamanho relatado corresponda ao que foi de fato entregue.
- Um arquivo de origem muda entre o momento em que o Context Pack foi gerado e o momento em que o agente volta a consultá-lo — o fingerprint de cada item deve permitir detectar essa divergência.
- Um pedido no modo leve (somente metadados) para uma seleção vazia — a resposta deve deixar claro que não há itens, não confundir "vazio" com "modo sem conteúdo".

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST oferecer um modo de saída que entrega o conteúdo textual completo de cada trecho selecionado junto com a resposta, sem exigir uma leitura adicional do arquivo de origem para obter esse texto.
- **FR-002**: Cada item retornado nesse modo MUST incluir: o texto selecionado, o caminho de origem, o heading (quando aplicável), a localização, um fingerprint do conteúdo, e o motivo pelo qual foi selecionado.
- **FR-003**: A localização reportada para qualquer item MUST ser absoluta ao arquivo de origem inteiro, nunca relativa a uma representação intermediária (como o corpo do documento após o frontmatter).
- **FR-004**: O sistema MUST oferecer modos de saída explícitos e distintos — ao menos um modo leve (somente metadados) e um modo de pacote completo (com conteúdo) — permitindo que quem chama escolha o que precisa.
- **FR-005**: Por padrão, o sistema MUST NOT duplicar o mesmo conteúdo selecionado em mais de uma representação (por exemplo, campo estruturado e bloco de texto corrido) dentro da mesma resposta.
- **FR-006**: A resposta MUST identificar explicitamente qual versão do contrato de saída foi usada.
- **FR-007**: As métricas de tamanho/consumo reportadas na resposta MUST refletir o tamanho real do que foi efetivamente entregue no modo pedido, não apenas o conteúdo bruto dos trechos selecionados.
- **FR-008**: Uma chamada que já dependia do formato de saída existente antes desta funcionalidade MUST continuar funcionando sem modificação, ou o sistema MUST fornecer um caminho de migração explícito e documentado para o novo formato.
- **FR-009**: Um item derivado de uma Tarefa dentro de um arquivo de tarefas compartilhado MUST manter uma localização que identifique essa Tarefa especificamente, não apenas o arquivo inteiro.
- **FR-010**: O sistema MUST permitir que um agente obtenha, em uma única chamada, tudo o que hoje exige um pedido de Context Pack seguido de releitura manual dos arquivos de origem para obter o conteúdo.

### Key Entities

- **Context Pack Item**: uma unidade de conteúdo selecionada, com texto, origem, heading, localização absoluta, fingerprint e motivo de seleção.
- **Output Mode (Modo de saída)**: o formato nomeado e explícito que quem chama escolhe (por exemplo, leve/metadados ou pacote completo/conteúdo) — modos diferentes produzem formas diferentes, nunca uma mistura implícita.
- **Context Pack Envelope**: o invólucro versionado da resposta — identifica a versão do contrato, o alvo, a intenção, o orçamento pedido e os diagnósticos, independentemente do modo de saída escolhido.
- **Diagnostics (Diagnósticos)**: as métricas de tamanho/consumo que acompanham a resposta, refletindo o que foi de fato entregue no modo pedido.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Um agente que pede o Context Pack no modo de conteúdo completo consegue prosseguir sem nenhuma leitura adicional de arquivo para obter o conteúdo já incluído, em 100% dos casos testados.
- **SC-002**: Todo item de um Context Pack pode ser localizado exatamente no arquivo de origem (arquivo + posição), verificado em 100% de uma amostra de itens testados.
- **SC-003**: Uma integração construída contra o formato de saída anterior a esta funcionalidade continua funcionando sem modificação, ou recebe um caminho de atualização documentado, cobrindo 100% das Skills atualmente distribuídas.
- **SC-004**: O tamanho/consumo relatado para um modo de saída pedido corresponde ao tamanho real do que foi retornado, dentro de uma margem pequena definida, em todo caso testado.
- **SC-005**: Nenhuma resposta no modo padrão contém o mesmo conteúdo selecionado duplicado em duas representações diferentes, em 100% dos pedidos testados.

## Assumptions

- O comportamento de o que fazer quando o conteúdo obrigatório não cabe no orçamento pedido (truncar, excluir, sinalizar) é tratado por uma iniciativa relacionada e permanece fora do escopo desta especificação; aqui interessa apenas que o tamanho relatado corresponda ao que foi entregue no modo escolhido.
- "Fingerprint" aqui significa um identificador determinístico do conteúdo de um item, suficiente para detectar se a fonte mudou — não implica nenhum algoritmo específico.
- Os modos de saída leve e de pacote completo são os dois modos mínimos necessários; um modo adicional de texto corrido (Markdown) pode continuar existindo, mas não deve duplicar o conteúdo já presente no modo de pacote completo na mesma resposta.
- A compatibilidade retroativa (FR-008) cobre as Skills e integrações já distribuídas no momento desta especificação; uma integração hipotética futura não é considerada aqui.
- Esta especificação depende do modelo de identidade de tarefas já estabelecido (031-canonical-task-identity) para que um item derivado de uma Tarefa (FR-009) possa ser localizado de forma inequívoca dentro de um arquivo de tarefas compartilhado.
