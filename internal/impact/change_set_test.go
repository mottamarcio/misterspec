package impact

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// initFixtureRepo creates a real Git repository, mirroring
// internal/vcs's own test fixture idiom (real git, not a hand-rolled
// stand-in), with one Spec directory already in place.
func initFixtureRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "--allow-empty", "-q", "-m", "init")
	return dir
}

func writeFixtureFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", relPath, err)
	}
}

func commitFixture(t *testing.T, dir, message string) string {
	t.Helper()
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add -A: %v\n%s", err, out)
	}
	cmd = exec.Command("git", "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", message)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	return string(out[:len(out)-1])
}

func fixtureConfig() project.Configuration {
	return project.Configuration{
		IDWidth:      3,
		ProgramsRoot: "ai/programs",
		KnowledgeDir: "ai/knowledge",
		LearningsDir: "ai/memory/learnings",
	}
}

const specFrontmatter = "---\nid: SPEC-014\ntype: spec\nstatus: draft\n---\n"

// TestBuildChangeSet_WholeArtifactAddedModifiedRemoved is 042-impact-
// analysis-review T010 (Foundational): a Spec added, a separate
// Knowledge note modified, and a Learning note removed between two
// commits each produce exactly one whole-artifact ChangedElement
// (data-model.md "ChangedElement").
func TestBuildChangeSet_WholeArtifactAddedModifiedRemoved(t *testing.T) {
	dir := initFixtureRepo(t)
	specDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"
	writeFixtureFile(t, dir, "ai/knowledge/KNOW-003-auth.md", "v1")
	writeFixtureFile(t, dir, "ai/memory/learnings/LRN-005-perf.md", "v1")
	from := commitFixture(t, dir, "base")

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nBody.\n")
	writeFixtureFile(t, dir, "ai/knowledge/KNOW-003-auth.md", "v2")
	if err := os.Remove(filepath.Join(dir, "ai/memory/learnings/LRN-005-perf.md")); err != nil {
		t.Fatalf("removing learning file: %v", err)
	}
	to := commitFixture(t, dir, "change")

	cs, err := BuildChangeSet(dir, fixtureConfig(), from, to, "")
	if err != nil {
		t.Fatalf("BuildChangeSet() unexpected error: %v", err)
	}

	byPath := map[string]ChangedElement{}
	for _, e := range cs.Elements {
		if e.RequirementNumber == nil {
			byPath[e.Path] = e
		}
	}

	spec, ok := byPath[specDir+"/spec.md"]
	if !ok || spec.Status != Added || spec.ID.Type != ids.Spec || spec.ID.Number != 14 {
		t.Errorf("spec.md entry = %+v, want Added SPEC-014", spec)
	}
	know, ok := byPath["ai/knowledge/KNOW-003-auth.md"]
	if !ok || know.Status != Modified || know.ID.Type != ids.Knowledge || know.ID.Number != 3 {
		t.Errorf("KNOW-003 entry = %+v, want Modified KNOW-003", know)
	}
	learn, ok := byPath["ai/memory/learnings/LRN-005-perf.md"]
	if !ok || learn.Status != Removed || learn.ID.Type != ids.Learning || learn.ID.Number != 5 {
		t.Errorf("LRN-005 entry = %+v, want Removed LRN-005", learn)
	}
}

