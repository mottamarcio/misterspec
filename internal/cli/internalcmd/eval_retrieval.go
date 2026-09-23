package internalcmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/context/index"
	"github.com/mottamarcio/misterspec/internal/eval"
	"github.com/mottamarcio/misterspec/internal/project"
)

// evalRetrievalSchemaVersion is "eval-retrieval"'s own response
// contract version (037-eval-quality-efficiency contracts/
// eval-commands-contract.md §1).
const evalRetrievalSchemaVersion = 1

// NewEvalRetrievalCmd builds "misterspec internal eval-retrieval"
// (037-eval-quality-efficiency contracts/eval-commands-contract.md
// §1). Runs every EvaluationCase under --cases against the existing
// contextengine collector/ranker in-process (plan.md Summary) and
// reports a per-case pass/fail plus an aggregate summary — entirely
// offline and deterministic (spec User Story 1, FR-001).
func NewEvalRetrievalCmd() *cobra.Command {
	var casesDir, out, variant, repoRevision string

	cmd := &cobra.Command{
		Use:           "eval-retrieval",
		Short:         "Run a deterministic set of retrieval evaluation cases",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if casesDir == "" {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: --cases is required", ErrInvalidArgument))
			}

			cases, err := eval.LoadCases(casesDir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			results := make([]eval.CaseResult, 0, len(cases))
			for _, c := range cases {
				proj, err := project.Detect(c.ResolvedDir)
				if err != nil {
					return WriteError(cmd.OutOrStdout(), err)
				}

				cachePath := filepath.Join(proj.Root, ".misterspec", "cache", "context.db")
				store, err := index.Open(cachePath)
				if err != nil {
					return WriteError(cmd.OutOrStdout(), err)
				}
				if _, err := store.Sync(proj.Root, proj.Config); err != nil {
					store.Close()
					return WriteError(cmd.OutOrStdout(), err)
				}

				result, err := eval.RunRetrievalCase(proj.Root, proj.Config, store, c)
				store.Close()
				if err != nil {
					return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %v", eval.ErrInvalidCase, err))
				}
				results = append(results, result)
			}

			if repoRevision == "" {
				repoRevision = detectRepoRevision(casesDir)
			}
			if variant == "" {
				variant = "default"
			}

			resultsJSON, err := json.Marshal(results)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			run := eval.RunRecord{
				RunID:     fmt.Sprintf("retrieval-%s", time.Now().UTC().Format(time.RFC3339)),
				Kind:      eval.KindRetrieval,
				CreatedAt: time.Now().UTC(),
				Config: map[string]string{
					"variant":                variant,
					"repo_revision":          repoRevision,
					"ranking_version":        strconv.Itoa(rankingVersion),
					"context_schema_version": strconv.Itoa(contextSchemaVersion),
				},
				Results: resultsJSON,
			}

			if out != "" {
				if err := eval.SaveRunRecord(out, run); err != nil {
					return WriteError(cmd.OutOrStdout(), err)
				}
			}

			summary := eval.SummarizeCaseResults(results)

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"schema_version": evalRetrievalSchemaVersion,
				"result": map[string]any{
					"run":     run,
					"summary": summary,
				},
			})
		},
	}

	cmd.Flags().StringVar(&casesDir, "cases", "", "directory of retrieval evaluation case files (required)")
	cmd.Flags().StringVar(&out, "out", "", "path to also write the full RunRecord JSON")
	cmd.Flags().StringVar(&variant, "variant", "", "config label for this run's variant (default: \"default\")")
	cmd.Flags().StringVar(&repoRevision, "repo-revision", "", "config label for this run's repo revision (default: best-effort git rev-parse --short HEAD)")
	return cmd
}

// detectRepoRevision best-effort shells out to `git -C dir rev-parse
// --short HEAD`, returning "" (never an error) when dir is not a Git
// repository or git is unavailable — repo_revision is a config label
// for comparisons, not a requirement (research.md #3).
func detectRepoRevision(dir string) string {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
