# Feature Specification: Robust Text Search and Measurable Ranking

**Feature Branch**: `036-text-search-ranking`
**Created**: 2026-09-21
**Status**: Draft
**Input**: User description: "PROP-06 — Busca textual robusta e ranking mensurável: tornar segura e útil a consulta por texto livre e aproveitar o sinal de relevância calculado pelo SQLite (BM25), separando modo de texto livre de modo de expressão avançada, definindo desempate estável, expondo componentes de score em diagnóstico, e versionando mudanças de ranking com evidência de avaliação antes de promovê-las."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Free-text queries never fail on punctuation (Priority: P1)

An agent or user asks for context using an ordinary phrase copied from a task title, an error message, or a requirement description — text that may contain punctuation such as quotes, hyphens, colons, or parentheses. Today this text is passed verbatim into a full-text search expression; certain characters are reserved syntax there and cause the query to fail outright instead of returning results.

**Why this priority**: This is the core reliability problem. A retrieval feature that can throw a syntax error on ordinary input is worse than one with imperfect ranking — it blocks the task entirely instead of degrading gracefully. This must work before ranking quality is worth investing in.

**Independent Test**: Can be fully tested by running free-text queries built from real task titles and requirement text (including quotes, hyphens, colons, and parentheses) against an indexed project and confirming every query returns a result set (possibly empty) rather than an error.

**Acceptance Scenarios**:

1. **Given** an indexed project, **When** a free-text query contains a double quote, hyphen, colon, or parenthesis, **Then** the search completes and returns a normal (possibly empty) result set instead of a syntax error.
2. **Given** an indexed project, **When** a free-text query is plain prose with no special characters, **Then** results are unaffected compared to the current behavior for that same prose.
3. **Given** an indexed project, **When** a free-text query contains terms in Portuguese and English mixed together, **Then** the search still completes and returns relevance-ordered results without error.

---

### User Story 2 - Relevance signal reflects actual computed relevance (Priority: P1)

Today the system computes a real relevance score for each text match but then discards it, substituting a much cruder heuristic (a capped count of term occurrences) when ordering results. This means two matches that the underlying search engine judged very differently in relevance can end up scored identically or in the wrong order once results are combined with other context.

**Why this priority**: This is the "measurable ranking" half of the proposal — without it, "search" is only as good as a keyword-count guess, even though a better signal is already being computed and thrown away. It is independently valuable and independently testable from the query-safety fix in User Story 1.

**Independent Test**: Can be fully tested by issuing a query with several free-text matches of clearly differing relevance and confirming the ordering among same-tier text matches reflects the underlying relevance signal rather than a raw term-occurrence count.

**Acceptance Scenarios**:

1. **Given** two candidate passages matched by the same free-text query, **When** one passage is substantively more relevant to the query terms than the other (as judged by the underlying relevance computation), **Then** the more relevant passage is ordered ahead of the less relevant one among same-priority-tier results.
2. **Given** a set of free-text matches with some scores tied, **When** results are ordered, **Then** the tie-break order is stable and deterministic (same inputs always produce the same output order).
3. **Given** results assembled from multiple priority tiers (e.g., a formally required document plus free-text matches), **When** ordering the combined set, **Then** free-text relevance never promotes a lower-priority-tier result above a higher-priority-tier one.

---

### User Story 3 - Advanced query syntax remains available for power users (Priority: P2)

A user who understands the underlying search engine's query syntax (phrase search, prefix matching, boolean operators) wants to use it deliberately for a precise query, rather than having every special character silently escaped.

**Why this priority**: Useful for advanced/precise retrieval but not required for the core reliability and ranking fixes to deliver value — most queries are ordinary free text. This can ship after or alongside Stories 1 and 2.

**Independent Test**: Can be fully tested by issuing a query explicitly marked as using advanced syntax and confirming it is interpreted using the engine's native query language rather than being escaped as literal text.

**Acceptance Scenarios**:

1. **Given** a query explicitly marked as an advanced-syntax query, **When** it uses phrase, prefix, or boolean operators, **Then** those operators are honored as query syntax rather than treated as literal characters.
2. **Given** a query not marked as advanced, **When** it happens to contain characters that are also advanced-syntax operators, **Then** those characters are treated as literal text, not as operators.

---

### Edge Cases

