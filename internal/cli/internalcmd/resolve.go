package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewResolveCmd builds "misterspec internal resolve <id>"
// (docs/architecture-specification.md §9.2). It calls operations.Resolve
// directly — no logic of its own beyond argument parsing and JSON
// shaping (spec.md FR-009).
func NewResolveCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "resolve <id>",
		Short:         "Resolve an entity ID to its canonical artifact",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			loc, err := operations.Resolve(proj.Root, proj.Config, args[0])
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"entity": map[string]any{
					"id":   loc.ID.String(),
					"type": loc.Type.String(),
					"path": loc.Path,
				},
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}
