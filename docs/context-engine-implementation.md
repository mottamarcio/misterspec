# MisterSpec Context Engine — Implementation Specification

## 1. Purpose

This document specifies the implementation of the MisterSpec Context Engine: a local-first, deterministic-assisted context retrieval layer that reduces the amount of project context a coding agent must load while preserving MisterSpec's existing architectural boundaries.

The Context Engine is not a replacement for MisterSpec's artifact model, Skills, filesystem state, or the coding agent's semantic judgment. It is a derived retrieval layer over the existing project knowledge graph.

Its primary goal is:

> Find and assemble the smallest sufficient project context for the current agent operation.

The intended outcome is lower prompt/context token usage, less repeated repository exploration, faster agent startup for each operation, and more consistent use of project knowledge, decisions, learnings, and structural relationships.

---

## 2. Architectural Principles

The implementation MUST preserve the following MisterSpec principles.

### 2.1 Filesystem remains authoritative

Markdown artifacts under the configured MisterSpec directories remain the source of truth.

The Context Engine MUST NOT make SQLite, FTS indexes, embeddings, or any other derived storage authoritative.

The model is:

```text
Markdown / filesystem   = authoritative state
Git                     = historical state
SQLite                  = disposable derived index
Context Pack            = disposable derived view
```

Deleting the Context Engine database MUST NOT destroy project information. The database MUST be fully reconstructible from the filesystem.

### 2.2 Agent owns semantic judgment

MisterSpec may deterministically resolve IDs, traverse explicit relationships, parse wikilinks, rank retrieval candidates, enforce budgets, and assemble context.

The coding agent remains responsible for interpreting meaning and deciding whether additional context is necessary.

```text
If meaning is required:
    Agent decides.

If structure is sufficient:
    MisterSpec computes.

If context is too large:
    MisterSpec retrieves and ranks.
    Agent judges.
```

### 2.3 Reuse the existing artifact model

The Context Engine MUST NOT create an independent interpretation of MisterSpec project structure.

Existing packages and operations such as `ids`, `artifacts`, `project`, `operations.Resolve`, `operations.Inspect`, parent/children traversal, validation, and fingerprinting MUST be reused where applicable.

Do not introduce a second filesystem scanner that independently decides what constitutes a Program, Feature, Spec, Knowledge artifact, Learning, Task, or related entity.

### 2.4 Local-first

The MVP MUST work locally and offline after installation.

No remote vector database, embedding service, hosted search system, or required network service may be introduced.

### 2.5 Minimal configuration

The first implementation SHOULD require no new user configuration.

Context budgets, retrieval limits, index paths, and ranking defaults should initially be internal defaults. Configuration options should only be introduced after dogfooding demonstrates a real need.

---

## 3. Scope

### 3.1 MVP includes

The MVP MUST implement:

- artifact wikilinks;
- wikilink parsing and validation;
- formal and semantic artifact references;
- backlinks;
- Markdown section-aware chunking;
- disposable SQLite storage;
- SQLite FTS5 full-text retrieval;
- incremental index synchronization;
- graph-assisted retrieval;
- BM25 textual retrieval;
- intent-aware ranking;
- token estimation;
- context budgeting;
- deterministic context-pack assembly;
- internal CLI operations;
- retrieval diagnostics and token-reduction metrics;
- tests for each layer.

### 3.2 MVP explicitly excludes

The MVP MUST NOT include:

- embeddings;
- vector databases;
- external search services;
- graph databases;
- DuckDB;
- source-code indexing;
- automatic semantic mutation of artifacts;
- automatic promotion of Learnings into Knowledge or Constitution;
- automatic insertion of arbitrary wikilinks by the Go binary;
- an Obsidian-like graphical UI;
- a visual graph explorer;
- editor/plugin functionality.

These may be reconsidered only after measuring the MVP.

---

## 4. High-Level Architecture

```text
                       Coding Agent / Skill
                               │
                               │ operation + target + intent
                               ▼
                    ┌──────────────────────┐
                    │    Context Engine    │
                    └──────────┬───────────┘
                               │
             ┌─────────────────┼──────────────────┐
             │                 │                  │
             ▼                 ▼                  ▼
      Structural Context   Artifact Graph     Text Retrieval
             │                 │                  │
      existing operations  formal links       SQLite FTS5
             │             + wikilinks           BM25
             │             + backlinks            │
             └─────────────────┼──────────────────┘
                               ▼
                         Candidate Ranker
                               │
                               ▼
                         Context Budgeter
                               │
                               ▼
                          Context Pack
                               │
                               ▼
                         Coding Agent
```

The Context Engine is a consumer of MisterSpec's existing deterministic model. It does not replace it.

---

## 5. Artifact Graph Foundation

Before implementing SQLite retrieval, MisterSpec should gain an explicit artifact graph.

The graph has two classes of relationships.

### 5.1 Formal relationships

Formal relationships are authoritative structural relationships already represented by MisterSpec artifact metadata or structure.

Examples include:

- `parent`;
- `depends_on`;
- `supersedes`;
- `for`;
- Program → Feature containment;
- Feature → Spec containment;
- Spec → Plan/Tasks/Validation association;
- Task → requirement references where structurally available.

