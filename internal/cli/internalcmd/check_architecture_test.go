package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

type architectureResultJSON struct {
	RuleIndex int    `json:"rule_index"`
	Status    string `json:"status"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Message   string `json:"message"`
	Reason    string `json:"reason"`
}

type architectureEnvelopeJSON struct {
	OK           bool `json:"ok"`
	Architecture struct {
		Adapter string                   `json:"adapter"`
		Results []architectureResultJSON `json:"results"`
	} `json:"architecture"`
}

// TestCheckArchitectureCmd_ForbiddenDependencyReported is
// 044-architecture-code-context-rules T017 (US1): the CLI envelope
// surfaces a forbidden-dependency violation with its own file/line.
func TestCheckArchitectureCmd_ForbiddenDependencyReported(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, testutil.DefaultConfigYAML+
		"architecture_rules:\n  - kind: forbidden_dependency\n    from: a/**\n    to: b\n")
	testutil.WriteFile(t, root, "go.mod", "module fixture.example\n\ngo 1.23\n")
	testutil.WriteFile(t, root, "a/a.go", "package a\n\nimport \"fixture.example/b\"\n\nfunc UseB() { b.Thing() }\n")
	testutil.WriteFile(t, root, "b/b.go", "package b\n\nfunc Thing() {}\n")

	cmd := internalcmd.NewCheckArchitectureCmd()
	cmd.SetArgs([]string{"--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded architectureEnvelopeJSON
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Fatalf("ok = false, want true (output: %s)", output)
	}
	if decoded.Architecture.Adapter != "go" {
		t.Errorf("adapter = %q, want \"go\"", decoded.Architecture.Adapter)
	}

	var fails []architectureResultJSON
	for _, r := range decoded.Architecture.Results {
		if r.Status == "fail" {
			fails = append(fails, r)
		}
	}
	if len(fails) != 1 || fails[0].Path != "a/a.go" || fails[0].Line != 3 {
		t.Errorf("fails = %+v, want exactly one at a/a.go:3", fails)
	}
}

// TestCheckArchitectureCmd_NoRulesDeclaredIsEmptySuccess is
// 044-architecture-code-context-rules T017: zero declared rules is
// not an error.
func TestCheckArchitectureCmd_NoRulesDeclaredIsEmptySuccess(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewCheckArchitectureCmd()
	cmd.SetArgs([]string{"--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded architectureEnvelopeJSON
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Fatalf("ok = false, want true (output: %s)", output)
	}
	if decoded.Architecture.Results == nil {
		t.Error("results is null, want an empty (possibly zero-length) array")
	}
}