// TestBuildChangeSet_RequirementLevelChanges is 042-impact-analysis-
// review T010: a modified spec.md yields one ChangedElement per
// Requirement number whose own text changed, was added, or was removed
// (research.md #3) — a title-only or unrelated-prose edit to a
// Requirement that keeps its own body identical does not.
func TestBuildChangeSet_RequirementLevelChanges(t *testing.T) {
	dir := initFixtureRepo(t)
	specDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"
	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+
		"### R1 — First\n\nOriginal body.\n\n### R2 — Second\n\nUnchanged body.\n")
	from := commitFixture(t, dir, "base")

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+
		"### R1 — First\n\nCHANGED body.\n\n### R2 — Second\n\nUnchanged body.\n\n### R3 — Third\n\nNew requirement.\n")
	to := commitFixture(t, dir, "change requirement text")

	cs, err := BuildChangeSet(dir, fixtureConfig(), from, to, "")
	if err != nil {
		t.Fatalf("BuildChangeSet() unexpected error: %v", err)
	}

	byNumber := map[int]ChangedElement{}
	for _, e := range cs.Elements {
		if e.RequirementNumber != nil {
			byNumber[*e.RequirementNumber] = e
		}
	}

	if e, ok := byNumber[1]; !ok || e.Status != Modified {
		t.Errorf("R1 entry = %+v (ok=%v), want Modified", e, ok)
	}
	if _, ok := byNumber[2]; ok {
		t.Errorf("R2 unexpectedly reported changed — its body is identical between revisions")
	}
	if e, ok := byNumber[3]; !ok || e.Status != Added {
		t.Errorf("R3 entry = %+v (ok=%v), want Added", e, ok)
	}
}

// TestBuildChangeSet_PlanAndTasksAttributedToOwningSpec is 042-impact-
// analysis-review T010: a changed plan.md/tasks.md is attributed to its
// owning Spec's own ID (research.md #8) — no separate "Plan"/"Tasks"
// entity type is introduced.
func TestBuildChangeSet_PlanAndTasksAttributedToOwningSpec(t *testing.T) {
	dir := initFixtureRepo(t)
	specDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"
	writeFixtureFile(t, dir, specDir+"/plan.md", "v1")
	from := commitFixture(t, dir, "base")

	writeFixtureFile(t, dir, specDir+"/plan.md", "v2")
	to := commitFixture(t, dir, "change plan")

	cs, err := BuildChangeSet(dir, fixtureConfig(), from, to, "")
	if err != nil {
		t.Fatalf("BuildChangeSet() unexpected error: %v", err)
	}
	if len(cs.Elements) != 1 {
		t.Fatalf("Elements = %+v, want exactly 1", cs.Elements)
	}
	got := cs.Elements[0]
	if got.ID.Type != ids.Spec || got.ID.Number != 14 {
		t.Errorf("plan.md entry ID = %v, want SPEC-014", got.ID)
	}
}

// TestBuildChangeSet_UnmappedCodePath is 042-impact-analysis-review
// T010 (research.md #9): a changed path outside the five entity types
// and plan.md/tasks.md increments UnmappedCodePaths instead of
// appearing in Elements.
func TestBuildChangeSet_UnmappedCodePath(t *testing.T) {
	dir := initFixtureRepo(t)
	writeFixtureFile(t, dir, "internal/foo/foo.go", "package foo\n")
	from := commitFixture(t, dir, "base")

	writeFixtureFile(t, dir, "internal/foo/foo.go", "package foo\n\nfunc Bar() {}\n")
	to := commitFixture(t, dir, "change code")

	cs, err := BuildChangeSet(dir, fixtureConfig(), from, to, "")
	if err != nil {
		t.Fatalf("BuildChangeSet() unexpected error: %v", err)
	}
	if cs.UnmappedCodePaths != 1 {
		t.Errorf("UnmappedCodePaths = %d, want 1", cs.UnmappedCodePaths)
	}
	if len(cs.Elements) != 0 {
		t.Errorf("Elements = %+v, want none for an unmapped code-only change", cs.Elements)
	}
}

// TestBuildChangeSet_NotARepository is 042-impact-analysis-review T010.
func TestBuildChangeSet_NotARepository(t *testing.T) {
	dir := t.TempDir()
	if _, err := BuildChangeSet(dir, fixtureConfig(), "HEAD", "", ""); err == nil {
		t.Error("BuildChangeSet() error = nil outside a Git repository, want ErrNotARepository")
	}
}
