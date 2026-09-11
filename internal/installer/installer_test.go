package installer_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/mottamarcio/misterspec/internal/installer"
	"github.com/mottamarcio/misterspec/kit"
)

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
