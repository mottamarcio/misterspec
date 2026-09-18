# Feature Specification: Orçamento sobre a saída efetiva de contexto

**Feature Branch**: `035-context-budget-accuracy`
**Created**: 2026-09-18
**Status**: Draft
**Input**: User description: "PROP-05 — Orçamento sobre a saída efetiva de contexto: aproximar o orçamento informado do tamanho realmente entregue ao agente, com comportamento explícito quando o conteúdo obrigatório excede o limite."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Saber exatamente o que o orçamento está medindo (Priority: P1)

Um agente que pede um Context Pack com um orçamento de tokens precisa confiar que o número reportado corresponde ao que foi de fato entregue, e precisa saber qual método de estimativa produziu esse número. Hoje já existe uma interface de estimador (`Estimator`) declarada no código, mas ela nunca é usada — toda contagem de tokens chama diretamente a mesma função de aproximação fixa, sem que a resposta diga qual estimador foi usado nem permita trocar por outro (por exemplo, um tokenizador local mais preciso).

**Why this priority**: Sem isso, o orçamento é uma promessa não verificável — o agente não tem como saber se o número reflete a aproximação padrão ou algo mais preciso, nem como confirmar que a contagem corresponde ao que realmente recebeu.

**Independent Test**: Pedir um Context Pack e confirmar que a resposta identifica explicitamente qual estimador de tokens foi usado, e que o número reportado para o conteúdo corresponde ao tamanho real do texto entregue.

**Acceptance Scenarios**:

1. **Given** um pedido de Context Pack, **When** a resposta é recebida, **Then** ela identifica explicitamente qual estimador de tokens produziu os números reportados.
2. **Given** o mesmo pedido, **When** o número de tokens do conteúdo é comparado ao texto de fato entregue, **Then** o número corresponde a esse texto, não a uma aproximação genérica desconectada do que foi enviado.
3. **Given** nenhum estimador mais preciso configurado, **When** a resposta é recebida, **Then** o sistema usa sua própria aproximação padrão e identifica-a como tal — nunca apresenta uma aproximação como se fosse uma contagem exata sem dizer que é uma aproximação.

---

### User Story 2 - Comportamento previsível quando o conteúdo obrigatório não cabe (Priority: P1)

Hoje, o orçamento pedido pelo agente serve simultaneamente para duas coisas diferentes: o alvo de quanto conteúdo opcional incluir, e o limite contra o qual se decide se o conteúdo obrigatório "excedeu" o orçamento. Isso significa que um agente pedindo um orçamento pequeno para uma exploração leve pode ser informado, de forma confusa, que o conteúdo obrigatório "excedeu" um número que ele nunca pretendeu usar como teto rígido para conteúdo que ele nem controla. É preciso separar um limite flexível (o alvo para conteúdo opcional) de um limite rígido (o teto real acima do qual o conteúdo obrigatório é considerado excessivo), e nunca truncar conteúdo obrigatório silenciosamente — sempre retornar um diagnóstico claro quando ele não couber no limite rígido.

**Why this priority**: Sem essa separação, o sinal de "orçamento excedido" é ambíguo e pode ser tanto um falso alarme quanto, inversamente, mascarar um caso genuinamente crítico — é tão fundamental quanto a User Story 1, por isso tem a mesma prioridade.

**Independent Test**: Pedir um Context Pack com um limite flexível pequeno para um alvo cujo conteúdo obrigatório é maior que esse limite flexível, mas menor que o limite rígido; confirmar que o sistema não relata isso como "excedido" no sentido crítico. Depois, pedir com um limite rígido menor que o conteúdo obrigatório e confirmar que o sistema retorna um diagnóstico claro, sem cortar silenciosamente o conteúdo obrigatório.

**Acceptance Scenarios**:

1. **Given** um conteúdo obrigatório maior que o limite flexível pedido, mas menor que o limite rígido, **When** o Context Pack é montado, **Then** o conteúdo obrigatório é entregue por completo, sem ser relatado como uma condição crítica de excesso.
2. **Given** um conteúdo obrigatório maior que o limite rígido, **When** o Context Pack é montado, **Then** o sistema retorna um diagnóstico explícito descrevendo o excesso, e o conteúdo obrigatório continua sendo entregue por completo — nunca cortado silenciosamente para caber.
3. **Given** nenhum limite rígido explicitamente informado, **When** o Context Pack é montado, **Then** o sistema aplica um limite rígido padrão razoável, nunca tratando a ausência de um limite rígido como "sem limite algum".

---

### User Story 3 - Aproveitar melhor o espaço disponível sem violar prioridades (Priority: P2)

