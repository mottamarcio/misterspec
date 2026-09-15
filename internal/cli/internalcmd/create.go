package internalcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewCreateCmd builds "misterspec internal create <type> [--parent P]
// [--slug S]" (docs/architecture-specification.md §10-11). It calls
// operations.Create directly — no logic of its own beyond argument/flag
// parsing and JSON shaping (spec.md FR-009).
func NewCreateCmd() *cobra.Command {
	var dir, parent, slug string

	cmd := &cobra.Command{
		Use:           "create <type>",
		Short:         "Atomically allocate the next ID and scaffold its initial artifact",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, ok := entityTypeNames[args[0]]
			if !ok {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: unrecognized type %q", ErrInvalidArgument, args[0]))
			}

			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			result, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
				Type:   t,
				Parent: parent,
				Slug:   slug,
			})
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			created := map[string]any{
				"id":   result.ID.String(),
				"type": result.ID.Type.String(),
				"path": result.Path,
			}
			if git := gitOutcomeJSON(result); git != nil {
				created["git"] = git
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"created": created,
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	cmd.Flags().StringVar(&parent, "parent", "", "parent entity ID (required for feature, spec)")
	cmd.Flags().StringVar(&slug, "slug", "", "filename slug (required for knowledge, learning); optional readable branch-name suffix for feature")
	return cmd
}

// gitOutcomeJSON shapes result's Git-related fields per
// contracts/create-git.md (022-feature-branch-automation), or returns
// nil for an entity type this feature never touches (Program, Task,
// Knowledge, Learning — none of which ever populate GitBranch/
// GitSkippedReason).
func gitOutcomeJSON(result operations.CreateResult) map[string]any {
	switch result.ID.Type {
	case ids.Feature:
		if result.GitSkippedReason != "" {
			return map[string]any{"skipped_reason": result.GitSkippedReason}
		}
		return map[string]any{
			"branch":  result.GitBranch,
			"created": result.GitBranchCreated,
		}

	case ids.Spec:
		if result.GitSkippedReason != "" {
			return map[string]any{"skipped_reason": result.GitSkippedReason}
		}
		git := map[string]any{"branch": result.GitBranch}
		if result.GitWarning != "" {
			git["warning"] = result.GitWarning
		}
		return git

	default:
		return nil
	}
}
