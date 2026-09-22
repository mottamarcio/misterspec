package internalcmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewCaptureEvidenceCmd builds "misterspec internal capture-evidence
// SPEC-### --task TASK-###" (041-task-evidence-fingerprint contracts
// §4). It resolves the Spec and Task exactly as "internal prepare"
// does, calls operations.CaptureEvidence, and renders the returned
// evidence.EvidenceFields as JSON — it never writes into tasks.md
// itself (that stays /mister-implement's own job, research.md #4).
func NewCaptureEvidenceCmd() *cobra.Command {
	var dir, task, origin, by, result, command, execDir string
	var args []string
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:           "capture-evidence SPEC-### --task TASK-###",
		Short:         "Compute a Task's content fingerprint, Git snapshot, and (automated) verification result",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, cmdArgs []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			loc, err := operations.Resolve(proj.Root, proj.Config, cmdArgs[0])
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}
			if loc.ID.Type != ids.Spec {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %s is not a Spec", operations.ErrInvalidTarget, cmdArgs[0]))
			}
			spec := loc.ID
			specDir := strings.TrimSuffix(loc.Path, "/spec.md")

			if task == "" {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: --task is required", ErrInvalidArgument))
			}
			taskNum, parseErr := parseTaskNumberArg(task, spec, proj.Config)
			if parseErr != nil {
				return WriteError(cmd.OutOrStdout(), parseErr)
			}
			taskID := ids.EntityID{Type: ids.Task, Prefix: ids.Task.Prefix(), Number: taskNum, Width: proj.Config.IDWidth}

			fields, err := operations.CaptureEvidence(proj.Root, proj.Config, operations.EvidenceCaptureRequest{
				Spec: spec, Task: taskID, SpecDir: specDir,
				Origin: origin, By: by, Result: result,
				Command: command, Args: args, Dir: execDir, Timeout: timeout,
			})
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"evidence": renderEvidenceFields(fields),
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	cmd.Flags().StringVar(&task, "task", "", "the Task to capture evidence for (bare TASK-### or composite SPEC-###:TASK-###)")
	cmd.Flags().StringVar(&origin, "origin", "", "\"automated\" or \"declared\"")
	cmd.Flags().StringVar(&by, "by", "", "the tool or person capturing this evidence")
	cmd.Flags().StringVar(&result, "result", "", "\"pass\" or \"fail\" (required for --origin declared)")
	cmd.Flags().StringVar(&command, "command", "", "the exact command to run (required for --origin automated)")
	cmd.Flags().StringArrayVar(&args, "args", nil, "one argument to the command (repeatable)")
	cmd.Flags().StringVar(&execDir, "exec-dir", ".", "working directory for --command, relative to the project root (correction found during implementation: named exec-dir, not dir, since --dir already means \"target project directory\" on every internal command)")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "timeout for --command")
	return cmd
}

// renderEvidenceFields shapes evidence.EvidenceFields as JSON-ready
// data per contracts §4 — empty/zero fields are still emitted as empty
// strings, matching the declared-origin example's own shape (no
// command/log keys when they don't apply).
func renderEvidenceFields(f evidence.EvidenceFields) map[string]any {
	return map[string]any{
		"task":              f.Task.String(),
		"result":            f.Result,
		"origin":            f.Origin,
		"by":                f.By,
		"captured_at":       f.CapturedAt,
		"command":           f.Command,
		"git_revision":      f.GitRevision,
		"working_tree":      f.WorkingTree,
		"input_fingerprint": f.Fingerprint,
		"log":               f.Log,
	}
}