Hoje, ao montar o conteúdo opcional de um Context Pack, o sistema para de incluir itens no primeiro candidato que não caiba no espaço restante, mesmo que um candidato menor, mais adiante na lista (mas do mesmo nível de prioridade), ainda coubesse. Isso desperdiça espaço de orçamento que poderia ser usado por conteúdo relevante. Da mesma forma, uma seção de conteúdo muito grande pode ser excluída inteira por não caber, mesmo que uma parte coerente dela (preservando blocos de código e requisitos como unidades indivisíveis) pudesse ser incluída.

**Why this priority**: É uma melhoria de aproveitamento sobre as User Stories 1 e 2 — só faz sentido depois que a contagem (US1) e os limites (US2) já são confiáveis. Prioridade menor porque não é uma condição de segurança/correção, apenas de eficiência.

**Independent Test**: Montar um Context Pack onde o primeiro candidato opcional após o conteúdo obrigatório não cabe, mas um candidato menor mais adiante na mesma faixa de prioridade cabe; confirmar que esse candidato menor é incluído, e que a exclusão do maior é registrada explicitamente no diagnóstico. Separadamente, montar um Context Pack com uma seção grande demais para caber inteira e confirmar que uma unidade coerente menor dela (não um corte arbitrário no meio de um bloco de código ou requisito) é considerada para inclusão.

**Acceptance Scenarios**:

1. **Given** um candidato opcional que não cabe no espaço restante, seguido por um candidato menor do mesmo nível de prioridade que cabe, **When** o Context Pack é montado, **Then** o candidato menor é incluído, e o candidato maior que foi pulado aparece explicitamente no diagnóstico como excluído.
2. **Given** dois candidatos do mesmo nível de prioridade que ambos coubessem, **When** o Context Pack é montado, **Then** nenhum candidato de prioridade mais baixa é incluído à frente de um candidato de prioridade mais alta só porque é menor — a ordem de prioridade nunca é violada em nome do aproveitamento de espaço.
3. **Given** uma seção de conteúdo grande demais para caber por completo no espaço restante, **When** o Context Pack é montado, **Then** o sistema considera unidades menores e coerentes dessa seção para inclusão, nunca cortando um bloco de código ou um requisito no meio.
4. **Given** uma seção sem nenhuma forma de divisão coerente menor que o espaço disponível, **When** o Context Pack é montado, **Then** ela é excluída por completo, registrada no diagnóstico — nunca incluída parcialmente de forma que quebre seu próprio significado.

---

### Edge Cases

- Um pedido cujo limite flexível é maior que o limite rígido (uma configuração provavelmente equivocada) — o sistema deve ter um comportamento definido e não ambíguo, não um resultado indefinido.
- Um limite rígido igual a zero ou negativo, quando o conteúdo obrigatório existe.
- Uma seção cujo conteúdo é inteiramente um único bloco de código ou um único requisito indivisível maior que qualquer limite disponível.
- Nenhum candidato opcional cabe em nenhuma hipótese — o Context Pack deve ainda ser retornado com apenas o conteúdo obrigatório, sem erro.
- Comparar a saída de duas requisições idênticas para confirmar que o comportamento de preenchimento (inclusive o backfill da User Story 3) é determinístico, não apenas "melhor esforço" variável entre chamadas.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST identificar explicitamente, na resposta, qual estimador de tamanho/tokens foi usado para calcular os números reportados.
- **FR-002**: O sistema MUST permitir que um estimador diferente da aproximação padrão seja usado no lugar dela, sem exigir mudança na forma como o restante do pipeline consome o resultado.
- **FR-003**: Quando nenhum estimador mais preciso está disponível, o sistema MUST usar sua própria aproximação padrão já existente, identificando-a como uma aproximação, nunca como uma contagem exata de um tokenizador real.
- **FR-004**: A contagem de tokens reportada para o conteúdo MUST corresponder ao texto de fato entregue na resposta, não a uma estimativa desconectada do conteúdo real.
- **FR-005**: O sistema MUST distinguir um limite flexível (o alvo usado para decidir quanto conteúdo opcional incluir) de um limite rígido (o teto acima do qual o conteúdo obrigatório é considerado excessivo) — os dois nunca são o mesmo número tratado com dois significados.
- **FR-006**: Quando nenhum limite rígido é explicitamente informado, o sistema MUST aplicar um valor padrão razoável para ele, nunca tratando sua ausência como "sem limite".
- **FR-007**: O sistema MUST NOT truncar conteúdo obrigatório silenciosamente sob nenhuma circunstância — quando o conteúdo obrigatório excede o limite rígido, o sistema MUST retornar um diagnóstico explícito e ainda assim entregar o conteúdo obrigatório por completo.
- **FR-008**: Ao montar o conteúdo opcional, o sistema MUST avaliar candidatos menores mais adiante na mesma faixa de prioridade quando um candidato anterior não cabe no espaço restante, em vez de parar de considerar qualquer conteúdo adicional daquela faixa.
- **FR-009**: A ordem de prioridade entre os níveis já estabelecidos MUST NUNCA ser violada em nome de aproveitar melhor o espaço — um candidato de prioridade mais baixa nunca é incluído à frente de um candidato de prioridade mais alta que ainda não foi decidido.
- **FR-010**: Toda exclusão de um candidato que poderia caber, mas não foi incluído por causa da ordem de avaliação ou de espaço, MUST ser registrada explicitamente no diagnóstico da resposta.
- **FR-011**: O sistema MUST oferecer uma forma de dividir uma seção de conteúdo grande demais em unidades menores e coerentes, preservando um bloco de código ou um requisito como unidades indivisíveis — nunca cortadas no meio.
- **FR-012**: Uma seção sem nenhuma divisão coerente menor que o espaço disponível MUST ser excluída por completo, nunca incluída de forma parcial que quebre seu próprio significado.
- **FR-013**: O comportamento de preenchimento do orçamento (incluindo a avaliação de candidatos menores e a divisão de seções) MUST ser determinístico — a mesma entrada sempre produz a mesma seleção.

