package installer_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/mottamarcio/misterspec/internal/installer"
	"github.com/mottamarcio/misterspec/kit"
)

func fixtureSkillsFS() fstest.MapFS {
	return fstest.MapFS{
		"skills/one.md": {Data: []byte("# Skill One\n")},
		"skills/two.md": {Data: []byte("# Skill Two\n")},
	}
}

// fixtureNestedSkillsFS mirrors a real canonical Skill layout — one
// directory per Skill, each containing its own SKILL.md
// (009-canonical-skills-content's research.md) — the shape ListFS/
// InstallFS could not discover before this feature's recursive-walk
// change (a single-level fs.ReadDir explicitly skips directory
// entries). Rooted at "." (no "skills/" prefix inside the FS itself),
// matching exactly how kit.SkillsFS (already fs.Sub'd) and
// claude.Install's own sourceDir "." already use it — this is the
// actual production shape, not the "skills/"-prefixed layout
// kit.TemplatesFS uses.
func fixtureNestedSkillsFS() fstest.MapFS {
	return fstest.MapFS{
		"create-plan/SKILL.md":  {Data: []byte("# create-plan\n")},
		"create-tasks/SKILL.md": {Data: []byte("# create-tasks\n")},
	}
}

func TestList_ReturnsEveryTemplate(t *testing.T) {
	resources := installer.List()
	if len(resources) != 8 {
		t.Fatalf("List() returned %d resources, want 8: %+v", len(resources), resources)
	}

	var artifactTypes []string
	for _, r := range resources {
		if r.Kind != "template" {
			t.Errorf("Resource %q Kind = %q, want %q", r.Name, r.Kind, "template")
		}
		artifactTypes = append(artifactTypes, r.ArtifactType)
	}
	sort.Strings(artifactTypes)

	want := []string{"feature", "knowledge", "learning", "plan", "program", "spec", "tasks", "validation"}
	if len(artifactTypes) != len(want) {
		t.Fatalf("ArtifactTypes = %v, want %v", artifactTypes, want)
	}
	for i := range want {
		if artifactTypes[i] != want[i] {
			t.Errorf("ArtifactTypes = %v, want %v", artifactTypes, want)
			break
		}
	}
}

func TestList_ArtifactTypeDerivedFromFilename(t *testing.T) {
	resources := installer.List()
	for _, r := range resources {
		if r.Name == "program.md.tmpl" {
			if r.ArtifactType != "program" {
				t.Errorf("program.md.tmpl ArtifactType = %q, want %q", r.ArtifactType, "program")
			}
			return
		}
	}
	t.Fatal("List() did not include program.md.tmpl")
}

func TestList_IdenticalAcrossRepeatedCalls(t *testing.T) {
	first := installer.List()
	second := installer.List()
	if len(first) != len(second) {
		t.Fatalf("List() call 1 = %d resources, call 2 = %d", len(first), len(second))
	}
	seen := map[string]bool{}
	for _, r := range first {
		seen[r.Name] = true
	}
	for _, r := range second {
		if !seen[r.Name] {
			t.Errorf("List() second call included %q, not present in the first call", r.Name)
		}
	}
}

func TestInstall_FreshDirectory(t *testing.T) {
	target := t.TempDir()

	outcomes, err := installer.Install(target, false)
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}
	if len(outcomes) != 8 {
		t.Fatalf("Install() = %d outcomes, want 8: %+v", len(outcomes), outcomes)
	}

	for _, o := range outcomes {
		if o.Status != installer.Installed {
			t.Errorf("Outcome for %q Status = %v, want Installed (Err: %v)", o.Resource.Name, o.Status, o.Err)
			continue
		}
		got, err := os.ReadFile(filepath.Join(target, o.Path))
		if err != nil {
			t.Fatalf("reading installed file %s: %v", o.Path, err)
		}
		want, err := kit.TemplatesFS.ReadFile("templates/" + o.Resource.Name)
		if err != nil {
			t.Fatalf("reading embedded source %s: %v", o.Resource.Name, err)
		}
		if string(got) != string(want) {
			t.Errorf("installed file %s not byte-identical to embedded source", o.Path)
		}
	}
}