- What happens when a free-text query is empty or contains only characters with no searchable terms (e.g., only punctuation or only stop-word-length fragments)? The search should return an empty result set, not an error.
- What happens when a free-text query has no matching content anywhere in the index? The search should return an empty result set cleanly.
- What happens when an advanced-syntax query (User Story 3) is malformed (e.g., unbalanced quotes or an invalid operator)? The system should surface a clear, actionable diagnostic rather than a raw underlying engine error, and must not silently fall back to treating it as free text without saying so.
- What happens when relevance scores for two results are exactly equal? The existing deterministic tie-break order (by location within the same tier) must still apply.
- What happens to a proposed ranking-weight change that has not yet been evaluated against a baseline? It must not ship as the default behavior until comparative evidence supports it.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST accept free-text queries containing arbitrary characters (including but not limited to quotes, hyphens, colons, parentheses, and asterisks) without producing a query-syntax error.
- **FR-002**: The system MUST distinguish two query modes: a default free-text mode (special characters treated as literal text) and an explicit advanced mode (special characters interpreted as the underlying search engine's native query syntax).
- **FR-003**: In free-text mode, the system MUST tokenize and neutralize any characters that the underlying search engine would otherwise interpret as reserved syntax, before the query reaches the engine.
- **FR-004**: In advanced mode, a malformed expression MUST produce a clear, distinguishable diagnostic (not a generic or leaked low-level engine error, and not a silent empty result).
- **FR-005**: The system MUST retain the relevance score computed by the underlying search engine for each free-text match and MUST use it (not a separately recomputed term-count heuristic) when ordering free-text matches against each other.
- **FR-006**: The relevance-based ordering of free-text matches MUST NOT override the existing priority-tier ordering — matches in a higher-priority tier are always ordered ahead of matches in a lower-priority tier regardless of relevance score.
- **FR-007**: When two or more results have identical relevance scores within the same tier, the system MUST apply a deterministic, stable tie-break (consistent with the existing tie-break convention) so that repeated identical queries return results in the same order.
- **FR-008**: The system MUST expose the individual components contributing to a result's final score (e.g., tier, underlying relevance value, any bonus weights) in a diagnostic/debug output mode, without requiring this detail in the normal compact output.
- **FR-009**: Any change to ranking weights or scoring formula MUST be versioned, so that which ranking behavior produced a given result set is identifiable.
- **FR-010**: A ranking-weight or scoring-formula change MUST NOT become the default behavior until it has comparative evaluation evidence (per the project's existing evaluation protocol) showing it is not a regression.
- **FR-011**: The system MUST return correct, non-error results for free-text queries containing punctuation-heavy terms, bare identifiers/codes (e.g., `SPEC-014`, `TASK-003`), Portuguese text, English text, and terms with no matching content in the index.
- **FR-012**: The system MUST NOT let free-text relevance score cause a lower-priority-tier candidate to be selected or presented as more relevant than a higher-priority-tier candidate, in either compact or diagnostic output.

### Key Entities

- **Search Query**: The free-text or advanced-syntax input a caller supplies; has a mode (free-text or advanced) and a raw text value.
- **Search Result**: One matched passage with its origin document/section, location, and the relevance value computed for it by the underlying search engine.
- **Score Components**: The individual, named contributors to a candidate's final ordering score (tier, underlying relevance value, relation weight, intent bonus), visible together only in diagnostic output.
- **Ranking Version**: An identifier for the specific scoring formula/weight set in effect, allowing a given result set to be traced back to the behavior that produced it.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of free-text queries built from real task titles, requirement text, and error messages (including punctuation) complete without a query-syntax error, across a representative sample of at least 50 such queries.
- **SC-002**: Among free-text matches within the same priority tier, ordering matches the underlying engine's own relevance ranking in at least 95% of sampled multi-match queries, replacing the current term-count heuristic.
- **SC-003**: Repeated identical queries against an unchanged index produce byte-identical result ordering 100% of the time.
- **SC-004**: A person debugging a surprising result ranking can identify, without reading source code, which score component (tier, relevance, relation weight, intent bonus) drove the ordering, using diagnostic output alone.
- **SC-005**: No ranking-weight change ships as default behavior without an accompanying before/after comparison recorded against the project's evaluation baseline.

## Assumptions

- The underlying search engine (SQLite FTS5) and its BM25 relevance computation remain the retrieval mechanism; this feature changes how that signal is escaped, surfaced, and consumed — not what computes it.
- "Advanced mode" is an explicit, separate query path (e.g., a distinct flag/parameter) rather than a heuristic that guesses intent from characters present in the input — ambiguity between the two modes would undermine the safety guarantee of Story 1.
- The existing priority-tier model (formal dependencies and other structural relations outranking free-text matches) is out of scope to change; this feature only fixes how relevance is computed and surfaced within and below that model.
- Comparative ranking evaluation (SC-005, FR-010) relies on the project's existing end-to-end evaluation capability; this feature does not define a new evaluation harness, only requires evidence be produced before promoting a weight change.
- Diagnostic score-component output is for debugging/tuning by humans or tooling and is not part of the compact, default context output consumed by agents during normal task execution.
