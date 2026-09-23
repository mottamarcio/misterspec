// This file implements 039-lean-skills-integration-contracts's User
// Story 3 (FR-007/SC-004): for each of three representative agent
// integrations covering three distinct install target paths, install
// the real canonical Skills, then run the real
// `internal resolve` → `internal context` → `internal validate` chain
// mister-plan's own text documents against a small fixture Spec,
// asserting each step exits 0 with its documented JSON envelope shape,
// and that the Skill content actually landed at that adapter's own
// TargetPath(). No LLM/agent behavior is simulated — this is
// Skill-content-and-binary integration only.
package example

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/agents/builtin"
	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
	"github.com/mottamarcio/misterspec/kit"
	"github.com/spf13/cobra"
)

// smokeTestScenario is one representative-integration row
// (data-model.md "SmokeTestScenario").
type smokeTestScenario struct {
	AdapterID  string
	TargetPath string
}

// representativeSmokeScenarios is research.md #4's fixed set: three
// adapters covering three distinct install target directories — the
// minimum set that already covers every distinct target shape among
// the six registered adapters.
var representativeSmokeScenarios = []smokeTestScenario{
	{AdapterID: "claude-code", TargetPath: ".claude/skills"},
	{AdapterID: "cursor-agent", TargetPath: ".cursor/skills"},
	{AdapterID: "copilot", TargetPath: ".github/skills"},
}

func TestSkillSmoke(t *testing.T) {
	registry := builtin.Default()

	for _, scenario := range representativeSmokeScenarios {
		scenario := scenario
		t.Run(scenario.AdapterID, func(t *testing.T) {
			adapter, ok := registry.Get(scenario.AdapterID)
			if !ok {
				t.Fatalf("registry.Get(%q) ok = false, want true", scenario.AdapterID)
			}
			if adapter.TargetPath() != scenario.TargetPath {
				t.Fatalf("%s.TargetPath() = %q, want %q", scenario.AdapterID, adapter.TargetPath(), scenario.TargetPath)
			}

			root := testutil.Project(t)
			writeSkillSmokeFixture(t, root)

			result, err := adapter.Install(context.Background(), agents.InstallRequest{
				ProjectRoot: root,
				Skills:      kit.SkillsFS,
			})
			if err != nil {
				t.Fatalf("%s.Install() unexpected error: %v", scenario.AdapterID, err)
			}
			if len(result.Outcomes) == 0 {
				t.Fatalf("%s.Install() produced no outcomes", scenario.AdapterID)
			}
			installedPath := filepath.Join(root, scenario.TargetPath, "mister-plan", "SKILL.md")
			installed, err := os.ReadFile(installedPath)
			if err != nil {
				t.Fatalf("%s: Skill content did not land at its own TargetPath %s: %v", scenario.AdapterID, scenario.TargetPath, err)
			}
			want, err := fs.ReadFile(kit.SkillsFS, "mister-plan/SKILL.md")
			if err != nil {
				t.Fatalf("reading kit.SkillsFS mister-plan/SKILL.md: %v", err)
			}
			if string(installed) != string(want) {
				t.Errorf("%s: installed mister-plan/SKILL.md does not match kit.SkillsFS content", scenario.AdapterID)
			}

			// Step 1: internal resolve SPEC-014 --dir root.
			resolveOut, exitCode := runInternalCmd(internalcmd.NewResolveCmd(), []string{"SPEC-014", "--dir", root})
			if exitCode != 0 {
				t.Fatalf("%s: `internal resolve` step failed, exit %d: %s", scenario.AdapterID, exitCode, resolveOut)
			}
			assertOKEnvelope(t, scenario.AdapterID, "resolve", resolveOut)

			// Step 2: internal context SPEC-014 --dir root --intent planning.
			contextOut, exitCode := runInternalCmd(internalcmd.NewContextCmd(), []string{"SPEC-014", "--dir", root, "--intent", "planning"})
			if exitCode != 0 {
				t.Fatalf("%s: `internal context` step failed, exit %d: %s", scenario.AdapterID, exitCode, contextOut)
			}
			assertOKEnvelope(t, scenario.AdapterID, "context", contextOut)

			// Step 3: internal validate SPEC-014 --dir root.
			validateOut, exitCode := runInternalCmd(internalcmd.NewValidateCmd(), []string{"SPEC-014", "--dir", root})
			if exitCode != 0 {
				t.Fatalf("%s: `internal validate` step failed, exit %d: %s", scenario.AdapterID, exitCode, validateOut)
			}
			assertOKEnvelope(t, scenario.AdapterID, "validate", validateOut)
		})
	}
}

