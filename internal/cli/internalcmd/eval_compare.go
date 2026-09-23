package internalcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/eval"
)

// evalCompareSchemaVersion is "eval-compare"'s own response contract
// version (037-eval-quality-efficiency contracts/
// eval-commands-contract.md §2).
const evalCompareSchemaVersion = 1

// NewEvalCompareCmd builds "misterspec internal eval-compare"
// (037-eval-quality-efficiency contracts/eval-commands-contract.md
// §2). Compares a candidate RunRecord against a named Baseline,
// flagging multi-dimension differences and surfacing per-case
// regressions separately from aggregates (spec User Story 3,
// FR-006/FR-012/FR-013).
func NewEvalCompareCmd() *cobra.Command {
	var dir, baselineName, candidatePath, repetitionsDir string

	cmd := &cobra.Command{
		Use:           "eval-compare",
		Short:         "Compare a candidate evaluation run against a named baseline",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if baselineName == "" {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: --baseline is required", ErrInvalidArgument))
			}
			if candidatePath == "" {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: --candidate is required", ErrInvalidArgument))
			}

			root := dir
			if root == "" {
				root = "."
			}

			baselinePath := filepath.Join(root, "eval", "baselines", baselineName+".json")
			baseline, err := eval.LoadBaseline(baselinePath)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %v", ErrInvalidArgument, err))
			}
			if len(baseline.RunIDs) == 0 {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: baseline %q has no recorded run_ids", ErrInvalidArgument, baselineName))
			}

			baselineRun, err := eval.LoadRunRecord(filepath.Join(root, "eval", "runs", baseline.RunIDs[0]+".json"))
			if err != nil {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %v", ErrInvalidArgument, err))
			}

			candidate, err := eval.LoadRunRecord(candidatePath)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %v", ErrInvalidArgument, err))
			}

			// Baseline-native repetitions (data-model.md "Baseline":
			// run_ids beyond the first are "repetitions of the same
			// config" as the baseline itself, not the candidate's) —
			// validated against baselineRun's own config, never the
			// candidate's, since the candidate's config is legitimately
			// expected to differ (that difference is what eval-compare
			// exists to isolate).
			var repetitions []eval.RunRecord
			for _, runID := range baseline.RunIDs[1:] {
				rep, err := eval.LoadRunRecord(filepath.Join(root, "eval", "runs", runID+".json"))
				if err != nil {
					return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %v", ErrInvalidArgument, err))
				}
				if diff := eval.DimensionDiff(baselineRun.Config, rep.Config); len(diff) != 0 {
					return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: baseline repetition %q's config differs from the baseline's own config in %v — a baseline's own repetitions must share the baseline's config (minus repetition_index)", ErrInvalidArgument, rep.RunID, diff))
				}
				repetitions = append(repetitions, rep)
			}
			// --repetitions files represent repeated candidate runs, so
			// they are validated against the candidate's own config.
			if repetitionsDir != "" {
				extra, err := loadRepetitions(repetitionsDir)
				if err != nil {
					return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %v", ErrInvalidArgument, err))
				}
				for _, rep := range extra {
					if diff := eval.DimensionDiff(candidate.Config, rep.Config); len(diff) != 0 {
						return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: repetition %q's config differs from the candidate's own config in %v — repetitions must share the candidate's config (minus repetition_index)", ErrInvalidArgument, rep.RunID, diff))
					}
				}
				repetitions = append(repetitions, extra...)
			}

			report, err := eval.Compare(baselineRun, candidate, repetitions)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}
			report.BaselineName = baselineName

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"schema_version": evalCompareSchemaVersion,
				"result": map[string]any{
					"comparison": report,
				},
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "project directory holding eval/baselines and eval/runs")
	cmd.Flags().StringVar(&baselineName, "baseline", "", "name of a recorded baseline under eval/baselines (required)")
	cmd.Flags().StringVar(&candidatePath, "candidate", "", "path to a candidate RunRecord JSON file (required)")
	cmd.Flags().StringVar(&repetitionsDir, "repetitions", "", "path to a directory of additional same-config RunRecord files, used to compute variance")
	return cmd
}

// loadRepetitions loads every *.json file directly under dir as a
// RunRecord, in filename-sorted order.
func loadRepetitions(dir string) ([]eval.RunRecord, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading repetitions directory %q: %w", dir, err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)

	out := make([]eval.RunRecord, 0, len(files))
	for _, f := range files {
		r, err := eval.LoadRunRecord(f)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}
