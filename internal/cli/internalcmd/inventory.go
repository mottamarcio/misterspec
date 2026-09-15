package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewInventoryCmd builds "misterspec internal inventory <dir>"
// (docs/architecture-specification.md §14). It calls
// operations.Inventory directly with <dir> taken as a literal
// project-relative path — no scope-keyword translation, no logic of its
// own beyond argument parsing and JSON shaping (spec.md FR-009,
// research.md).
func NewInventoryCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "inventory <dir>",
		Short:         "Return a deterministic file inventory for a directory",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			entries, err := operations.Inventory(proj.Root, args[0])
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			files := make([]map[string]any, 0, len(entries))
			for _, e := range entries {
				files = append(files, map[string]any{
					"path":      e.Path,
					"extension": e.Extension,
					"size":      e.Size,
				})
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{"files": files})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}