func TestInstall_AlreadyPopulatedWithoutOverwriteSkipsEverything(t *testing.T) {
	target := t.TempDir()

	if _, err := installer.Install(target, false); err != nil {
		t.Fatalf("first Install() unexpected error: %v", err)
	}

	// Record mtimes to prove nothing was rewritten.
	before := map[string]int64{}
	outcomes, _ := installer.Install(target, false)
	for _, o := range outcomes {
		info, err := os.Stat(filepath.Join(target, o.Path))
		if err != nil {
			t.Fatalf("stat %s: %v", o.Path, err)
		}
		before[o.Path] = info.ModTime().UnixNano()
	}

	outcomes, err := installer.Install(target, false)
	if err != nil {
		t.Fatalf("second Install() unexpected error: %v", err)
	}
	for _, o := range outcomes {
		if o.Status != installer.Skipped {
			t.Errorf("Outcome for %q Status = %v, want Skipped", o.Resource.Name, o.Status)
		}
		info, err := os.Stat(filepath.Join(target, o.Path))
		if err != nil {
			t.Fatalf("stat %s: %v", o.Path, err)
		}
		if info.ModTime().UnixNano() != before[o.Path] {
			t.Errorf("%s was modified despite overwrite:false", o.Path)
		}
	}
}

func TestInstall_OverwriteReplacesEverything(t *testing.T) {
	target := t.TempDir()

	if _, err := installer.Install(target, false); err != nil {
		t.Fatalf("first Install() unexpected error: %v", err)
	}

	outcomes, err := installer.Install(target, true)
	if err != nil {
		t.Fatalf("Install(overwrite) unexpected error: %v", err)
	}
	for _, o := range outcomes {
		if o.Status != installer.Installed {
			t.Errorf("Outcome for %q Status = %v, want Installed", o.Resource.Name, o.Status)
		}
	}
}

// Note: a black-box "Install() escapes targetDir" test isn't reachable
// through the public API today — every Resource.Name List() can ever
// produce comes from kit.TemplatesFS's own directory entries, which
// go:embed guarantees are plain basenames with no path separators or
// "..". The same situation 003-entity-creation documented for Create's
// "already exists" check: the containment guard is real, defensive
// infrastructure (FR-006), but nothing in today's resource set can
// trigger it black-box. Its real test coverage is the white-box unit
// test of the underlying destination computation in
// containment_test.go, which exercises it directly with a contrived,
// deliberately malicious resource name.

func TestListFS_SourceIndependent(t *testing.T) {
	resources := installer.ListFS(fixtureSkillsFS(), "skills", "skill")
	if len(resources) != 2 {
		t.Fatalf("ListFS() = %d resources, want 2: %+v", len(resources), resources)
	}
	for _, r := range resources {
		if r.Kind != "skill" {
			t.Errorf("Resource %q Kind = %q, want %q", r.Name, r.Kind, "skill")
		}
	}
}

func TestList_IsListFSOverTemplates(t *testing.T) {
	viaList := installer.List()
	viaListFS := installer.ListFS(kit.TemplatesFS, "templates", "template")
	if len(viaList) != len(viaListFS) {
		t.Fatalf("List() = %d resources, ListFS(kit.TemplatesFS, \"templates\", \"template\") = %d — want identical", len(viaList), len(viaListFS))
	}
	for i := range viaList {
		if viaList[i] != viaListFS[i] {
			t.Errorf("List()[%d] = %+v, ListFS(...)[%d] = %+v — want identical", i, viaList[i], i, viaListFS[i])
		}
	}
}

func TestInstallFS_FreshDirectory(t *testing.T) {
	target := t.TempDir()
	source := fixtureSkillsFS()

	outcomes, err := installer.InstallFS(source, "skills", "skill", target, false)
	if err != nil {
		t.Fatalf("InstallFS() unexpected error: %v", err)
	}
	if len(outcomes) != 2 {
		t.Fatalf("InstallFS() = %d outcomes, want 2: %+v", len(outcomes), outcomes)
	}
	for _, o := range outcomes {
		if o.Status != installer.Installed {
			t.Fatalf("Outcome for %q Status = %v, want Installed (Err: %v)", o.Resource.Name, o.Status, o.Err)
		}
		got, err := os.ReadFile(filepath.Join(target, o.Path))
		if err != nil {
			t.Fatalf("reading installed file %s: %v", o.Path, err)
		}
		want, err := fs.ReadFile(source, "skills/"+o.Resource.Name)
		if err != nil {
			t.Fatalf("reading fixture source %s: %v", o.Resource.Name, err)
		}
		if string(got) != string(want) {
			t.Errorf("installed file %s not byte-identical to fixture source", o.Path)
		}
	}
}

func TestInstallFS_AlreadyPopulatedWithoutOverwriteSkips(t *testing.T) {
	target := t.TempDir()
	source := fixtureSkillsFS()

	if _, err := installer.InstallFS(source, "skills", "skill", target, false); err != nil {
		t.Fatalf("first InstallFS() unexpected error: %v", err)
	}
	outcomes, err := installer.InstallFS(source, "skills", "skill", target, false)
	if err != nil {
		t.Fatalf("second InstallFS() unexpected error: %v", err)
	}
	for _, o := range outcomes {
		if o.Status != installer.Skipped {
			t.Errorf("Outcome for %q Status = %v, want Skipped", o.Resource.Name, o.Status)
		}
	}
}