These relationships retain their existing semantics.

### 5.2 Semantic/navigation relationships

Wikilinks introduce explicit semantic relationships between artifacts.

Examples:

```markdown
## Relevant Knowledge

- [[KNOW-003|Authentication Model]]
- [[KNOW-008|Security Constraints]]

## Related Specs

- [[SPEC-011]]
- [[SPEC-014|Refresh Token Rotation]]

## Related Learnings

- [[LRN-008]]
```

Wikilinks are navigation/retrieval signals. They do not replace formal metadata.

For example, `depends_on: [SPEC-011]` has stronger structural semantics than merely writing `[[SPEC-011]]` in a prose section.

---

## 6. Wikilink Specification

### 6.1 Supported syntax

The MVP MUST support:

```text
[[SPEC-014]]
[[SPEC-014|Refresh Token Rotation]]
```

Also supported, since 040-stable-section-anchors — an optional explicit
anchor qualifying the target down to one specific Section:

```text
[[KNOW-003#retry-policy]]
[[KNOW-003#retry-policy|Política de retries]]
```

The target MUST be a MisterSpec entity ID.

The alias is presentation-only.

Titles and filesystem paths MUST NOT be accepted as canonical wikilink targets in the MVP.

This guarantees stable links when titles or paths change.

### 6.2 Parsed representation

Add a representation similar to:

```go
type WikiLink struct {
    Target string
    Alias  string
    Line   int
}
```

The exact package placement should follow the existing `internal/artifacts` organization. Prefer keeping MisterSpec-specific Markdown parsing near the artifact model rather than introducing a generic Markdown framework.

Suggested files:

```text
internal/artifacts/
├── document.go
├── markdown.go
├── wikilink.go
├── wikilink_test.go
└── ... existing files
```

### 6.3 Parsing API

Provide a small deterministic API such as:

```go
func ExtractWikiLinks(body []byte) ([]WikiLink, error)
```

The parser MUST:

- preserve source line numbers;
- support aliases;
- return links in document order;
- avoid interpreting malformed syntax as a valid link;
- avoid resolving targets during lexical parsing.

Resolution is a separate operation.

### 6.4 Validation

Every parsed wikilink target MUST be resolvable through the existing MisterSpec ID resolution rules.

Reuse `operations.Resolve` or the appropriate underlying deterministic primitive rather than duplicating ID/path rules.

Validation should be able to report conditions such as:

```text
broken_wikilink
ambiguous_wikilink
invalid_wikilink
unknown_anchor
duplicate_anchor
```

`unknown_anchor` (040-stable-section-anchors) marks an anchor-qualified
link whose target artifact exists but declares no Section with that
anchor — kept distinct from `broken_wikilink`, which means the target
artifact itself doesn't exist. `duplicate_anchor` marks two Sections in
the same artifact declaring the same explicit anchor, independent of
whether anything references it.

A link such as:

```text
[[SPEC-999]]
```

must be detected if no corresponding entity exists.

A link must also respect the project's configured ID-width and normal ID rules.

### 6.5 Wikilinks are not mandatory everywhere

Do not mechanically add wikilinks to every artifact or every mention of an ID.

Links should represent useful navigation or semantic relationships. Excessive links reduce graph quality and retrieval signal.

The Go binary should parse and validate links, but semantic link creation remains an agent/human responsibility.

### 6.6 Chunk-level provenance (038-wikilink-chunk-provenance)

