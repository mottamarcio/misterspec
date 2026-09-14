package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewStatusCmd builds "misterspec internal status"
// (docs/architecture-specification.md §18). It calls operations.Status
// directly — no logic of its own beyond JSON shaping (spec.md FR-009).
// counts is keyed by EntityType.String() — singular, lowercase,
// consistent with every other command's "type" field (research.md).
func NewStatusCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Return a machine-facing structural status summary",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			summary, err := operations.Status(proj.Root, proj.Config)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			counts := make(map[string]int, len(summary.Counts))
			for t, n := range summary.Counts {
				counts[t.String()] = n
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"counts":            counts,
				"specs":             summary.SpecsByState,
				"structural_errors": summary.StructuralErrors,
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}