func TestInstall_IsInstallFSOverTemplates(t *testing.T) {
	targetA := t.TempDir()
	targetB := t.TempDir()

	viaInstall, err := installer.Install(targetA, false)
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}
	viaInstallFS, err := installer.InstallFS(kit.TemplatesFS, "templates", "template", targetB, false)
	if err != nil {
		t.Fatalf("InstallFS() unexpected error: %v", err)
	}
	if len(viaInstall) != len(viaInstallFS) {
		t.Fatalf("Install() = %d outcomes, InstallFS(kit.TemplatesFS, ...) = %d — want identical", len(viaInstall), len(viaInstallFS))
	}
	for i := range viaInstall {
		if viaInstall[i].Resource != viaInstallFS[i].Resource || viaInstall[i].Status != viaInstallFS[i].Status || viaInstall[i].Path != viaInstallFS[i].Path {
			t.Errorf("Install()[%d] = %+v, InstallFS(...)[%d] = %+v — want identical shape", i, viaInstall[i], i, viaInstallFS[i])
		}
	}
}

// --- 009-canonical-skills-content: ListFS/InstallFS become recursive ---

func TestListFS_DiscoversNestedResources(t *testing.T) {
	resources := installer.ListFS(fixtureNestedSkillsFS(), ".", "skill")
	if len(resources) != 2 {
		t.Fatalf("ListFS() = %d resources, want 2: %+v", len(resources), resources)
	}

	names := make([]string, len(resources))
	for i, r := range resources {
		names[i] = r.Name
	}
	sort.Strings(names)
	want := []string{"create-plan/SKILL.md", "create-tasks/SKILL.md"}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("names = %v, want %v — Resource.Name must be the path relative to sourceDir, not a bare filename, for a nested resource", names, want)
			break
		}
	}
}

func TestInstallFS_NestedResourceInstallsAtNestedPath(t *testing.T) {
	target := t.TempDir()

	outcomes, err := installer.InstallFS(fixtureNestedSkillsFS(), ".", "skill", target, false)
	if err != nil {
		t.Fatalf("InstallFS() unexpected error: %v", err)
	}
	if len(outcomes) != 2 {
		t.Fatalf("InstallFS() = %d outcomes, want 2: %+v", len(outcomes), outcomes)
	}

	for _, o := range outcomes {
		if o.Status != installer.Installed {
			t.Fatalf("Outcome for %q Status = %v, want Installed (Err: %v)", o.Resource.Name, o.Status, o.Err)
		}
		got, err := os.ReadFile(filepath.Join(target, o.Path))
		if err != nil {
			t.Fatalf("reading installed file at nested path %s: %v", o.Path, err)
		}
		want, err := fs.ReadFile(fixtureNestedSkillsFS(), o.Resource.Name)
		if err != nil {
			t.Fatalf("reading fixture source %s: %v", o.Resource.Name, err)
		}
		if string(got) != string(want) {
			t.Errorf("installed file %s not byte-identical to fixture source", o.Path)
		}
	}

	// Parent directories (e.g. "create-plan/") must be created
	// automatically — WriteAtomicFile's existing MkdirAll, unchanged.
	if _, err := os.Stat(filepath.Join(target, "create-plan")); err != nil {
		t.Errorf("nested parent directory not created: %v", err)
	}
}

func TestListFS_FlatFixturesUnaffectedByRecursion(t *testing.T) {
	// The exact resource set for every existing flat fixture must be
	// identical before and after the recursive-walk change — asserted
	// directly here, not only inferred from other tests still passing
	// (research.md's backward-compatibility claim).
	resources := installer.ListFS(fixtureSkillsFS(), "skills", "skill")
	if len(resources) != 2 {
		t.Fatalf("ListFS() = %d resources, want 2: %+v", len(resources), resources)
	}
	names := make([]string, len(resources))
	for i, r := range resources {
		names[i] = r.Name
	}
	sort.Strings(names)
	want := []string{"one.md", "two.md"}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("names = %v, want %v — a flat fixture's Resource.Name must remain a bare filename", names, want)
			break
		}
	}

	templateResources := installer.List()
	if len(templateResources) != 8 {
		t.Fatalf("List() = %d resources, want 8 (kit.TemplatesFS is flat — recursion must not change its count)", len(templateResources))
	}
	for _, r := range templateResources {
		if strings.Contains(r.Name, "/") {
			t.Errorf("template Resource.Name = %q contains a path separator, want a bare filename (kit.TemplatesFS is flat)", r.Name)
		}
	}
}
