package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/validation"
)

// NewValidateCmd builds "misterspec internal validate [<id>]"
// (docs/architecture-specification.md §16-17). With no argument it
// calls validation.ValidateProject; with one, validation.ValidateEntity
// — never both, never a third implementation (spec.md FR-009).
//
// Its exit code is the one command-specific exception to classify's
// error-path table: a successful run reporting valid: false still exits
// 4, even though ok stays true (research.md's dedicated decision,
// reconciling §16's "ok ≠ valid" distinction with §8's reserved
// structural-validation-failure exit code).
func NewValidateCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "validate [<id>]",
		Short:         "Run structural validation",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			var findings []validation.Finding
			if len(args) == 1 {
				findings, err = validation.ValidateEntity(proj.Root, proj.Config, args[0])
			} else {
				findings, err = validation.ValidateProject(proj.Root, proj.Config)
			}
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			out := make([]map[string]any, 0, len(findings))
			for _, f := range findings {
				out = append(out, map[string]any{
					"code":     f.Code,
					"severity": severityString(f.Severity),
					"path":     f.Path,
					"message":  f.Message,
				})
			}

			valid := len(findings) == 0
			if writeErr := WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"valid":    valid,
				"findings": out,
			}); writeErr != nil {
				return writeErr
			}
			if !valid {
				return &ExitCodeError{Code: 4}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}

// severityString renders a validation.Severity as its JSON string form.
// Only SeverityError exists today (internal/validation's own doc
// comment); this stays a switch so a future warning severity is a
// one-line addition, not a signature change.
func severityString(s validation.Severity) string {
	switch s {
	case validation.SeverityError:
		return "error"
	default:
		return "error"
	}
}
