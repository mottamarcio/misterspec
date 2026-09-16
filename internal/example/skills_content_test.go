// This file machine-checks every canonical Skill's content against
// specs/009-canonical-skills-content/data-model.md's contract:
// frontmatter present, every §39 heading present and in order, every
// "internal <op>" mention is a real, registered command matching that
// Skill's own operations allowlist, and the Completion Contract section
// names §50's required concepts (research.md's "structural conformance
// is machine-checked, not only reviewed by eye" decision).
package example

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/agents/claude"
	"github.com/mottamarcio/misterspec/kit"
)

// requiredSkillHeadings is docs/architecture-specification.md §39's
// required SKILL.md body structure, in this exact order.
var requiredSkillHeadings = []string{
	"Purpose", "Invocation", "Responsibility", "Inputs", "Outputs", "Preconditions",
	"Required Context", "Optional Context", "Conditional Context", "Unnecessary Context",
	"Authority", "Allowed Reads", "Allowed Creates", "Allowed Modifications", "Forbidden Mutations",
	"Deterministic Operations", "Procedure", "Decision Rules", "Interaction Rules",
	"Validation Rules", "Failure Conditions", "Stop Conditions", "Success Criteria",
	"Postconditions", "Idempotency", "Resume Behavior",
	"Completion Contract", "Recommended Next Step", "Related Skills",
}

// knownInternalCommands is every command 008-cli-cobra actually
// registered under "misterspec internal" — the allowlist FR-003/SC-004
// check every Skill's Deterministic Operations section against. No
// "project", no "references", no Task-ID allocator (research.md).
var knownInternalCommands = map[string]bool{
	"resolve": true, "inspect": true, "parent": true, "children": true,
	"create": true, "create-artifact": true, "fingerprint": true,
	"inventory": true, "validate": true, "status": true, "context": true,
}

// skillOperationsAllowlist is data-model.md's per-Skill operations
// table — the specific subset of knownInternalCommands each Skill is
// expected to use.
var skillOperationsAllowlist = map[string][]string{
	"mister-knowledge-base": {"inventory", "fingerprint", "create", "inspect", "validate"},
	"mister-constitution":   {"inventory", "resolve", "validate"},
	"mister-program":        {"status", "create", "validate"},
	"mister-features":       {"resolve", "children", "create", "validate"},
	"mister-specify":        {"resolve", "children", "create", "validate"},
	"mister-plan":           {"resolve", "inspect", "context", "create-artifact", "validate"},
	"mister-tasks":          {"resolve", "inspect", "context", "create-artifact", "validate"},
	"mister-implement":      {"resolve", "inspect", "context", "validate"},
	"mister-analyze":        {"resolve", "inspect", "context", "create-artifact", "validate"},
}

// internalOpRe matches an inline-code-formatted operation mention, e.g.
// "`internal resolve SPEC-###`" — the convention every Skill's own
// Deterministic Operations section is authored to follow.
var internalOpRe = regexp.MustCompile("`internal ([a-z][a-z-]*)")

// assertSkillConformant reads kit/skills/<name>/SKILL.md (via
// kit.SkillsFS) and asserts it satisfies
// specs/009-canonical-skills-content/data-model.md's content contract
// in full. Shared by every story's own TestSkillsContent_* function.
func assertSkillConformant(t *testing.T, name string) {
	t.Helper()

	data, err := fs.ReadFile(kit.SkillsFS, name+"/SKILL.md")
	if err != nil {
		t.Fatalf("reading kit/skills/%s/SKILL.md: %v", name, err)
	}
	content := string(data)

	assertFrontmatter(t, name, content)
	assertSectionsPresentInOrder(t, name, content)
	assertOperationsAllowlisted(t, name, content)
	assertCompletionContractConcepts(t, name, content)
}

func assertFrontmatter(t *testing.T, name, content string) {
	t.Helper()

	if !strings.HasPrefix(content, "---\n") {
		t.Fatalf("%s: SKILL.md does not start with a YAML frontmatter block", name)
	}
	closeIdx := strings.Index(content[4:], "\n---")
	if closeIdx == -1 {
		t.Fatalf("%s: SKILL.md frontmatter has no closing delimiter", name)
	}
	frontmatter := content[4 : 4+closeIdx]

	if !strings.Contains(frontmatter, "name:") {
		t.Errorf("%s: frontmatter missing \"name:\"", name)
	}
	if !strings.Contains(frontmatter, "description:") {
		t.Errorf("%s: frontmatter missing \"description:\"", name)
	}
}

func assertSectionsPresentInOrder(t *testing.T, name, content string) {
	t.Helper()

	lastIdx := -1
	for _, heading := range requiredSkillHeadings {
		marker := "## " + heading
		idx := strings.Index(content, marker)
		if idx == -1 {
			t.Errorf("%s: missing required section %q", name, marker)
			continue
		}
		if idx < lastIdx {
			t.Errorf("%s: section %q appears out of §39's required order", name, marker)
		}
		lastIdx = idx
	}
}

