package internalcmd

import (
	"encoding/json"
	"fmt"
	"io"
)

// ExitCodeError carries the process exit code a failed command should
// terminate with (data-model.md's Error Code table, plus validate's own
// findings-based rule). cli.Execute extracts Code via errors.As; the
// JSON error envelope is already written to the command's output by
// WriteError before this is returned, so ExitCodeError's own Error()
// text is never shown to a user — only used by tests and as a defensive
// fallback.
type ExitCodeError struct {
	Code int
}

func (e *ExitCodeError) Error() string {
	return fmt.Sprintf("internalcmd: exit code %d", e.Code)
}

// cliError is the shape WriteError marshals for {"error": {"code",
// "message"}} (§7).
type cliError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteSuccess marshals {"ok": true, <fields...>} to w as one
// newline-terminated JSON object. fields supplies every key beyond
// "ok" — a single key for most commands ("entity", "parent", ...), more
// than one for validate ("valid" + "findings") and status ("counts" +
// "specs" + "structural_errors"). Every command's RunE calls this
// exactly once on success (research.md; refined from the original
// single-key draft during implementation — see contracts/cli.md's
// reconciliation).
func WriteSuccess(w io.Writer, fields map[string]any) error {
	obj := make(map[string]any, len(fields)+1)
	obj["ok"] = true
	for k, v := range fields {
		obj[k] = v
	}
	return json.NewEncoder(w).Encode(obj)
}

// WriteError classifies err (via classify), marshals {"ok": false,
// "error": {"code", "message"}} to w, and returns an *ExitCodeError
// carrying the matching exit code — never nil (barring an encoding
// failure), so a command's RunE can simply `return
// internalcmd.WriteError(w, err)` on any failure. Every command's RunE
// calls this exactly once on failure (research.md).
func WriteError(w io.Writer, err error) error {
	code, exitCode := classify(err)
	obj := map[string]any{
		"ok":    false,
		"error": cliError{Code: code, Message: err.Error()},
	}
	if encErr := json.NewEncoder(w).Encode(obj); encErr != nil {
		return encErr
	}
	return &ExitCodeError{Code: exitCode}
}
