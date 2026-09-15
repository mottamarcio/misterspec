package internalcmd

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	contextengine "github.com/mottamarcio/misterspec/internal/context"
	"github.com/mottamarcio/misterspec/internal/context/index"
	"github.com/mottamarcio/misterspec/internal/project"
)

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
func NewContextCmd() *cobra.Command {
	var dir, intent, task, query, budget string
	var render bool

	cmd := &cobra.Command{
		Use:           "context <id>",
		Short:         "Return a budgeted, ranked context pack for an artifact",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			if cmd.Flags().Changed("budget") {
				n, err := strconv.Atoi(budget)
				if err != nil {
					return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: --budget %q is not a valid integer", ErrInvalidArgument, budget))
				}
				req.Budget = &n
			}

			candidates, err := contextengine.Collect(proj.Root, proj.Config, store, req)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}
			ranked := contextengine.Rank(candidates, req)
			result := contextengine.ApplyBudget(ranked, req)

			var rendered any
			if render {
				rendered = contextengine.Render(req, result)
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"context": map[string]any{
					"target":           req.Target,
					"intent":           string(req.Intent),
					"budget":           resolvedBudget(req),
					"estimated_tokens": result.Diagnostics.TokensSelected,
					"budget_exceeded":  result.BudgetExceeded,
					"overage":          result.Overage,
					"items":            renderContextItems(result.Items),
					"diagnostics": map[string]any{
						"candidates_considered": result.Diagnostics.CandidatesConsidered,
						"items_selected":        result.Diagnostics.ItemsSelected,
						"tokens_available":      result.Diagnostics.TokensAvailable,
						"tokens_selected":       result.Diagnostics.TokensSelected,
						"tokens_excluded":       result.Diagnostics.TokensExcluded,
						"reduction_percent":     result.Diagnostics.ReductionPercent,
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
	cmd.Flags().BoolVar(&render, "render", false, "additionally include a rendered Markdown context pack")
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
// non-nil slice (empty, never null, when items is empty).
func renderContextItems(items []contextengine.ResultItem) []map[string]any {
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
		out = append(out, map[string]any{
			"path":    item.Path,
			"heading": item.Heading,
			"tier":    tier.String(),
			"reasons": reasons,
			"score":   item.Score,
			"tokens":  item.Tokens,
		})
	}
	return out
}
