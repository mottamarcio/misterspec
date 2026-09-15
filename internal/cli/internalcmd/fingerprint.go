package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewFingerprintCmd builds "misterspec internal fingerprint <path>"
// (docs/architecture-specification.md §13). It calls
// operations.Fingerprint directly — no logic of its own beyond argument
// parsing and JSON shaping (spec.md FR-009).
func NewFingerprintCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "fingerprint <path>",
		Short:         "Calculate a source file's content fingerprint",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			fp, err := operations.Fingerprint(proj.Root, args[0])
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"source": map[string]any{
					"path":        args[0],
					"algorithm":   fp.Algorithm,
					"fingerprint": fp.String(),
				},
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}