// TestSkillSmoke_SingleAdapterFailureDoesNotMaskOthers is T029's
// regression case (spec Acceptance Scenario US3.1): a failure in one
// row's own scenario data must be reported for that adapter+step only,
// without affecting the other rows' own pass/fail outcome.
func TestSkillSmoke_SingleAdapterFailureDoesNotMaskOthers(t *testing.T) {
	registry := builtin.Default()

	scenarios := []smokeTestScenario{
		{AdapterID: "claude-code", TargetPath: ".claude/skills"},
		{AdapterID: "cursor-agent", TargetPath: "wrong/target/path"}, // deliberately wrong
		{AdapterID: "copilot", TargetPath: ".github/skills"},
	}

	var failedAdapters []string
	for _, scenario := range scenarios {
		adapter, ok := registry.Get(scenario.AdapterID)
		if !ok {
			t.Fatalf("registry.Get(%q) ok = false, want true", scenario.AdapterID)
		}
		if adapter.TargetPath() != scenario.TargetPath {
			failedAdapters = append(failedAdapters, scenario.AdapterID)
		}
	}

	if len(failedAdapters) != 1 || failedAdapters[0] != "cursor-agent" {
		t.Errorf("failedAdapters = %v, want exactly [cursor-agent] — a single wrong scenario must be reported for that adapter alone", failedAdapters)
	}
}

// writeSkillSmokeFixture builds a minimal but real Spec/Plan/Tasks
// fixture — the same shape internal/cli/internalcmd/prepare_test.go's
// writePrepareFixture already establishes — sufficient for
// resolve/context/validate to all run successfully against SPEC-014.
func writeSkillSmokeFixture(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n"+
			"## Requirements\n\n"+
			"### R1 — Example requirement for the smoke test.\n\nDetailed rule for R1.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/plan.md",
		"---\ntype: plan\nfor: SPEC-014\nstatus: draft\n---\n"+
			"## Implementation Sequence\n\nCover R1.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/tasks.md",
		"---\ntype: tasks\nfor: SPEC-014\n---\n# Tasks\n\n"+
			"## TASK-001 — Example task\n\n"+
			"- [ ] Complete\n\n"+
			"Serves: SPEC-014:R1\n\n"+
			"Scope: internal/example/smoke.go\n\n"+
			"Verify: go test ./internal/example/... -run TestSkillSmoke\n")
}

// runInternalCmd executes cmd directly with args, mirroring
// internal/cli/internalcmd's own run_test.go runCmd convention (used
// identically across that package's tests).
func runInternalCmd(cmd *cobra.Command, args []string) (output string, exitCode int) {
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

// assertOKEnvelope confirms output decodes as JSON with a top-level
// "ok": true field — the documented envelope shape every `internal`
// command shares (Constitution Principle IX).
func assertOKEnvelope(t *testing.T, adapterID, step, output string) {
	t.Helper()
	var decoded struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("%s: `internal %s` output is not valid JSON: %v (%s)", adapterID, step, err, output)
	}
	if !decoded.OK {
		t.Errorf("%s: `internal %s` output has \"ok\": false, want true (%s)", adapterID, step, output)
	}
}
