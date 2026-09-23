package internalcmd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func writeEvalRetrievalFixture(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nReferences [[KNOW-002]] directly.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-y.md",
		"---\nid: KNOW-002\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSome unrelated facts.\n")
}

func writeCaseFileFixture(t *testing.T, dir, name string, contents string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

type evalRetrievalResponse struct {
	OK            bool `json:"ok"`
	SchemaVersion int  `json:"schema_version"`
	Result        struct {
		Run struct {
			RunID   string            `json:"run_id"`
			Kind    string            `json:"kind"`
			Config  map[string]string `json:"config"`
			Results json.RawMessage   `json:"results"`
		} `json:"run"`
		Summary struct {
			Total  int `json:"total"`
			Passed int `json:"passed"`
			Failed int `json:"failed"`
		} `json:"summary"`
	} `json:"result"`
}

func TestEvalRetrievalCmd_Success(t *testing.T) {
	root := testutil.Project(t)
	writeEvalRetrievalFixture(t, root)

	casesDir := t.TempDir()
	writeCaseFileFixture(t, casesDir, "c1.yaml", `
id: c1
target: KNOW-001
dir: `+root+`
required: [KNOW-001]
`)

	cmd := internalcmd.NewEvalRetrievalCmd()
	cmd.SetArgs([]string{"--cases", casesDir})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded evalRetrievalResponse
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Fatalf("ok = false, want true (output: %s)", output)
	}
	if decoded.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", decoded.SchemaVersion)
	}
	if decoded.Result.Summary.Total != 1 || decoded.Result.Summary.Passed != 1 {
		t.Errorf("summary = %+v, want {total:1 passed:1}", decoded.Result.Summary)
	}
	if decoded.Result.Run.Kind != "retrieval" {
		t.Errorf("run.kind = %q, want retrieval", decoded.Result.Run.Kind)
	}
}

func TestEvalRetrievalCmd_MissingCasesFlagIsInvalidArgument(t *testing.T) {
	cmd := internalcmd.NewEvalRetrievalCmd()
	cmd.SetArgs([]string{})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}

func TestEvalRetrievalCmd_NonexistentCasesDirIsInvalidCase(t *testing.T) {
	cmd := internalcmd.NewEvalRetrievalCmd()
	cmd.SetArgs([]string{"--cases", filepath.Join(t.TempDir(), "does-not-exist")})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_case")
}

func TestEvalRetrievalCmd_MissingRequiredIdentifierFailsCaseNotCommand(t *testing.T) {
	root := testutil.Project(t)
	writeEvalRetrievalFixture(t, root)

	casesDir := t.TempDir()
	writeCaseFileFixture(t, casesDir, "c1.yaml", `
id: c1
target: KNOW-001
dir: `+root+`
required: [KNOW-001, KNOW-999]
`)

	cmd := internalcmd.NewEvalRetrievalCmd()
	cmd.SetArgs([]string{"--cases", casesDir})
	output, exitCode := runCmd(cmd)
	// A required identifier that does not exist at all is a malformed
	// case file, not a runtime retrieval failure (contract §1: "A
	// case's own dir does not resolve" sibling condition).
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_case")
}

func TestEvalRetrievalCmd_OutWritesFullRunRecord(t *testing.T) {
	root := testutil.Project(t)
	writeEvalRetrievalFixture(t, root)

	casesDir := t.TempDir()
	writeCaseFileFixture(t, casesDir, "c1.yaml", `
id: c1
target: KNOW-001
dir: `+root+`
required: [KNOW-001]
`)

	outPath := filepath.Join(t.TempDir(), "run.json")
	cmd := internalcmd.NewEvalRetrievalCmd()
	cmd.SetArgs([]string{"--cases", casesDir, "--out", outPath})
	_, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("--out file not written: %v", err)
	}
}

func TestEvalRetrievalCmd_RepeatedRunsAreIdentical(t *testing.T) {
	root := testutil.Project(t)
	writeEvalRetrievalFixture(t, root)

	casesDir := t.TempDir()
	writeCaseFileFixture(t, casesDir, "c1.yaml", `
id: c1
target: KNOW-001
dir: `+root+`
required: [KNOW-001]
forbidden: [KNOW-002]
`)

	run := func() evalRetrievalResponse {
		cmd := internalcmd.NewEvalRetrievalCmd()
		cmd.SetArgs([]string{"--cases", casesDir})
		output, exitCode := runCmd(cmd)
		if exitCode != 0 {
			t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
		}
		var decoded evalRetrievalResponse
		if err := json.Unmarshal([]byte(output), &decoded); err != nil {
			t.Fatalf("output not valid JSON: %v", err)
		}
		return decoded
	}

	r1 := run()
	r2 := run()
	if string(r1.Result.Run.Results) != string(r2.Result.Run.Results) {
		t.Fatalf("repeated runs produced different results:\n%s\nvs\n%s", r1.Result.Run.Results, r2.Result.Run.Results)
	}
	if r1.Result.Summary != r2.Result.Summary {
		t.Fatalf("repeated runs produced different summaries: %+v vs %+v", r1.Result.Summary, r2.Result.Summary)
	}
}
