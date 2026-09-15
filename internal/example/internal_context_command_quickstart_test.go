// This file compiles and runs specs/017-internal-context-command/
// quickstart.md's end-to-end flow (a cold-cache first invocation, a
// repeated invocation proving determinism, an unknown-target error, an
// unsupported-intent error, a non-numeric-budget error, a deliberately
// tiny budget proving mandatory content still wins, and a --render
// invocation) against the actual misterspec internal context command
// (internal/cli/internalcmd.NewContextCmd), on top of the same fixture
// ranking_budgeting_quickstart_test.go (016) already builds.
package example

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func runContextCmd(args []string) (output string, exitCode int) {
	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs(args)
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
	return buf.String(), -1
}

func TestInternalContextCommandQuickstart_EndToEnd(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/memory/constitution.md",
		"---\ntype: constitution\n---\n## Principles\n\nFilesystem is the source of truth.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md",
		"---\nid: FEAT-004\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-011/spec.md",
		"---\nid: SPEC-011\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nRefresh token rotation details.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-auth.md",
		"---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n## Summary\n\nAuthentication model facts.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on:\n  - SPEC-011\nsupersedes: []\n---\n## Intent\n\nSee [[KNOW-003]] for background.\n")

	type resultShape struct {
		OK      bool `json:"ok"`
		Context struct {
			Budget         int     `json:"budget"`
			BudgetExceeded bool    `json:"budget_exceeded"`
			Rendered       *string `json:"rendered"`
			Items          []struct {
				Path string `json:"path"`
			} `json:"items"`
		} `json:"context"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decode := func(t *testing.T, output string) resultShape {
		t.Helper()
		var r resultShape
		if err := json.Unmarshal([]byte(output), &r); err != nil {
			t.Fatalf("output is not valid JSON: %v (%s)", err, output)
		}
		return r
	}

	// 1. First invocation against a project with no cache yet
	// (quickstart.md §1): succeeds on the very first call.
	out1, code1 := runContextCmd([]string{"SPEC-014", "--intent", "implementation", "--query", "refresh token rotation", "--dir", root})
	if code1 != 0 {
		t.Fatalf("first invocation exitCode = %d, want 0 (output: %s)", code1, out1)
	}
	r1 := decode(t, out1)
	if len(r1.Context.Items) == 0 {
		t.Fatal("first invocation returned no items")
	}

	// 2. Repeating the same request is byte-for-byte identical
	// (quickstart.md §2, FR-009).
	out2, code2 := runContextCmd([]string{"SPEC-014", "--intent", "implementation", "--query", "refresh token rotation", "--dir", root})
	if code2 != 0 {
		t.Fatalf("second invocation exitCode = %d, want 0 (output: %s)", code2, out2)
	}
	if out1 != out2 {
		t.Errorf("repeated invocation output differs:\n1: %s\n2: %s", out1, out2)
	}

	// 3. An unknown target (quickstart.md §3).
	out3, code3 := runContextCmd([]string{"SPEC-999", "--dir", root})
	if code3 != 3 {
		t.Fatalf("unknown-target exitCode = %d, want 3 (output: %s)", code3, out3)
	}
	if decode(t, out3).Error.Code != "entity_not_found" {
		t.Errorf("unknown-target error.code = %q, want entity_not_found", decode(t, out3).Error.Code)
	}

	// 4. An unsupported intent (quickstart.md §4).
	out4, code4 := runContextCmd([]string{"SPEC-014", "--intent", "bogus", "--dir", root})
	if code4 != 2 {
		t.Fatalf("unsupported-intent exitCode = %d, want 2 (output: %s)", code4, out4)
	}
	if decode(t, out4).Error.Code != "unsupported_intent" {
		t.Errorf("unsupported-intent error.code = %q, want unsupported_intent", decode(t, out4).Error.Code)
	}

	// 5. A non-numeric budget (quickstart.md §5).
	out5, code5 := runContextCmd([]string{"SPEC-014", "--budget", "notanumber", "--dir", root})
	if code5 != 2 {
		t.Fatalf("non-numeric-budget exitCode = %d, want 2 (output: %s)", code5, out5)
	}
	if decode(t, out5).Error.Code != "invalid_argument" {
		t.Errorf("non-numeric-budget error.code = %q, want invalid_argument", decode(t, out5).Error.Code)
	}

	// 6. A deliberately tiny budget — mandatory content still wins
	// (quickstart.md §6, 016 FR-006/FR-007).
	out6, code6 := runContextCmd([]string{"SPEC-014", "--budget", "10", "--dir", root})
	if code6 != 0 {
		t.Fatalf("tiny-budget exitCode = %d, want 0 (output: %s)", code6, out6)
	}
	r6 := decode(t, out6)
	if !r6.Context.BudgetExceeded {
		t.Error("tiny-budget context.budget_exceeded = false, want true")
	}

	// 7. Rendered mode (quickstart.md §7): every item present in
	// context.items also appears in context.rendered.
	out7, code7 := runContextCmd([]string{"SPEC-014", "--intent", "implementation", "--render", "--dir", root})
	if code7 != 0 {
		t.Fatalf("render exitCode = %d, want 0 (output: %s)", code7, out7)
	}
	r7 := decode(t, out7)
	if r7.Context.Rendered == nil || *r7.Context.Rendered == "" {
		t.Fatal("render context.rendered is nil/empty, want a Markdown pack")
	}
	for _, item := range r7.Context.Items {
		if !strings.Contains(*r7.Context.Rendered, item.Path) {
			t.Errorf("context.rendered does not mention item path %q", item.Path)
		}
	}
}
