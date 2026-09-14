package internalcmd

import (
	"fmt"

	"github.com/spf13/cobra"

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

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"created": map[string]any{
					"id":   result.ID.String(),
					"type": result.ID.Type.String(),
					"path": result.Path,
				},
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	cmd.Flags().StringVar(&parent, "parent", "", "parent entity ID (required for feature, spec)")
	cmd.Flags().StringVar(&slug, "slug", "", "filename slug (required for knowledge, learning)")
	return cmd
}
