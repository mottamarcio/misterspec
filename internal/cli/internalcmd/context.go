package internalcmd

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	contextengine "github.com/mottamarcio/misterspec/internal/context"
	"github.com/mottamarcio/misterspec/internal/context/index"
	"github.com/mottamarcio/misterspec/internal/project"
)

// contextOutputMode is the caller's explicit choice of response shape
// (033-context-pack-output-contract data-model.md "OutputMode").
type contextOutputMode string

const (
	modeManifest contextOutputMode = "manifest"
	modePackage  contextOutputMode = "package"
	modeMarkdown contextOutputMode = "markdown"
)

// validContextModes is the fixed --mode vocabulary; anything else is a
// CLI-boundary invalid_argument (research.md Decision 2).
var validContextModes = map[contextOutputMode]bool{
	modeManifest: true,
	modePackage:  true,
	modeMarkdown: true,
}

// contextSchemaVersion is the "context" envelope's own contract version
// (033-context-pack-output-contract research.md Decision 5) — present
// in every mode, including the unchanged manifest default. Bumped to 2
// by 035-context-budget-accuracy: budget_exceeded/overage now compare
// mandatory content against the new, separate hard limit instead of
// the soft budget (contracts/budget-and-estimator-contract.md §2).
// Bumped to 3 by 036-text-search-ranking: same-tier free-text ordering
// is now BM25-derived, and diagnostics gains ranking_version
// (contracts/search-and-ranking-contract.md §3).
const contextSchemaVersion = 3

// rankingVersion identifies the scoring formula/weight set that
// produced a response's ordering (036-text-search-ranking spec FR-009).
// Bumped only after evaluation evidence supports the change
// (specs/019-dogfooding-evaluation/contracts/evaluation-protocol.md;
// spec FR-010, SC-005) — never silently under the same value.
const rankingVersion = 1

// NewContextCmd builds "misterspec internal context <id>"
// (docs/context-engine-implementation.md §21,
// specs/017-internal-context-command/contracts/context-command.md). It
// orchestrates capability 011-016 already built — opening/synchronizing
// the disposable search index (014, transparently establishing "Index
// Readiness" per FR-007 with no new repair logic of its own —
// research.md #2), then calling contextengine.Collect (015),
// contextengine.Rank and contextengine.ApplyBudget (016) — and renders
// the result as one stable JSON envelope, optionally including a
// Markdown context pack (contextengine.Render) when --render is passed
// (research.md #7). No new ranking or budgeting logic of its own.
// --mode (033-context-pack-output-contract) selects manifest (default,
// today's exact shape) / package (full content) / markdown (--render's
// own shape, requestable by name); --mode=package combined with
// --render is rejected (research.md Decision 3).
// validQueryModes is the fixed --query-mode vocabulary; anything else
// is a CLI-boundary invalid_argument (036-text-search-ranking
// contracts §1).
var validQueryModes = map[string]bool{
	"":         true,
	"free":     true,
	"advanced": true,
}

