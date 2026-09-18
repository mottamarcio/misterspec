package internalcmd_test

import (
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestCommitsSinceFileCmd_OrdinaryCase(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)

	testutil.WriteFile(t, root, "plan.md", "the plan")
	runGitCmd(t, root, "add", "plan.md")
	runGitCmd(t, root, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", "add plan.md")

	cmd := internalcmd.NewCommitsSinceFileCmd()
	cmd.SetArgs([]string{"--path", "plan.md", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		OK        bool `json:"ok"`
		Available bool `json:"available"`
		Commits   []struct {
			Hash       string `json:"hash"`
			Subject    string `json:"subject"`
			AuthorDate string `json:"author_date"`
		} `json:"commits"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Error("ok = false, want true")
	}
	if !decoded.Available {
		t.Error("available = false, want true")
	}
	if len(decoded.Commits) != 1 {
		t.Fatalf("commits = %d, want 1: %+v", len(decoded.Commits), decoded.Commits)
	}
}

func TestCommitsSinceFileCmd_NotAvailable(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)

	cmd := internalcmd.NewCommitsSinceFileCmd()
	cmd.SetArgs([]string{"--path", "never-committed.md", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		OK        bool          `json:"ok"`
		Available bool          `json:"available"`
		Commits   []interface{} `json:"commits"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Error("ok = false, want true — not-available is not an error")
	}
	if decoded.Available {
		t.Error("available = true for a file never committed, want false")
	}
	if len(decoded.Commits) != 0 {
		t.Errorf("commits = %d, want 0", len(decoded.Commits))
	}
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