### Key Entities

- **Estimator (Estimador)**: o método usado para calcular o tamanho/custo de um texto; identificado explicitamente na resposta, com uma aproximação padrão sempre disponível como alternativa quando nenhum outro está configurado.
- **Soft Limit (Limite flexível)**: o alvo usado para decidir quanto conteúdo opcional incluir — pode ser excedido apenas no sentido de "não há mais espaço para incluir mais nada opcional", nunca no sentido de cortar o que já foi decidido.
- **Hard Limit (Limite rígido)**: o teto acima do qual o conteúdo obrigatório é considerado excessivo, produzindo um diagnóstico explícito — nunca usado para cortar esse conteúdo.
- **Exclusion Record (Registro de exclusão)**: uma entrada de diagnóstico nomeando um candidato ou uma seção que poderia ter sido incluído, mas não foi, e o motivo.
- **Coherent Unit (Unidade coerente)**: uma parte de uma seção grande que pode ser incluída isoladamente sem quebrar seu próprio significado — um bloco de código ou um requisito completo, nunca um fragmento arbitrário.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em 100% dos pedidos testados, a contagem de tokens reportada para o conteúdo corresponde ao tamanho real do texto entregue, e a resposta identifica qual estimador a produziu.
- **SC-002**: Em 100% dos casos testados, um conteúdo obrigatório menor que o limite rígido nunca é relatado como uma condição crítica de excesso, mesmo quando maior que o limite flexível.
- **SC-003**: Em 100% dos casos testados, um conteúdo obrigatório que excede o limite rígido é entregue por completo, acompanhado de um diagnóstico explícito — nunca cortado.
- **SC-004**: Em cenários onde um candidato menor mais adiante na mesma prioridade poderia caber após um candidato maior ser pulado, esse candidato menor é incluído em 100% dos casos testados, sem violar a ordem de prioridade.
- **SC-005**: Repetir o mesmo pedido produz exatamente a mesma seleção de conteúdo em 100% dos casos testados.

## Assumptions

- Esta especificação depende do contrato de saída completo já estabelecido (033-context-pack-output-contract) — os números de orçamento e diagnóstico aqui descritos se aplicam à saída realmente entregue naquele contrato, não a uma estimativa separada dele.
- Um estimador mais preciso que a aproximação padrão (por exemplo, um tokenizador local específico de um provedor) é uma capacidade opcional que este recurso habilita, não uma obrigação de implementar um tokenizador real específico.
- A mudança na política de preenchimento (avaliar candidatos menores após pular um maior, dividir seções grandes) é a própria entrega desta funcionalidade; uma avaliação comparativa formal de qualidade/eficiência entre essa nova política e a política anterior (parar no primeiro candidato que não cabe) é responsabilidade de uma iniciativa de avaliação separada, e não é um bloqueio para esta especificação entregar a nova política corretamente.
- O orçamento tratado aqui é o do Context Pack de uma única chamada, não o orçamento da sessão inteira do agente — um Context Pack pequeno não impede o agente de fazer novas chamadas depois.
- Um limite rígido padrão razoável, quando nenhum é informado, é definido como um múltiplo fixo do limite flexível padrão já existente — grande o suficiente para não disparar diagnósticos de excesso em cenários comuns, mas finito.
