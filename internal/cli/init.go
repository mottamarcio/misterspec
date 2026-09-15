package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/mottamarcio/misterspec/internal/agents/builtin"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/installer"
	"github.com/mottamarcio/misterspec/internal/tui"
	"github.com/mottamarcio/misterspec/kit"
)

// isInteractiveTerminal reports whether both stdin and stdout are
// attached to a real terminal — an overridable package variable
// (010-interactive-init-tui's research.md) so newInitCmd's own tests
// never depend on go test's actual TTY status.
var isInteractiveTerminal = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// tuiRunInit is tui.RunInit, as an overridable package variable so
// tests can substitute a stub that never launches a real Bubble Tea
// program (010-interactive-init-tui's research.md).
var tuiRunInit = tui.RunInit

// newInitCmd builds the public "init" command: --agent (required for
// the non-interactive path, checked explicitly inside RunE rather than
// via Cobra's MarkFlagRequired so its own failure still goes through
// the JSON envelope), --dir (default "."). With --agent provided, calls
// bootstrap.Bootstrap directly — 008-cli-cobra's own behavior,
// unchanged (SC-005). Without it, in an interactive terminal, launches
// the interactive flow (010-interactive-init-tui); without one, fails
// with the exact same JSON error shape a missing --agent already
// produced before this feature existed (research.md).
func newInitCmd() *cobra.Command {
	var dir, agent string

	cmd := &cobra.Command{
		Use:           "init",
		Short:         "Bootstrap a new misterspec project",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if agent != "" {
				outcome, err := bootstrap.Bootstrap(dir, agent, builtin.Default(), kit.SkillsFS)
				if err != nil {
					return internalcmd.WriteError(cmd.OutOrStdout(), err)
				}

				return internalcmd.WriteSuccess(cmd.OutOrStdout(), map[string]any{
					"bootstrap": map[string]any{
						"project_root":   outcome.ProjectRoot,
						"config_written": outcome.ConfigWritten,
						"directories":    outcome.DirectoriesScaffolded,
						"agent": map[string]any{
							"adapter_id":       outcome.AgentInstall.AdapterID,
							"integration_path": outcome.AgentInstall.IntegrationPath,
							"outcomes":         outcomesJSON(outcome.AgentInstall.Outcomes),
						},
					},
				})
			}

			if !isInteractiveTerminal() {
				return internalcmd.WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: --agent is required outside an interactive terminal", internalcmd.ErrInvalidArgument))
			}

			return tuiRunInit(dir, builtin.Default(), kit.SkillsFS)
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target directory to bootstrap")
	cmd.Flags().StringVar(&agent, "agent", "", "agent ID to install Skills for (non-interactive path)")
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
