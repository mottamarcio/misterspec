package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/agents/builtin"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/installer"
	"github.com/mottamarcio/misterspec/kit"
)

// newInitCmd builds the public "init" command: --agent (required),
// --dir (default "."). Calls bootstrap.Bootstrap directly against the
// real builtin.Default() registry and the real kit.SkillsFS, and
// reports the result through internalcmd's envelope helper (User Story
// 2, research.md).
func newInitCmd() *cobra.Command {
	var dir, agent string

	cmd := &cobra.Command{
		Use:           "init",
		Short:         "Bootstrap a new misterspec project",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if agent == "" {
				return internalcmd.WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: --agent is required", internalcmd.ErrInvalidArgument))
			}

			outcome, err := bootstrap.Bootstrap(dir, agent, builtin.Default(), kit.SkillsFS)
			if err != nil {
				return internalcmd.WriteError(cmd.OutOrStdout(), err)
			}

			return internalcmd.WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"bootstrap": map[string]any{
					"project_root":   outcome.ProjectRoot,
					"config_written": outcome.ConfigWritten,
					"templates":      outcomesJSON(outcome.TemplateOutcomes),
					"agent": map[string]any{
						"adapter_id":       outcome.AgentInstall.AdapterID,
						"integration_path": outcome.AgentInstall.IntegrationPath,
						"outcomes":         outcomesJSON(outcome.AgentInstall.Outcomes),
					},
				},
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target directory to bootstrap")
	cmd.Flags().StringVar(&agent, "agent", "", "agent ID to install Skills for (required)")
	return cmd
}

// outcomeStatusString renders an installer.OutcomeStatus as its JSON
// string form.
func outcomeStatusString(s installer.OutcomeStatus) string {
	switch s {
	case installer.Installed:
		return "installed"
	case installer.Skipped:
		return "skipped"
	case installer.Failed:
		return "failed"
	default:
		return "unknown"
	}
}

// outcomesJSON renders []installer.Outcome as a non-nil slice of JSON
// objects — always an array, never null, even when empty.
func outcomesJSON(outcomes []installer.Outcome) []map[string]any {
	out := make([]map[string]any, 0, len(outcomes))
	for _, o := range outcomes {
		entry := map[string]any{
			"name":   o.Resource.Name,
			"kind":   o.Resource.Kind,
			"status": outcomeStatusString(o.Status),
			"path":   o.Path,
		}
		if o.Err != nil {
			entry["error"] = o.Err.Error()
		}
		out = append(out, entry)
	}
	return out
}
