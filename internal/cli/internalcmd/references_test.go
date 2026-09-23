package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestReferencesCmd_Success(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-011/spec.md", "---\nid: SPEC-011\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on: []\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-x.md", "---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on:\n  - SPEC-011\nsupersedes: []\n---\nSee [[KNOW-003]].\n")

	cmd := internalcmd.NewReferencesCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		Target     string `json:"target"`
		References struct {
			Formal []struct {
				Relation string `json:"relation"`
				Target   string `json:"target"`
			} `json:"formal"`
			Semantic []struct {
				Relation string `json:"relation"`
				Target   string `json:"target"`
			} `json:"semantic"`
		} `json:"references"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Target != "SPEC-014" {
		t.Errorf("target = %q, want SPEC-014", decoded.Target)
	}
	if len(decoded.References.Formal) != 2 {
		t.Fatalf("references.formal = %+v, want 2 entries", decoded.References.Formal)
	}
	if decoded.References.Formal[0].Relation != "parent" || decoded.References.Formal[0].Target != "FEAT-004" {
		t.Errorf("formal[0] = %+v, want relation=parent target=FEAT-004", decoded.References.Formal[0])
	}
	if decoded.References.Formal[1].Relation != "depends_on" || decoded.References.Formal[1].Target != "SPEC-011" {
		t.Errorf("formal[1] = %+v, want relation=depends_on target=SPEC-011", decoded.References.Formal[1])
	}
	if len(decoded.References.Semantic) != 1 || decoded.References.Semantic[0].Relation != "wikilink" || decoded.References.Semantic[0].Target != "KNOW-003" {
		t.Errorf("semantic = %+v, want [{wikilink KNOW-003}]", decoded.References.Semantic)
	}
}

func TestReferencesCmd_EmptyResult(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md", "---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n")

	cmd := internalcmd.NewReferencesCmd()
	cmd.SetArgs([]string{"KNOW-001", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		References struct {
			Formal   []any `json:"formal"`
			Semantic []any `json:"semantic"`
		} `json:"references"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.References.Formal == nil {
		t.Error("references.formal = null, want an empty array")
	}
	if decoded.References.Semantic == nil {
		t.Error("references.semantic = null, want an empty array")
	}
}

func TestReferencesCmd_NotFound(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewReferencesCmd()
	cmd.SetArgs([]string{"SPEC-999", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 3 {
		t.Fatalf("exitCode = %d, want 3 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "entity_not_found")
}

func TestReferencesCmd_InvalidTarget(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do it\n\n- [ ] Complete\n")

	cmd := internalcmd.NewReferencesCmd()
	cmd.SetArgs([]string{"TASK-001", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_target")
}

func TestReferencesCmd_EntriesIncludeSourceOccurrence(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-011/spec.md", "---\nid: SPEC-011\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on: []\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-x.md", "---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n")
	specPath := "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md"
	testutil.WriteFile(t, root, specPath,
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on:\n  - SPEC-011\nsupersedes: []\n---\n## Related Knowledge\n\nSee [[KNOW-003]].\n")

	cmd := internalcmd.NewReferencesCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		References struct {
			Formal []struct {
				Relation      string `json:"relation"`
				SourcePath    string `json:"source_path"`
				SourceSection string `json:"source_section"`
				SourceLine    int    `json:"source_line"`
			} `json:"formal"`
			Semantic []struct {
				Relation      string `json:"relation"`
				SourcePath    string `json:"source_path"`
				SourceSection string `json:"source_section"`
				SourceLine    int    `json:"source_line"`
			} `json:"semantic"`
		} `json:"references"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}

	if len(decoded.References.Semantic) != 1 {
		t.Fatalf("semantic = %+v, want 1 entry", decoded.References.Semantic)
	}
	sem := decoded.References.Semantic[0]
	if sem.SourcePath != specPath {
		t.Errorf("semantic source_path = %q, want %q", sem.SourcePath, specPath)
	}
	if sem.SourceSection != "Related Knowledge" {
		t.Errorf("semantic source_section = %q, want %q", sem.SourceSection, "Related Knowledge")
	}
	if sem.SourceLine <= 0 {
		t.Errorf("semantic source_line = %d, want a positive line", sem.SourceLine)
	}

	for _, f := range decoded.References.Formal {
		if f.SourcePath != specPath {
			t.Errorf("formal entry %+v: source_path = %q, want %q", f, f.SourcePath, specPath)
		}
		if f.SourceSection != "" || f.SourceLine != 0 {
			t.Errorf("formal entry %+v: want source_section empty and source_line 0", f)
		}
	}
}

func TestReferencesCmd_Ambiguous(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-002/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: draft\nparent: PRG-002\n---\n")

	cmd := internalcmd.NewReferencesCmd()
	cmd.SetArgs([]string{"FEAT-004", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 3 {
		t.Fatalf("exitCode = %d, want 3 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "entity_ambiguous")
}

// TestReferencesCmd_TargetAnchorPresentForAnchorQualifiedWikilink
// proves spec 040 contracts §4: an anchor-qualified wikilink's
// target_anchor is populated in the JSON output; a plain wikilink's
// stays empty.
func TestReferencesCmd_TargetAnchorPresentForAnchorQualifiedWikilink(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-x.md", "---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\n---\nSee [[KNOW-003#retry-policy]] and [[KNOW-003]].\n")

	cmd := internalcmd.NewReferencesCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		References struct {
			Semantic []struct {
				TargetAnchor string `json:"target_anchor"`
			} `json:"semantic"`
		} `json:"references"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if len(decoded.References.Semantic) != 2 {
		t.Fatalf("semantic = %+v, want 2 entries", decoded.References.Semantic)
	}
	if decoded.References.Semantic[0].TargetAnchor != "retry-policy" {
		t.Errorf("semantic[0].target_anchor = %q, want %q", decoded.References.Semantic[0].TargetAnchor, "retry-policy")
	}
	if decoded.References.Semantic[1].TargetAnchor != "" {
		t.Errorf("semantic[1].target_anchor = %q, want \"\" for a plain wikilink", decoded.References.Semantic[1].TargetAnchor)
	}
}