func NewContextCmd() *cobra.Command {
	var dir, intent, task, query, budget, hardLimit, mode, queryMode string
	var render, diagnosticScores bool

	cmd := &cobra.Command{
		Use:           "context <id>",
		Short:         "Return a budgeted, ranked context pack for an artifact",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Cheap CLI-boundary validation first, before any filesystem
			// or index I/O — an invalid --query-mode must not pay for a
			// full store.Sync only to be rejected afterward.
			if !validQueryModes[queryMode] {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: unrecognized --query-mode %q", ErrInvalidArgument, queryMode))
			}

			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			cachePath := filepath.Join(proj.Root, ".misterspec", "cache", "context.db")
			store, err := index.Open(cachePath)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}
			defer store.Close()

			if _, err := store.Sync(proj.Root, proj.Config); err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			req := contextengine.Request{
				Target: args[0],
				Task:   task,
				Intent: contextengine.Intent(intent),
				Query:  query,
			}
			if queryMode == "advanced" {
				req.QueryMode = contextengine.QueryModeAdvanced
			}
			if cmd.Flags().Changed("budget") {
				n, err := strconv.Atoi(budget)
				if err != nil {
					return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: --budget %q is not a valid integer", ErrInvalidArgument, budget))
				}
				req.Budget = &n
			}
			if cmd.Flags().Changed("hard-limit") {
				n, err := strconv.Atoi(hardLimit)
				if err != nil {
					return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: --hard-limit %q is not a valid integer", ErrInvalidArgument, hardLimit))
				}
				req.HardLimit = &n
			}

			outputMode := contextOutputMode(mode)
			if outputMode == "" {
				outputMode = modeManifest
			}
			if !validContextModes[outputMode] {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: unrecognized --mode %q", ErrInvalidArgument, mode))
			}
			if outputMode == modePackage && render {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: --render cannot be combined with --mode=package", ErrInvalidArgument))
			}

			candidates, err := contextengine.Collect(proj.Root, proj.Config, store, req)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}
			ranked := contextengine.Rank(candidates, req)
			estimator := artifacts.DefaultEstimator{}
			result := contextengine.ApplyBudget(ranked, req, estimator)

			var rendered any
			if render || outputMode == modeMarkdown {
				rendered = contextengine.Render(req, result)
			}

			var items any
			if outputMode == modePackage {
				items = renderPackageItems(contextengine.BuildPackageItems(result.Items), result.Items, diagnosticScores)
			} else {
				items = renderContextItems(result.Items, diagnosticScores)
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"context": map[string]any{
					"schema_version":   contextSchemaVersion,
					"target":           req.Target,
					"intent":           string(req.Intent),
					"budget":           resolvedBudget(req),
					"estimated_tokens": result.Diagnostics.TokensSelected,
					"budget_exceeded":  result.BudgetExceeded,
					"overage":          result.Overage,
					"items":            items,
					"diagnostics": map[string]any{
						"candidates_considered": result.Diagnostics.CandidatesConsidered,
						"items_selected":        result.Diagnostics.ItemsSelected,
						"tokens_available":      result.Diagnostics.TokensAvailable,
						"tokens_selected":       result.Diagnostics.TokensSelected,
						"tokens_excluded":       result.Diagnostics.TokensExcluded,
						"reduction_percent":     result.Diagnostics.ReductionPercent,
						"payload_tokens":        payloadTokens(outputMode, result, rendered, estimator),
						"estimator":             result.Diagnostics.Estimator,
						"hard_limit":            result.Diagnostics.HardLimit,
						"exclusions":            renderExclusions(result.Diagnostics.Exclusions),
						"ranking_version":       rankingVersion,
					},
					"rendered": rendered,
				},
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	cmd.Flags().StringVar(&intent, "intent", "", "intent: planning, tasks, implementation, validation, or analysis")
	cmd.Flags().StringVar(&task, "task", "", "current task's own text")
	cmd.Flags().StringVar(&query, "query", "", "free-text query")
	cmd.Flags().StringVar(&budget, "budget", "", "token budget as an integer, including negative (default: contextengine.DefaultBudget when omitted)")
	cmd.Flags().StringVar(&hardLimit, "hard-limit", "", "hard token ceiling above which mandatory content is flagged as exceeded (default: contextengine.DefaultHardLimit when omitted)")
	cmd.Flags().BoolVar(&render, "render", false, "additionally include a rendered Markdown context pack")
	cmd.Flags().StringVar(&mode, "mode", "", "output mode: manifest (default), package, or markdown")
	cmd.Flags().BoolVar(&diagnosticScores, "diagnostic-scores", false, "additionally include each item's score_components breakdown")
	cmd.Flags().StringVar(&queryMode, "query-mode", "", "query interpretation: free (default, literal text) or advanced (native FTS5 syntax)")
	return cmd
}

// resolvedBudget reports the budget req actually resolves to for
// echoing in the response — req.Budget's own pointed-to value when set,
// contextengine.DefaultBudget otherwise.
func resolvedBudget(req contextengine.Request) int {
	if req.Budget != nil {
		return *req.Budget
	}
	return contextengine.DefaultBudget
}

