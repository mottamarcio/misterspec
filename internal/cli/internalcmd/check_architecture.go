package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/architecture"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewCheckArchitectureCmd builds "misterspec internal
// check-architecture" (044-architecture-code-context-rules contracts
// §5). Evaluates proj.Config's own declared ArchitectureRules against
// the current Go source tree — read-only, never mutates the
// filesystem.
func NewCheckArchitectureCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "check-architecture",
		Short:         "Evaluate declared architecture rules against the current source tree",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			report, err := architecture.CheckArchitecture(proj.Root, proj.Config.ArchitectureRules, proj.Config.CodeExclusions)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"architecture": renderArchitectureReport(report),
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}

// renderArchitectureReport shapes one architecture.Report as JSON-
// ready data per contracts §5.
func renderArchitectureReport(report architecture.Report) map[string]any {
	results := make([]map[string]any, 0, len(report.Results))
	for _, r := range report.Results {
		results = append(results, map[string]any{
			"rule_index": r.RuleIndex,
			"status":     r.Status,
			"path":       r.Path,
			"line":       r.Line,
			"message":    r.Message,
			"reason":     r.Reason,
		})
	}
	return map[string]any{
		"adapter": report.Adapter,
		"results": results,
	}
}
