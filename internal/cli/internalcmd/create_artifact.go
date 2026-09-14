package internalcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// artifactKindNames maps a create-artifact positional argument to
// artifacts.ArtifactType — the three subordinate kinds Create does not
// handle (docs/architecture-specification.md §11-12).
var artifactKindNames = map[string]artifacts.ArtifactType{
	"plan":       artifacts.TypePlan,
	"tasks":      artifacts.TypeTasks,
	"validation": artifacts.TypeValidation,
}

// NewCreateArtifactCmd builds "misterspec internal create-artifact
// <kind> --for <specID>" (docs/architecture-specification.md §12). It
// calls operations.CreateArtifact directly — no logic of its own beyond
// argument/flag parsing and JSON shaping (spec.md FR-009).
func NewCreateArtifactCmd() *cobra.Command {
	var dir, forID string

	cmd := &cobra.Command{
		Use:           "create-artifact <kind>",
		Short:         "Create a canonical subordinate artifact (plan, tasks, validation) for a Spec",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, ok := artifactKindNames[args[0]]
			if !ok {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: unrecognized kind %q", ErrInvalidArgument, args[0]))
			}

			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			result, err := operations.CreateArtifact(proj.Root, proj.Config, operations.CreateArtifactRequest{
				Kind: kind,
				For:  forID,
			})
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"created": map[string]any{
					"type": kind.String(),
					"path": result.Path,
					"for":  forID,
				},
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	cmd.Flags().StringVar(&forID, "for", "", "the Spec ID this artifact belongs to (required)")
	return cmd
}
