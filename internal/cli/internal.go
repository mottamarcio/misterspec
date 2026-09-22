package cli

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
)

// newInternalCmd builds the "internal" parent *cobra.Command and
// attaches every internalcmd.NewXxxCmd() constructor as a subcommand —
// the single place the operation commands are wired into the tree
// (User Story 1, 008-cli-cobra). Hidden: true is set separately (User
// Story 3). references/backlinks added in 012-references-backlinks.
func newInternalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "internal",
		Short:  "Machine-facing deterministic operations (hidden from normal use)",
		Hidden: true,
	}
	cmd.AddCommand(
		internalcmd.NewResolveCmd(),
		internalcmd.NewInspectCmd(),
		internalcmd.NewParentCmd(),
		internalcmd.NewChildrenCmd(),
		internalcmd.NewCreateCmd(),
		internalcmd.NewCreateArtifactCmd(),
		internalcmd.NewFingerprintCmd(),
		internalcmd.NewInventoryCmd(),
		internalcmd.NewValidateCmd(),
		internalcmd.NewStatusCmd(),
		internalcmd.NewReferencesCmd(),
		internalcmd.NewBacklinksCmd(),
		internalcmd.NewContextCmd(),
		internalcmd.NewCommitsSinceFileCmd(),
		internalcmd.NewMigrationCheckTasksCmd(),
		internalcmd.NewPrepareCmd(),
		internalcmd.NewCaptureEvidenceCmd(),
		internalcmd.NewEvalRetrievalCmd(),
		internalcmd.NewEvalCompareCmd(),
		internalcmd.NewAnalyzeImpactCmd(),
	)
	return cmd
}
