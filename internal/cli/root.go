package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/agents/builtin"
	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/selfupdate"
	"github.com/mottamarcio/misterspec/kit"
)

// goosOverride/goarchOverride are the running process's own GOOS/GOARCH,
// read once as package vars purely so tests can build the exact asset
// name --update would look for without duplicating runtime.GOOS/GOARCH
// literals (specs/030-cli-version-update).
var (
	goosOverride   = runtime.GOOS
	goarchOverride = runtime.GOARCH
)

// httpClient is the *http.Client --update uses for every GitHub call —
// an overridable package variable (mirroring isInteractiveTerminal's own
// seam in init.go) so tests substitute a fake transport and never make a
// real network call (research.md).
var httpClient = http.DefaultClient

// executablePath is os.Executable, as an overridable package variable so
// tests can point --update at a scratch file instead of the real running
// test binary.
var executablePath = os.Executable

// updateWorkDir is os.Getwd, as an overridable package variable so tests
// control which directory --update checks for an existing
// .misterspec/install.json.
var updateWorkDir = os.Getwd

// confirmUpdate prompts for and returns the user's own confirmation
// before --update changes anything (FR-005) — an overridable package
// variable so tests never block on real stdin.
var confirmUpdate = defaultConfirmUpdate

func defaultConfirmUpdate(current, latest string) bool {
	if !isInteractiveTerminal() {
		return false
	}
	fmt.Fprintf(os.Stdout, "misterspec %s is available (current: %s). Update now? [y/N] ", latest, current)
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

// newRootCmd builds the root *cobra.Command. --help shows "init",
// "--version", and "--update" (and Cobra's own built-in help/completion)
// — the "internal" tree is attached but Hidden: true (FR-005, User Story
// 3; internal.go). --version and --update are the Constitution's own
// v1.1.0-amended public command surface (specs/030-cli-version-update).
func newRootCmd() *cobra.Command {
	var showVersion, doUpdate bool

	root := &cobra.Command{
		Use:           "misterspec",
		Short:         "misterspec — spec-driven development framework",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case showVersion:
				return internalcmd.WriteSuccess(cmd.OutOrStdout(), map[string]any{
					"version": selfupdate.Report(),
				})
			case doUpdate:
				return runUpdate(cmd)
			default:
				return cmd.Help()
			}
		},
	}

	root.PersistentFlags().BoolVar(&showVersion, "version", false, "print the installed misterspec version and exit")
	root.PersistentFlags().BoolVar(&doUpdate, "update", false, "check GitHub for a newer misterspec release and, if confirmed, install it")
	root.AddCommand(newInternalCmd())
	root.AddCommand(newInitCmd())
	return root
}

// runUpdate implements --update end to end: check GitHub, compare
// versions, require confirmation, verify a SHA-256 checksum, atomically
// replace the running binary, and — only inside an already-initialized
// project — reinstall Skills for its recorded agent integration, reusing
// the exact same agents.Adapter.Install path "misterspec init" already
// uses (never a second, parallel copy — Constitution Principle VI, DRY;
// research.md). --update is the sole command that ever calls
// selfupdate.LatestRelease/Download — the one deliberate exception to
// misterspec's otherwise network-free command surface (Constitution
// v1.1.0).
func runUpdate(cmd *cobra.Command) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	current := selfupdate.Report()

	rel, err := selfupdate.LatestRelease(ctx, httpClient)
	if err != nil {
		return internalcmd.WriteError(cmd.OutOrStdout(), err)
	}

	if !selfupdate.Newer(current, rel.TagName) {
		return internalcmd.WriteSuccess(cmd.OutOrStdout(), updateResult(current, rel.TagName, "up_to_date", nil))
	}

	if !confirmUpdate(current, rel.TagName) {
		return internalcmd.WriteSuccess(cmd.OutOrStdout(), updateResult(current, rel.TagName, "declined", nil))
	}

	asset, ok := selfupdate.SelectAsset(rel, goosOverride, goarchOverride)
	if !ok {
		return internalcmd.WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %s/%s in release %s", selfupdate.ErrNoCompatibleAsset, goosOverride, goarchOverride, rel.TagName))
	}
	sumsAsset, ok := selfupdate.ChecksumsAsset(rel)
	if !ok {
		return internalcmd.WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: no SHA256SUMS asset in release %s", selfupdate.ErrNoCompatibleAsset, rel.TagName))
	}

	content, err := selfupdate.Download(ctx, httpClient, asset)
	if err != nil {
		return internalcmd.WriteError(cmd.OutOrStdout(), err)
	}
	sumsData, err := selfupdate.Download(ctx, httpClient, sumsAsset)
	if err != nil {
		return internalcmd.WriteError(cmd.OutOrStdout(), err)
	}
	sums, err := selfupdate.ParseChecksums(sumsData)
	if err != nil {
		return internalcmd.WriteError(cmd.OutOrStdout(), err)
	}
	expected, ok := sums[asset.Name]
	if !ok || !selfupdate.VerifyChecksum(content, expected) {
		return internalcmd.WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %s", selfupdate.ErrChecksumMismatch, asset.Name))
	}

	exePath, err := executablePath()
	if err != nil {
		return internalcmd.WriteError(cmd.OutOrStdout(), err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(exePath), ".misterspec-update-*")
	if err != nil {
		return internalcmd.WriteError(cmd.OutOrStdout(), err)
	}
	tmpPath := tmpFile.Name()
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return internalcmd.WriteError(cmd.OutOrStdout(), err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return internalcmd.WriteError(cmd.OutOrStdout(), err)
	}

	if err := selfupdate.Replace(tmpPath, exePath, goosOverride); err != nil {
		os.Remove(tmpPath)
		return internalcmd.WriteError(cmd.OutOrStdout(), err)
	}

	skillsRefreshed := refreshProjectSkills()

	return internalcmd.WriteSuccess(cmd.OutOrStdout(), updateResult(current, rel.TagName, "updated", &skillsRefreshed))
}

// refreshProjectSkills checks the current working directory for an
// existing .misterspec/install.json and, if found, reinstalls Skills
// for its recorded agent via the exact same agents.Adapter.Install path
// "misterspec init" already uses (FR-010, FR-011). Absence of a project
// here is a normal, non-error outcome (FR-011) — false is returned, not
// an error.
func refreshProjectSkills() bool {
	workDir, err := updateWorkDir()
	if err != nil {
		return false
	}

	record, found, err := agents.CurrentInstall(workDir)
	if err != nil || !found {
		return false
	}

	adapter, ok := builtin.Default().Get(record.Agent.ID)
	if !ok {
		return false
	}

	_, err = adapter.Install(context.Background(), agents.InstallRequest{
		ProjectRoot: workDir,
		Skills:      kit.SkillsFS,
		Overwrite:   true,
	})
	return err == nil
}

// updateResult builds --update's own "update" JSON object per
// contracts/version-update.md — skillsRefreshed is nil for the
// no-download outcomes (up_to_date/declined), and always present
// (never absent) for "updated".
func updateResult(current, latest, status string, skillsRefreshed *bool) map[string]any {
	update := map[string]any{
		"current": current,
		"latest":  latest,
		"status":  status,
	}
	if skillsRefreshed != nil {
		update["skills_refreshed"] = *skillsRefreshed
	}
	return map[string]any{"update": update}
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