func assertOperationsAllowlisted(t *testing.T, name, content string) {
	t.Helper()

	allowed := skillOperationsAllowlist[name]
	section := extractSection(content, "Deterministic Operations")
	if section == "" {
		t.Errorf("%s: Deterministic Operations section is empty or missing", name)
		return
	}

	for _, m := range internalOpRe.FindAllStringSubmatch(section, -1) {
		op := m[1]
		if !knownInternalCommands[op] {
			t.Errorf("%s: Deterministic Operations references \"internal %s\", which is not a real registered command (FR-003)", name, op)
			continue
		}
		if !stringInSlice(allowed, op) {
			t.Errorf("%s: Deterministic Operations references \"internal %s\", not in this Skill's own allowlist %v (data-model.md)", name, op, allowed)
		}
	}
}

func assertCompletionContractConcepts(t *testing.T, name, content string) {
	t.Helper()

	section := strings.ToLower(extractSection(content, "Completion Contract"))
	if section == "" {
		t.Errorf("%s: Completion Contract section is empty or missing", name)
		return
	}

	for _, concept := range []string{"outcome", "artifact", "finding", "next step"} {
		if !strings.Contains(section, concept) {
			t.Errorf("%s: Completion Contract section does not mention %q (§50)", name, concept)
		}
	}
	if !strings.Contains(section, "attention") && !strings.Contains(section, "unresolved") {
		t.Errorf("%s: Completion Contract section does not mention attention/unresolved issues (§50)", name)
	}
}

// extractSection returns heading's own body text — everything after
// "## <heading>" up to (not including) the next "## " marker, or to the
// end of content if heading is the last section.
func extractSection(content, heading string) string {
	marker := "## " + heading
	start := strings.Index(content, marker)
	if start == -1 {
		return ""
	}
	rest := content[start+len(marker):]
	next := strings.Index(rest, "\n## ")
	if next == -1 {
		return rest
	}
	return rest[:next]
}

