package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewMigrationCheckTasksCmd builds "misterspec internal
// migration-check-tasks" — a read-only diagnostic listing every Task
// number claimed by more than one Spec project-wide
// (031-canonical-task-identity spec.md FR-007, FR-008,
// contracts/task-identity-resolution.md §4). It never writes to the
// filesystem; exit code 0 always, since a reported collision is
// informational, not a failure.
func NewMigrationCheckTasksCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "migration-check-tasks",
		Short:         "List Task numbers claimed by more than one Spec (read-only)",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			reports, err := operations.TaskIdentityMigrationDiagnostic(proj.Root, proj.Config)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			collisions := make([]map[string]any, 0, len(reports))
			for _, r := range reports {
				collisions = append(collisions, map[string]any{
					"task_number": r.TaskNumber,
					"specs":       idStrings(r.Specs),
					"paths":       r.Paths,
				})
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{"collisions": collisions})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}
