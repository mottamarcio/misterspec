package internalcmd_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func gitOutput(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

func gitCommit(t *testing.T, root, message string) string {
	t.Helper()
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add -A: %v\n%s", err, out)
	}
	cmd = exec.Command("git", "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", message)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	return gitOutput(t, root, "rev-parse", "--short", "HEAD")
}

type analyzeImpactEnvelope struct {
	OK        bool `json:"ok"`
	ChangeSet struct {
		From              string `json:"from"`
		To                string `json:"to"`
		UnmappedCodePaths int    `json:"unmapped_code_paths"`
		Elements          []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"elements"`
	} `json:"change_set"`
	AffectedItems []struct {
		ID             string `json:"id"`
		Classification string `json:"classification"`
	} `json:"affected_items"`
	NoKnownRelationElements []string `json:"no_known_relation_elements"`
}

// TestAnalyzeImpactCmd_SuccessEnvelopeShape is 042-impact-analysis-
// review T016 (US1): the success envelope's shape matches contracts
// §4's example.
func TestAnalyzeImpactCmd_SuccessEnvelopeShape(t *testing.T) {
	root := testutil.Project(t)
	specDir := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014"
	testutil.WriteFile(t, root, specDir+"/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: draft\n---\n### R1 — First\n\nOriginal.\n")
	testutil.InitGitRepo(t, root)
	from := gitOutput(t, root, "rev-parse", "--short", "HEAD")

	testutil.WriteFile(t, root, specDir+"/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: draft\n---\n### R1 — First\n\nCHANGED.\n")
	gitCommit(t, root, "change requirement")

	cmd := internalcmd.NewAnalyzeImpactCmd()
	cmd.SetArgs([]string{"--from", from, "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded analyzeImpactEnvelope
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Fatalf("ok = false, want true (output: %s)", output)
	}
	if decoded.ChangeSet.From != from {
		t.Errorf("change_set.from = %q, want %q", decoded.ChangeSet.From, from)
	}
	if decoded.ChangeSet.To == "" {
		t.Error("change_set.to is empty, want \"working-tree\" or a resolved revision")
	}
	if decoded.AffectedItems == nil {
		t.Error("affected_items is null, want a (possibly empty) array")
	}
	if decoded.NoKnownRelationElements == nil {
		t.Error("no_known_relation_elements is null, want a (possibly empty) array")
	}
}

// TestAnalyzeImpactCmd_RevisionNotFound is 042-impact-analysis-review
// T016.
func TestAnalyzeImpactCmd_RevisionNotFound(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)

	cmd := internalcmd.NewAnalyzeImpactCmd()
	cmd.SetArgs([]string{"--from", "does-not-exist", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode == 0 {
		t.Fatalf("exitCode = 0, want non-zero (output: %s)", output)
	}
	assertErrorCode(t, output, "revision_not_found")
}

// TestAnalyzeImpactCmd_NotARepository is 042-impact-analysis-review
// T016.
func TestAnalyzeImpactCmd_NotARepository(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewAnalyzeImpactCmd()
	cmd.SetArgs([]string{"--from", "HEAD", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode == 0 {
		t.Fatalf("exitCode = 0, want non-zero (output: %s)", output)
	}
	assertErrorCode(t, output, "not_a_repository")
}
