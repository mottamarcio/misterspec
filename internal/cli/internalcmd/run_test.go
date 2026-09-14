package internalcmd_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
)

// runCmd executes cmd (already configured via SetArgs), returning
// everything it wrote to stdout/stderr and the exit code its RunE
// ultimately produced — 0 for a nil error, or the ExitCodeError's Code
// for a failure. Shared by every command's test file so each one only
// asserts on JSON shape and exit code, never Cobra's own plumbing.
func runCmd(cmd *cobra.Command) (output string, exitCode int) {
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		return buf.String(), 0
	}
	var exitErr *internalcmd.ExitCodeError
	if errors.As(err, &exitErr) {
		return buf.String(), exitErr.Code
	}
	return buf.String(), -1 // unexpected: not an *ExitCodeError
}

// assertErrorCode fails t unless output decodes as {"ok":false,
// "error":{"code": wantCode, ...}} — the shared shape every failing
// command's test checks.
func assertErrorCode(t *testing.T, output, wantCode string) {
	t.Helper()

	var decoded struct {
		OK    bool `json:"ok"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.OK {
		t.Fatalf("ok = true, want false (output: %s)", output)
	}
	if decoded.Error.Code != wantCode {
		t.Errorf("error.code = %q, want %q", decoded.Error.Code, wantCode)
	}
	if decoded.Error.Message == "" {
		t.Error("error.message is empty, want a concise explanation")
	}
}
