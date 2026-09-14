package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
)

// newRootCmd builds the root *cobra.Command. --help shows only "init"
// (and Cobra's own built-in help/completion) — the "internal" tree is
// attached but Hidden: true (FR-005, User Story 3; internal.go).
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "misterspec",
		Short:         "misterspec — spec-driven development framework",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newInternalCmd())
	root.AddCommand(newInitCmd())
	return root
}

// Execute builds the root command tree, runs it against os.Args, and
// returns the process exit code cmd/misterspec/main.go should terminate
// with — 0 on success, or the code data-model.md's Error Code table (or
// validate's own findings-based rule, research.md) names otherwise.
//
// A failure that reaches Cobra itself before any command's RunE runs (an
// unknown command, an unrecognized flag, a missing required flag) is
// still reported the same structured JSON way every other failure is —
// never Cobra's own default "Error: ..." plus usage text, which would
// violate FR-010's no-decorative-output guarantee.
func Execute() int {
	root := newRootCmd()
	err := root.Execute()
	if err == nil {
		return 0
	}

	var exitErr *internalcmd.ExitCodeError
	if errors.As(err, &exitErr) {
		return exitErr.Code
	}

	// Cobra-level parsing failure — no command's RunE ran, so nothing
	// has written a JSON error yet.
	writeErr := internalcmd.WriteError(os.Stdout, fmt.Errorf("%w: %v", internalcmd.ErrInvalidArgument, err))
	var codeErr *internalcmd.ExitCodeError
	if errors.As(writeErr, &codeErr) {
		return codeErr.Code
	}
	return 1
}