func stringInSlice(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestSkillsContent_KnowledgeAndConstitution(t *testing.T) {
	assertSkillConformant(t, "mister-knowledge-base")
	assertSkillConformant(t, "mister-constitution")
}

// TestSkillsContent_ConstitutionFrontmatterDocumented is
// specs/027-constitution-frontmatter-task-deps's own contract check: the
// create-constitution Skill's Outputs section must explicitly document
// the Constitution's own required frontmatter fields
// (docs/architecture-specification.md §24), so a written Constitution's
// frontmatter is not left to whichever model happens to include it.
func TestSkillsContent_ConstitutionFrontmatterDocumented(t *testing.T) {
	data, err := fs.ReadFile(kit.SkillsFS, "mister-constitution/SKILL.md")
	if err != nil {
		t.Fatalf("reading kit/skills/mister-constitution/SKILL.md: %v", err)
	}
	section := extractSection(string(data), "Outputs")
	if section == "" {
		t.Fatal("mister-constitution: Outputs section is empty or missing")
	}
	if !strings.Contains(section, "type: constitution") {
		t.Error("mister-constitution: Outputs section does not document the required \"type: constitution\" frontmatter field")
	}
	if !strings.Contains(section, "schema_version") {
		t.Error("mister-constitution: Outputs section does not document the required \"schema_version\" frontmatter field")
	}
}

func TestSkillsContent_ProgramFeatureSpecs(t *testing.T) {
	assertSkillConformant(t, "mister-program")
	assertSkillConformant(t, "mister-features")
	assertSkillConformant(t, "mister-specify")
}

func TestSkillsContent_PlanTasksImplementAnalyze(t *testing.T) {
	assertSkillConformant(t, "mister-plan")
	assertSkillConformant(t, "mister-tasks")
	assertSkillConformant(t, "mister-implement")
	assertSkillConformant(t, "mister-analyze")
}

// TestSkillsContent_CreateTasksReportsDependenciesAndParallelism is
// specs/027-constitution-frontmatter-task-deps's own contract check: the
// create-tasks Skill's Completion Contract section must explicitly
// mention Task dependency relationships and parallel-safe groups, not
// leave that information implicit in tasks.md's own content.
func TestSkillsContent_CreateTasksReportsDependenciesAndParallelism(t *testing.T) {
	data, err := fs.ReadFile(kit.SkillsFS, "mister-tasks/SKILL.md")
	if err != nil {
		t.Fatalf("reading kit/skills/mister-tasks/SKILL.md: %v", err)
	}
	section := strings.ToLower(extractSection(string(data), "Completion Contract"))
	if section == "" {
		t.Fatal("mister-tasks: Completion Contract section is empty or missing")
	}
	if !strings.Contains(section, "depend") {
		t.Error("mister-tasks: Completion Contract section does not mention Task dependency reporting")
	}
	if !strings.Contains(section, "parallel") {
		t.Error("mister-tasks: Completion Contract section does not mention parallel-safe Task reporting")
	}
}

// TestSkillsContent_ImplementDualInvocation is specs/026-implement-single-task's
// own contract check (contracts/implement-invocation.md): the implement
// Skill's Invocation section must document both supported forms — the
// Spec-only form (all executable Tasks, sequentially) and the
// Spec-plus-Task form (exactly one named Task).
func TestSkillsContent_ImplementDualInvocation(t *testing.T) {
	data, err := fs.ReadFile(kit.SkillsFS, "mister-implement/SKILL.md")
	if err != nil {
		t.Fatalf("reading kit/skills/mister-implement/SKILL.md: %v", err)
	}
	section := extractSection(string(data), "Invocation")
	if section == "" {
		t.Fatal("mister-implement: Invocation section is empty or missing")
	}
	if !strings.Contains(section, "SPEC-###") {
		t.Error("mister-implement: Invocation section does not document the Spec-only form (\"SPEC-###\")")
	}
	if !strings.Contains(section, "SPEC-### TASK-NNN") {
		t.Error("mister-implement: Invocation section does not document the Spec-plus-Task form (\"SPEC-### TASK-NNN\")")
	}
}

// staleSkillNameRe matches any of the 9 pre-028-mister-prefixed-skill-
// names slash-command names as a whole word, with its own leading
// slash — e.g. "/implement" but never the "-implement" substring
// inside "/mister-implement", so a renamed Skill's own new name never
// false-positives this check.
var staleSkillNameRe = regexp.MustCompile(`/(analyze|create-constitution|create-feature|create-knowledge-base|create-plan|create-program|create-specs|create-tasks|implement)\b`)

// TestSkillsContent_NoStaleSkillNameReferences is
// specs/028-mister-prefixed-skill-names's own whole-feature regression
// guard: no Skill's own content — self-reference or cross-reference to
// another Skill — may mention any of the 9 pre-rename slash-command
// names anywhere under kit.SkillsFS.
func TestSkillsContent_NoStaleSkillNameReferences(t *testing.T) {
	err := fs.WalkDir(kit.SkillsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(kit.SkillsFS, path)
		if err != nil {
			return err
		}
		for _, m := range staleSkillNameRe.FindAllString(string(data), -1) {
			t.Errorf("%s: contains stale pre-rename skill reference %q", path, m)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking kit.SkillsFS: %v", err)
	}
}

// canonicalSkillNames is docs/architecture-specification.md §38's
// exact MVP Skill set.
var canonicalSkillNames = []string{
	"mister-knowledge-base", "mister-constitution", "mister-program",
	"mister-features", "mister-specify", "mister-plan", "mister-tasks",
	"mister-implement", "mister-analyze",
}

// TestSkillsContent_AllNineInstalled is this feature's whole-feature
// check (plan.md Phase 6, T024): exactly §38's nine Skills exist under
// kit.SkillsFS — no more, no fewer (catching a stray leftover file the
// same way T008's README.md removal was meant to) — and installing them
// for the real Claude Code adapter lands every one at
// .claude/skills/<name>/SKILL.md, byte-identical to kit.SkillsFS's own
// content (quickstart.md's validation strategy).
func TestSkillsContent_AllNineInstalled(t *testing.T) {
	entries, err := fs.ReadDir(kit.SkillsFS, ".")
	if err != nil {
		t.Fatalf("reading kit.SkillsFS root: %v", err)
	}

	var got []string
	for _, e := range entries {
		if !e.IsDir() {
			t.Errorf("kit.SkillsFS root contains a non-directory entry %q, want only per-Skill directories", e.Name())
			continue
		}
		got = append(got, e.Name())
	}
	sort.Strings(got)
	want := append([]string(nil), canonicalSkillNames...)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("kit.SkillsFS contains %d Skill directories %v, want exactly %d %v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("kit.SkillsFS Skill directories = %v, want exactly %v", got, want)
		}
	}

	// Real end-to-end install via the actual Claude Code adapter.
	target := t.TempDir()
	adapter := claude.New()
	result, err := adapter.Install(context.Background(), agents.InstallRequest{
		ProjectRoot: target,
		Skills:      kit.SkillsFS,
		Overwrite:   false,
	})
	if err != nil {
		t.Fatalf("adapter.Install() unexpected error: %v", err)
	}
	if len(result.Outcomes) != len(canonicalSkillNames) {
		t.Fatalf("Install() produced %d outcomes, want %d (one SKILL.md per Skill)", len(result.Outcomes), len(canonicalSkillNames))
	}

	for _, name := range canonicalSkillNames {
		installedPath := filepath.Join(target, ".claude", "skills", name, "SKILL.md")
		got, err := os.ReadFile(installedPath)
		if err != nil {
			t.Fatalf("Skill %q not installed at its Claude Code discovery path %s: %v", name, installedPath, err)
		}
		want, err := fs.ReadFile(kit.SkillsFS, name+"/SKILL.md")
		if err != nil {
			t.Fatalf("reading kit.SkillsFS source for %s: %v", name, err)
		}
		if string(got) != string(want) {
			t.Errorf("installed %s not byte-identical to kit.SkillsFS's own content", installedPath)
		}
	}
}