// renderContextItems renders result.Items as JSON-ready maps, always a
// non-nil slice (empty, never null, when items is empty). This is
// "manifest" mode's own item shape — byte-for-byte unchanged from
// before 033-context-pack-output-contract (research.md Decision 2).
// When diagnosticScores is true, each map additionally gains
// "score_components" (036-text-search-ranking spec FR-008, contracts
// §3.2); omitted entirely otherwise — never present as a null field.
func renderContextItems(items []contextengine.ResultItem, diagnosticScores bool) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		reasons := make([]string, 0, len(item.Reasons))
		tier := item.Reasons[0].Tier
		for _, r := range item.Reasons {
			reasons = append(reasons, r.Relation)
			if r.Tier < tier {
				tier = r.Tier
			}
		}
		m := map[string]any{
			"path":    item.Path,
			"heading": item.Heading,
			"tier":    tier.String(),
			"reasons": reasons,
			"score":   item.Score,
			"tokens":  item.Tokens,
		}
		if diagnosticScores {
			m["score_components"] = renderScoreComponents(item.Components)
		}
		out = append(out, m)
	}
	return out
}

// renderPackageItems renders "package"-mode items — every manifest
// field plus content/location/fingerprint (data-model.md
// "PackageItem", contracts §4). source is the same-order ResultItem
// slice items was built from (BuildPackageItems preserves order 1:1),
// used only to look up each item's own score_components when
// diagnosticScores is true.
func renderPackageItems(items []contextengine.PackageItem, source []contextengine.ResultItem, diagnosticScores bool) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for i, item := range items {
		m := map[string]any{
			"path":    item.Path,
			"heading": item.Heading,
			"tier":    item.Tier.String(),
			"reasons": item.Reasons,
			"score":   item.Score,
			"tokens":  item.Tokens,
			"content": item.Content,
			"location": map[string]any{
				"path":       item.Location.Path,
				"start_line": item.Location.StartLine,
				"end_line":   item.Location.EndLine,
			},
			"fingerprint": item.Fingerprint,
		}
		if diagnosticScores && i < len(source) {
			m["score_components"] = renderScoreComponents(source[i].Components)
		}
		out = append(out, m)
	}
	return out
}

// renderScoreComponents renders one ScoreComponents as a JSON-ready map
// (contracts §3.2).
func renderScoreComponents(c contextengine.ScoreComponents) map[string]any {
	return map[string]any{
		"tier":            c.Tier,
		"relation_weight": c.RelationWeight,
		"intent_bonus":    c.IntentBonus,
		"text_relevance":  c.TextRelevance,
		"total":           c.Total,
	}
}

// renderExclusions renders result.Diagnostics.Exclusions as JSON-ready
// maps, always a non-nil slice (035-context-budget-accuracy contracts
// §2.2).
func renderExclusions(exclusions []contextengine.ExclusionRecord) []map[string]any {
	out := make([]map[string]any, 0, len(exclusions))
	for _, e := range exclusions {
		out = append(out, map[string]any{
			"path":    e.Path,
			"heading": e.Heading,
			"reason":  e.Reason,
		})
	}
	return out
}

// payloadTokens estimates the serialized size of the response actually
// being returned (research.md Decision 6): package mode measures its
// own items' content plus metadata overhead; a non-nil rendered string
// (markdown mode, or the classic --render flag on its own) measures
// that string; otherwise (manifest, nothing rendered) there is no
// content beyond what tokens_selected already counts.
func payloadTokens(mode contextOutputMode, result contextengine.Result, rendered any, estimator artifacts.Estimator) int {
	if mode == modePackage {
		return contextengine.PayloadTokens(result.Diagnostics.TokensSelected, contextengine.BuildPackageItems(result.Items), "", estimator)
	}
	if s, ok := rendered.(string); ok {
		return contextengine.PayloadTokens(result.Diagnostics.TokensSelected, nil, s, estimator)
	}
	return result.Diagnostics.TokensSelected
}