Beyond §6.2's own `WikiLink{Target, Alias, Line}`, every wikilink-based
reference now also carries its own enclosing section and file-absolute
line through `operations.ReferenceEntry`/`BacklinkEntry` into the
Context Pack response, so a retrieved item can be explained by the
exact reference occurrence that justified it — not merely "some
artifact references this one." The `internal context` command exposes
this as an opt-in `--provenance` field, and a still-experimental,
off-by-default `--prefer-section` scoring capability can score a
reference written in a Requirements-bearing section above one written
elsewhere, gated behind the evaluation harness's own evidence-based
promotion process (see §30 Phase 9's own note on this) before it can
ever become default ordering. See `specs/038-wikilink-chunk-
provenance/` for the full contract.

### 6.7 Stable section anchors (040-stable-section-anchors)

A heading MAY declare an explicit, stable anchor as a trailing
`{#slug}` suffix — e.g. `## Retry Policy {#retry-policy}` — independent
of the heading's own title text, so a later title rewrite never breaks
a reference to it. §6.1's `[[ID#anchor]]`/`[[ID#anchor|Alias]]` syntax
addresses that Section specifically: retrieval (`internal context`)
returns exactly that Section's own content plus a `heading_path`
breadcrumb (its ancestor headings' titles, outermost first) — never the
whole target artifact, and never an embed/transclusion of the
referenced content into the referencing artifact. Every existing
non-anchor wikilink, heading, and Chunk is completely unaffected — the
anchor-aware code paths only ever activate when an anchor is actually
present. See `specs/040-stable-section-anchors/` for the full contract.

---

## 7. Reference Operations

Implement graph-oriented deterministic operations before building the retrieval database.

### 7.1 References

Provide an operation conceptually equivalent to:

```text
misterspec internal references SPEC-014
```

It should expose both formal and semantic outgoing relationships.

Example logical result:

```json
{
  "ok": true,
  "target": "SPEC-014",
  "references": {
    "formal": [
      {
        "relation": "parent",
        "target": "FEAT-004"
      },
      {
        "relation": "depends_on",
        "target": "SPEC-011"
      }
    ],
    "semantic": [
      {
        "relation": "wikilink",
        "target": "KNOW-003"
      },
      {
        "relation": "wikilink",
        "target": "LRN-008"
      }
    ]
  }
}
```

Exact JSON should follow existing MisterSpec internal-command conventions.

### 7.2 Backlinks

Provide:

```text
misterspec internal backlinks SPEC-014
```

Backlinks identify artifacts containing incoming formal or semantic references to the target.

The implementation may initially scan through the existing artifact model. Once the disposable index exists, the operation may use it as an optimization, but correctness MUST NOT depend on the cache existing.

### 7.3 Relationship strength

The Context Engine must preserve the distinction between relationship types.

Suggested conceptual priority:

```text
target                         highest
formal direct relationship     very high
explicit wikilink              high
backlink                        medium-high
second-hop graph relationship  medium
text-only retrieval            variable
```

Do not flatten all relationships into an untyped graph.

---

## 8. Markdown Document Model and Chunking

The existing metadata parser should remain focused on frontmatter. Do not turn metadata parsing into a full Markdown parser.

Introduce a separate document/body representation.

Example:

```go
type Document struct {
    Metadata Metadata
    Title    string
    Sections []Section
    Links    []WikiLink
}

type Section struct {
    Heading string
    Level   int
    Body    string
    Line    int
}
```

Exact field names may differ, but responsibilities should remain separated.

### 8.1 Section-aware chunking

Do not split Markdown into arbitrary fixed-size windows first.

MisterSpec artifacts already have meaningful structure. Use headings as natural semantic boundaries.

Example:

```text
SPEC-014/spec.md

# Refresh Token Rotation

## Intent
...

## Requirements
### R1
...
### R2
...

## Acceptance Scenarios
...

## Constraints
...
```

Natural chunks should correspond to sections/subsections such as Intent, individual requirements when useful, Acceptance Scenarios, Constraints, Known Facts, Unknowns, Evidence, and similar artifact-specific sections.

Large sections MAY be subdivided if they exceed practical retrieval size, but heading structure should remain attached as provenance.

### 8.2 Chunk provenance

Every chunk MUST preserve enough provenance to explain where it came from.

At minimum:

```go
type Chunk struct {
    ArtifactID string
    Path       string
    Heading    string
    Content    string
    StartLine  int
    EndLine    int
    Tokens     int
}
```

This provenance is required for diagnostics, ranking explanations, and future context rendering.

---

## 9. Token Estimation

Introduce a lightweight token-estimation abstraction.

Suggested interface:

```go
type Estimator interface {
    Estimate(text string) int
}
```

The MVP does not require an exact provider tokenizer.

A deterministic approximation is acceptable as long as it is documented and consistently used for budgeting and metrics.

Keep the abstraction replaceable so provider-specific tokenizers can be added later without changing retrieval logic.

---

## 10. Disposable SQLite Index

Use SQLite, not DuckDB, for the MVP.

The workload consists primarily of:

- lookup by ID/path;
- incremental small writes;
- graph-edge lookups;
- backlinks;
- full-text search;
- BM25 ranking;
- retrieving a small top-N result set.

SQLite is a better fit for this workload and provides FTS5 directly.

### 10.1 Ownership

The index exists only to support context retrieval.

Prefer package ownership such as:

```text
internal/context/
├── engine.go
├── request.go
├── result.go
├── collector.go
├── ranker.go
├── budget.go
├── render.go
└── index/
    ├── store.go
    ├── sqlite.go
    ├── schema.go
    ├── sync.go
    ├── search.go
    └── chunk.go
```

Avoid introducing a broad generic indexing subsystem unless a second concrete consumer appears.

### 10.2 Database location

Use a disposable cache location under `.misterspec`, for example:

```text
.misterspec/cache/context.db
```

The exact location may follow existing project conventions, but it MUST be clearly non-authoritative and safe to delete.

Generated cache data SHOULD NOT be committed to Git.

### 10.3 Suggested schema

The implementation may evolve the schema, but the logical model should cover documents, chunks, links, and FTS.

Example:

```sql
CREATE TABLE documents (
    id INTEGER PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    artifact_id TEXT,
    artifact_type TEXT,
    title TEXT,
    fingerprint TEXT NOT NULL,
    modified_at INTEGER NOT NULL
);

CREATE TABLE chunks (
    id INTEGER PRIMARY KEY,
    document_id INTEGER NOT NULL,
    heading TEXT,
    content TEXT NOT NULL,
    start_line INTEGER,
    end_line INTEGER,
    token_estimate INTEGER NOT NULL,
    FOREIGN KEY(document_id) REFERENCES documents(id)
);

CREATE TABLE links (
    id INTEGER PRIMARY KEY,
    source_artifact_id TEXT NOT NULL,
    target_artifact_id TEXT NOT NULL,
    relation TEXT NOT NULL,
    source_path TEXT,
    source_line INTEGER
);
```

Add an FTS5 table for searchable chunk content. The implementation may use external-content FTS5 or a simpler synchronized table depending on complexity.

Conceptually:

```sql
CREATE VIRTUAL TABLE chunks_fts USING fts5(
    heading,
    content
);
```

The implementation MUST ensure FTS rows remain synchronized with chunks.

### 10.4 Store abstraction

Do not expose SQL throughout `internal/context`.

Use a small interface or concrete store API that represents retrieval operations.

For example:

```go
type Store interface {
    Sync(...) error
    Search(query string, limit int) ([]SearchResult, error)
    Outgoing(id string) ([]Link, error)
    Incoming(id string) ([]Link, error)
}
```

Avoid over-generalizing this interface.

---

## 11. Incremental Synchronization

The index MUST support cheap incremental synchronization.

Reuse MisterSpec's existing fingerprint/SHA-256 concepts rather than inventing a separate change-detection strategy.

Conceptual algorithm:

```text
scan authoritative artifacts
        │
        ▼
calculate/retrieve fingerprint
        │
        ├── new       → parse + chunk + index
        ├── changed   → delete old derived rows + reindex
        ├── unchanged → skip
        └── deleted   → remove derived rows
```

Synchronization MUST be deterministic and idempotent.

A failed synchronization MUST NOT corrupt authoritative artifacts.

The database can always be discarded and rebuilt.

---

## 12. What Gets Indexed

The first release should index MisterSpec knowledge/artifact state only.

Index:

```text
Constitution
Knowledge
Learnings
Programs
Features
Specs
Plans
Tasks
Validation artifacts
```

Do NOT index arbitrary repository source code in the MVP:

```text
*.go
*.ts
*.js
*.py
...
```

Plans and Tasks may contain source paths. The Context Engine can return those paths as useful implementation hints, while the coding agent continues to inspect source code using its normal repository tools.

This keeps responsibilities clear:

```text
Context Engine → project knowledge and execution context
Coding Agent   → source-code exploration and code modification
```

Source indexing should only be considered later if measurements show it provides meaningful value.

---

## 13. Context Request Model

The Context Engine should operate from an explicit request.

Suggested shape:

```go
type Intent string

const (
    IntentPlanning       Intent = "planning"
    IntentTasks          Intent = "tasks"
    IntentImplementation Intent = "implementation"
    IntentValidation     Intent = "validation"
    IntentAnalysis       Intent = "analysis"
)

type Request struct {
    Target string
    Task   string
    Intent Intent
    Query  string
    Budget int
}
```

Exact intents should match actual MisterSpec Skill terminology.

The target artifact is required. Task and query may be optional depending on intent.

---

## 14. Context Collection Tiers

Context candidates MUST be grouped by semantic priority rather than treated as one flat search result list.

Recommended tiers:

### Tier 0 — mandatory project rules

Relevant Constitution content.

### Tier 1 — target

The target artifact and directly required associated artifacts.

For implementation this may include the current Task and requirement(s) it serves.

### Tier 2 — formal structural context

Examples:

- parent;
- direct dependencies;
- associated Plan;
- associated Tasks;
- required completed Task dependencies;
- explicit formal references.

### Tier 3 — explicit semantic graph

Direct wikilinks and high-value backlinks.

### Tier 4 — text retrieval

FTS5/BM25 results from Knowledge, Learnings, and other eligible artifact chunks.

### Tier 5 — wider graph expansion

Second-hop graph relationships only when budget remains and they have meaningful ranking value.

The MVP SHOULD normally avoid traversing more than two graph hops.

---

## 15. Intent-Aware Retrieval

Different operations require different context priorities.

### 15.1 Planning

Prefer:

```text
Constitution
Target Spec
parent Feature/Program when useful
dependencies
relevant Knowledge
relevant Learnings
existing implementation evidence/path hints
```

### 15.2 Task creation

Prefer:

```text
Constitution
Spec
Plan
requirements
acceptance scenarios
dependencies
relevant Knowledge
```

### 15.3 Implementation

Prefer:

```text
Constitution rules relevant to implementation
current Spec
current Task
requirement(s) served by Task
Plan sections relevant to Task
completed Task dependencies and evidence
explicit Knowledge/Learning links
relevant Knowledge/Learnings from retrieval
source paths named by Task/Plan
```

### 15.4 Validation / analysis

Prefer:

```text
Spec
Plan
Tasks
acceptance scenarios
validation/evidence
relevant implementation evidence
relevant Learnings
```

Intent should influence ranking, not change authoritative project semantics.

---

## 16. Retrieval Pipeline

Recommended retrieval order:

```text
1. Resolve target
2. Load mandatory context
3. Traverse direct formal relationships
4. Traverse direct wikilinks
5. Inspect backlinks
6. Build textual query
7. Run FTS5/BM25
8. Optionally expand high-value second-hop graph links
9. Deduplicate candidates
10. Rank candidates
11. Apply token budget
12. Render structured Context Pack
```

This ordering deliberately uses explicit project structure before probabilistic/textual similarity.

Explicit relationships are stronger signals than lexical coincidence.

---

## 17. Ranking

Do not rely exclusively on BM25 score.

Ranking should combine deterministic relationship strength, artifact type, intent, and text relevance.

A simple initial model is preferable to a complex learned model.

Conceptually:

```text
score =
    relationship_weight
  + intent_weight
  + artifact_type_weight
  + text_relevance_weight
```

Illustrative relative weights:

```text
target                         100
formal direct relationship      90
required task dependency        90
explicit wikilink               80
high-value backlink             65
second-hop graph relation       45
BM25 contribution             0..40
```

These values are examples, not immutable requirements. Tests should focus on ordering guarantees rather than hard-coding arbitrary numbers unnecessarily.

### 17.1 Important ranking invariant

A vaguely matching text chunk MUST NOT displace mandatory target context merely because its BM25 score is high.

Mandatory and structural tiers should be budgeted before optional textual retrieval.

---

## 18. Deduplication

The same information may be discovered through multiple paths:

```text
formal dependency
+ wikilink
+ backlink
+ FTS result
```

The Context Engine MUST deduplicate candidates before rendering.

If the same chunk is found through multiple signals, preserve those signals as ranking/explanation metadata rather than repeating the content.

Example:

```json
{
  "artifact": "KNOW-003",
  "section": "Constraints",
  "reasons": [
    "wikilink",
    "bm25"
  ]
}
```

---

## 19. Context Budgeting

The Context Engine exists primarily to avoid loading unnecessary context. Budgeting is therefore a core behavior, not an optional optimization.

Use an internal default initially, for example:

```go
const DefaultContextBudget = 6000
```

The exact default should be easy to change after measurement.

### 19.1 Budget strategy

Budget by priority tiers.

Conceptually:

```text
mandatory Constitution/context
        ↓
target artifact/context
        ↓
required structural relationships
        ↓
explicit semantic relationships
        ↓
FTS candidates
        ↓
optional graph expansion
```

Do not simply sort every chunk by one score and truncate blindly.

### 19.2 Oversized mandatory context

If mandatory context alone exceeds the budget, the engine should preserve correctness and explain the condition rather than silently omitting essential material.

Possible behavior:

- include mandatory content;
- mark `budget_exceeded: true`;
- report the estimated overage;
- omit lower-priority optional content.

Do not silently truncate requirements in a way that changes meaning.

---

## 20. Context Result Model

The internal result should be structured and explainable.

Suggested conceptual shape:

```go
type ContextItem struct {
    ArtifactID string
    Path       string
    Section    string
    Tier       int
    Score      float64
    Reasons    []string
    Content    string
    Tokens     int
}

type Result struct {
    Target          string
    Intent          Intent
    Budget          int
    EstimatedTokens int
    BudgetExceeded  bool
    Items           []ContextItem
    Stats           Stats
}
```

The engine should make it possible to answer:

- Why was this chunk included?
- Which artifact did it come from?
- Which section did it come from?
- How many tokens did it consume?
- Was it mandatory, structural, linked, or text-retrieved?

Explainability is important for dogfooding retrieval quality.

---

## 21. Internal CLI

Add an internal operation conceptually equivalent to:

```text
misterspec internal context SPEC-014 \
    --intent implementation \
    --task TASK-023 \
    --query "implement refresh token rotation" \
    --budget 6000
```

The command should initially return stable machine-readable JSON following existing MisterSpec internal-command conventions.

Example conceptual output:

```json
{
  "ok": true,
  "context": {
    "target": "SPEC-014",
    "intent": "implementation",
    "budget": 6000,
    "estimated_tokens": 4382,
    "budget_exceeded": false,
    "items": [
      {
        "artifact_id": null,
        "path": "ai/memory/constitution.md",
        "section": "Relevant Principles",
        "tier": "constitution",
        "reasons": ["mandatory"],
        "tokens": 421
      },
      {
        "artifact_id": "SPEC-014",
        "path": ".../SPEC-014/spec.md",
        "section": "Requirements / R3",
        "tier": "target",
        "reasons": ["target", "task_requirement"],
        "tokens": 640
      },
      {
        "artifact_id": "KNOW-003",
        "path": "ai/knowledge/KNOW-003-authentication.md",
        "section": "Constraints",
        "tier": "retrieval",
        "reasons": ["wikilink", "bm25"],
        "score": 0.91,
        "tokens": 380
      }
    ]
  }
}
```

Exact JSON should use the repository's established envelope/error patterns.

### 21.1 Rendered mode

After the JSON API is stable, optionally add:

```text
--render
```

which emits a compact agent-consumable Markdown context pack.

Example:

```markdown
<context>

## Constitution
...

## Target — SPEC-014
...

## Current Task — TASK-023
...

## Relevant Knowledge — KNOW-003 / Constraints
...

## Relevant Learnings — LRN-008
...

</context>
```

JSON should remain the canonical machine interface.

---

## 22. Retrieval Metrics

Token reduction must be measured rather than assumed.

For each context request, collect diagnostics such as:

```json
{
  "available_estimated_tokens": 48230,
  "selected_estimated_tokens": 4382,
  "excluded_estimated_tokens": 43848,
  "reduction_percent": 90.9,
  "selected_chunks": 14,
  "candidate_chunks": 126
}
```

These metrics are estimates and must be labeled accordingly.

Useful dogfooding metrics include:

- total indexed documents;
- total indexed chunks;
- candidate chunks per request;
- selected chunks per request;
- estimated available tokens;
- estimated selected tokens;
- estimated reduction percentage;
- number of structural hits;
- number of wikilink hits;
- number of backlink hits;
- number of FTS-only hits;
- number of additional files the agent had to inspect after receiving the pack.

The last metric is especially important: extreme token reduction is not useful if retrieval quality is poor and the agent must immediately rediscover the repository manually.

---

## 23. Skill Integration Strategy

Do NOT modify all Skills to depend on the Context Engine immediately.

First implement and dogfood the Context Engine independently.

The evaluation flow should be:

```text
existing Skill behavior
        │
        ├──────────────┐
        │              │
        ▼              ▼
manual/current     Context Engine
exploration        context pack
        │              │
        └──────┬───────┘
               ▼
       compare quality,
       token estimates,
       missing context,
       extra exploration
```

Only after retrieval quality is acceptable should Skills be updated.

### 23.1 Future `/implement` integration

The existing implementation flow conceptually becomes:

```text
resolve target
inspect target
select executable Task
        ↓
request implementation Context Pack
        ↓
read source files explicitly named by Task/Plan/context
        ↓
expand repository exploration only when needed
        ↓
implement
verify
validate
```

A Skill should be instructed to begin from the returned Context Pack rather than recursively reading unrelated Knowledge or project artifacts.

The agent MUST remain allowed to expand context when necessary.

The Context Engine is a bootstrap/retrieval optimization, not a sandbox preventing repository access.

---

## 24. Error Handling

Errors should follow existing MisterSpec conventions and remain machine-readable through internal commands.

Expected classes include:

- target does not exist;
- target ID is invalid;
- target ID is ambiguous;
- index cannot be opened;
- index schema incompatible;
- index synchronization failed;
- malformed Markdown structure;
- malformed wikilink;
- broken wikilink;
- ambiguous wikilink;
- unsupported intent;
- invalid budget.

Where possible, a corrupt/disposable index should be recoverable by rebuilding it rather than requiring user repair.

Authoritative artifact errors must not be hidden by the Context Engine.

---

## 25. Index Versioning

The SQLite database should have an explicit schema version.

Because the database is disposable, migration complexity should be minimized.

For early versions it is acceptable to detect an incompatible schema and rebuild the index automatically.

Conceptually:

```text
expected schema version != actual schema version
        ↓
discard/recreate derived database
        ↓
full synchronization
```

Never mutate authoritative Markdown as part of an index migration.

---

## 26. Concurrency and Transactions

SQLite writes should use transactions during synchronization so a failed update does not leave partially updated derived state.

Prefer simple synchronization semantics over premature concurrent indexing.

The expected MisterSpec artifact corpus is small enough that correctness and determinism matter more than maximizing write throughput.

Do not introduce worker pools or complicated concurrency until profiling demonstrates a need.

---

## 27. Performance Expectations

The Context Engine should optimize for interactive local use, not analytics throughput.

Expected workload characteristics:

```text
hundreds to thousands of artifacts
thousands to tens of thousands of chunks
small incremental updates
small top-N retrievals
frequent ID/relationship lookups
frequent FTS queries
```

SQLite should be more than sufficient for this scale.

Benchmark before introducing a different database.

Performance work should prioritize:

1. avoiding unnecessary reparsing;
2. incremental fingerprint-based synchronization;
3. efficient FTS queries;
4. bounded graph traversal;
5. bounded result counts;
6. avoiding unnecessary context rendering.

---

## 28. Security and Privacy

The MVP should keep the index entirely local.

No artifact content should be sent to an external embedding/search provider as part of Context Engine indexing.

The derived database may contain copies of artifact text and should therefore be treated as project-local data.

The cache path should inherit normal project filesystem permissions.

---

## 29. Testing Strategy

Each layer requires focused tests.

### 29.1 Wikilink parser tests

Cover:

- simple link;
- aliased link;
- multiple links on a line;
- multiple lines;
- malformed links;
- line-number preservation;
- link ordering;
- non-link bracket syntax;
- IDs of each supported entity type.

### 29.2 Wikilink validation tests

Cover:

- valid target;
- missing target;
- malformed ID;
- wrong configured width;
- ambiguous target;
- alias does not affect resolution.

### 29.3 Reference/backlink tests

Cover:

- formal references;
- semantic references;
- mixed references;
- backlinks;
- duplicate references;
- self-reference behavior;
- deterministic ordering.

### 29.4 Chunker tests

Cover:

- headings;
- nested headings;
- large sections;
- empty sections;
- frontmatter exclusion;
- provenance lines;
- wikilinks inside chunks;
- stable chunk output for unchanged content.

### 29.5 Index tests

Cover:

- initial build;
- unchanged synchronization;
- changed file;
- new file;
- deleted file;
- transaction rollback;
- database rebuild;
- FTS synchronization;
- incoming/outgoing links.

### 29.6 Retrieval tests

Construct small deterministic fixture repositories proving that:

- target context always wins over unrelated FTS hits;
- formal dependency outranks textual similarity;
- direct wikilink outranks weak FTS similarity;
- backlinks can contribute candidates;
- second-hop traversal is bounded;
- duplicate discovery does not duplicate content;
- intent changes ordering appropriately.

### 29.7 Budget tests

Cover:

- result below budget;
- optional content removed at budget boundary;
- mandatory context retained;
- mandatory context exceeds budget;
- deterministic output for the same request;
- token estimates included in result.

### 29.8 CLI tests

Follow existing CLI test conventions for:

- successful context request;
- invalid target;
- invalid intent;
- invalid budget;
- index rebuild behavior;
- stable JSON envelope.

---

## 30. Implementation Phases

Implement in the following order.

### Phase 1 — Artifact Graph Foundation

Implement:

- `WikiLink` representation;
- wikilink extraction;
- tests;
- integration with artifact body parsing;
- validation findings for invalid/broken/ambiguous wikilinks.

Do not add SQLite yet.

**Exit criteria:** MisterSpec can deterministically parse and validate semantic links between artifacts.

### Phase 2 — References and Backlinks

Implement:

- formal outgoing-reference collection;
- semantic outgoing-reference collection;
- backlink discovery;
- internal `references` operation;
- internal `backlinks` operation;
- tests.

**Exit criteria:** MisterSpec exposes a deterministic artifact graph without relying on a database.

### Phase 3 — Document Model and Chunking

Implement:

- Markdown body document representation;
- section parsing;
- semantic chunking;
- provenance;
- token estimator abstraction;
- tests.

**Exit criteria:** any eligible artifact can be converted into stable searchable chunks.

### Phase 4 — Disposable SQLite Index

Implement:

- context index package;
- schema/version handling;
- documents;
- chunks;
- links;
- FTS5;
- safe database creation/rebuild;
- tests.

**Exit criteria:** the full artifact corpus can be indexed and searched locally.

### Phase 5 — Incremental Synchronization

Implement:

- fingerprint comparison;
- new/change/delete detection;
- transactional updates;
- unchanged-file skipping;
- tests.

**Exit criteria:** repeated synchronization is cheap and deterministic.

### Phase 6 — Context Collector and Retrieval

Implement:

- request model;
- intent model;
- mandatory target collection;
- formal graph traversal;
- wikilink traversal;
- backlinks;
- BM25 retrieval;
- bounded second-hop traversal;
- candidate deduplication;
- tests.

**Exit criteria:** the engine can produce a ranked candidate set with reasons.

### Phase 7 — Ranking and Budgeting

Implement:

- tier model;
- intent weights;
- relationship weights;
- text relevance integration;
- token budgeting;
- mandatory-context handling;
- diagnostics;
- tests.

**Exit criteria:** the engine returns the smallest ranked context set that fits the requested budget while preserving mandatory information.

### Phase 8 — Internal Context Command

Implement:

```text
misterspec internal context
```

with target, intent, optional Task/query, and budget.

Return stable JSON.

Add `--render` only after the structured interface is stable.

**Exit criteria:** coding agents and tests can consume Context Engine results through the MisterSpec binary.

### Phase 9 — Dogfooding and Evaluation

Do not change core Skills yet.

Run Context Engine requests against real MisterSpec development tasks and record:

- selected context;
- omitted context;
- estimated token reduction;
- retrieval mistakes;
- additional exploration required by the agent;
- latency;
- index size.

Tune ranking only from observed failures, not intuition alone.

**Exit criteria:** retrieval is consistently useful enough to bootstrap actual coding work.

**Note (037-eval-quality-efficiency):** the one-off dogfooding exercise
described above (019-dogfooding-evaluation) has since been superseded
by a reusable, repeatable evaluation harness — `misterspec internal
eval-retrieval` (deterministic, CI-runnable, no LLM session required)
and `misterspec internal eval-compare` (baseline vs. candidate
comparison, one-dimension-at-a-time). Any future retrieval/ranking
proposal (e.g. a PROP-06/08/13-style change) should record a baseline
and compare against it through this harness — see
`specs/037-eval-quality-efficiency/quickstart.md`,
`contracts/eval-commands-contract.md`, and
`docs/eval-task-execution-protocol.md` for the live-agent-session half
— rather than re-inventing an ad hoc measurement.

### Phase 10 — Skill Integration

Update Skills incrementally, beginning with the highest-value workflows such as implementation, planning, task creation, and analysis.

Each Skill should:

1. resolve/inspect its target using existing operations;
2. request an intent-specific Context Pack;
3. consume that pack first;
4. expand context only when needed;
5. continue using normal MisterSpec validation and completion rules.

**Exit criteria:** Context Engine usage reduces repeated context loading without changing Skill correctness or authority boundaries.

---

## 31. Suggested Package Layout

This is a target shape, not a requirement to reorganize unrelated existing files.

```text
internal/
├── artifacts/
│   ├── ... existing artifact model
│   ├── document.go
│   ├── markdown.go
│   └── wikilink.go
│
├── operations/
│   ├── ... existing operations
│   ├── references.go
│   └── backlinks.go
│
├── context/
│   ├── engine.go
│   ├── request.go
│   ├── result.go
│   ├── collector.go
│   ├── ranker.go
│   ├── budget.go
│   ├── render.go
│   ├── tokens.go
│   └── index/
│       ├── store.go
│       ├── sqlite.go
│       ├── schema.go
│       ├── sync.go
│       ├── search.go
│       └── chunk.go
│
└── ... existing packages
```

Follow existing repository naming and error conventions when implementation details conflict with this illustrative layout.

---

## 32. Implementation Constraints for Claude Code

When implementing this specification:

1. Inspect the current `dev` branch before modifying code. Treat current code as authoritative over illustrative snippets in this document.
2. Reuse existing ID parsing, project configuration, artifact scanning, resolution, inspection, validation, fingerprinting, JSON envelopes, and CLI patterns.
3. Do not duplicate structural logic merely to make the Context Engine self-contained.
4. Keep new APIs small and package responsibilities narrow.
5. Prefer deterministic behavior over agent inference inside Go code.
6. Do not add embeddings or remote services.
7. Do not index arbitrary source code in the MVP.
8. Do not add user configuration unless required by an existing convention or proven implementation need.
9. Keep SQLite derived and disposable.
10. Add tests alongside each phase rather than implementing all phases before testing.
11. Preserve backward compatibility of existing Skills and internal operations until the explicit Skill Integration phase.
12. Do not silently modify existing artifact semantics to accommodate retrieval.
13. If an implementation choice conflicts with `docs/architecture-specification.md`, stop and resolve the architectural conflict rather than silently overriding it.
14. If the repository has evolved beyond assumptions in this document, adapt the implementation to the current architecture while preserving the goals and invariants specified here.

---

## 33. Definition of Done

The Context Engine MVP is complete when all of the following are true:

- MisterSpec artifacts can contain ID-based wikilinks.
- Wikilinks are parsed and structurally validated.
- Formal references and semantic links can be queried.
- Backlinks can be queried.
- Artifact Markdown can be chunked by semantic section with provenance.
- A disposable SQLite/FTS5 index can be fully rebuilt from filesystem state.
- Incremental synchronization skips unchanged artifacts.
- Context retrieval prioritizes target and explicit structure before text similarity.
- BM25 retrieval supplements graph retrieval.
- Retrieval is intent-aware.
- Results are deduplicated.
- A token budget limits optional context.
- Mandatory context is never silently discarded merely to satisfy a budget.
- `misterspec internal context` exposes a stable structured result.
- Every selected context item explains its provenance and inclusion reason.
- Token-reduction diagnostics are available.
- Existing MisterSpec workflows continue to work without the Context Engine.
- The index can be deleted and reconstructed without data loss.
- Automated tests cover parsing, graph operations, indexing, synchronization, retrieval, ranking, budgeting, and CLI behavior.
- Dogfooding demonstrates useful context reduction without unacceptable retrieval failures.

---

## 34. Future Work — Explicitly Deferred

The following ideas are intentionally deferred until the MVP has measurements.

### Hybrid semantic retrieval

If FTS5 + graph retrieval demonstrably misses semantically relevant Knowledge or Learnings, evaluate local or provider-based embeddings as an optional second-stage retrieval mechanism.

Do not introduce embeddings simply because they are common in RAG systems.

### Source-code indexing

If agents repeatedly need broad source exploration even after receiving artifact context and explicit source paths, evaluate syntax-aware source indexing separately.

Tree-sitter or language-specific symbol extraction would be preferable to naive fixed-size source chunks.

### Memory compaction

If Learnings grow large enough to degrade retrieval, consider derived summaries/compaction while preserving original artifacts and provenance.

Derived summaries must never silently become authoritative truth.

### Advanced graph traversal

Only consider more sophisticated graph algorithms if direct and second-hop traversal prove insufficient.

### Provider-aware tokenization

Exact Claude/OpenAI/etc. tokenizers may replace the approximate estimator if budgeting accuracy becomes important enough to justify provider-specific code.

### Visual knowledge graph

A human-facing graph visualization may eventually be useful, but it is independent from the Context Engine's core purpose and should not block retrieval work.

---

## 35. Final Architectural Model

The resulting MisterSpec model should be understood as:

```text
                         Filesystem
                      authoritative state
                             │
                             ▼
                    MisterSpec Artifacts
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
       Formal Structure   Wikilinks      Markdown Content
              │              │              │
              └──────────────┼──────────────┘
                             ▼
                       Artifact Graph
                             │
                  ┌──────────┴──────────┐
                  │                     │
                  ▼                     ▼
          Deterministic Graph      SQLite FTS5
             Traversal                Search
                  │                     │
                  └──────────┬──────────┘
                             ▼
                       Context Ranker
                             │
                             ▼
                       Token Budgeter
                             │
                             ▼
                        Context Pack
                             │
                             ▼
                       Coding Agent
                             │
                             ▼
                     Semantic Judgment
```

The key design principle is:

> MisterSpec should not try to replace the coding agent's reasoning. It should make the right project knowledge cheap to find, cheap to load, and difficult to accidentally ignore.

The Context Engine therefore turns MisterSpec's existing artifact system into a lightweight project knowledge graph and retrieval layer while preserving the filesystem-first, deterministic-core, agent-independent architecture.