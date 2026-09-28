package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewNextTaskIDCmd builds "misterspec internal next-task-id" — a
// read-only operation reporting the next available Task number,
// project-wide (045-global-task-numbering contracts/next-task-id-
// contract.md §1/§2). It never writes to the filesystem and never
// reserves the returned number in any persisted state (spec FR-003):
// calling it twice in a row with no intervening write returns the same
// value both times.
func NewNextTaskIDCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "next-task-id",
		Short:         "Report the next available Task number, project-wide (read-only)",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			next, err := ids.NextTaskID(proj.Root, proj.Config)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{"next_task_id": next.String()})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}
