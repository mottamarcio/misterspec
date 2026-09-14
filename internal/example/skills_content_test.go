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
	"inventory": true, "validate": true, "status": true,
}

// skillOperationsAllowlist is data-model.md's per-Skill operations
// table — the specific subset of knownInternalCommands each Skill is
// expected to use.
var skillOperationsAllowlist = map[string][]string{
	"create-knowledge-base": {"inventory", "fingerprint", "create", "inspect", "validate"},
	"create-constitution":   {"inventory", "resolve", "validate"},
	"create-program":        {"status", "create", "validate"},
	"create-feature":        {"resolve", "children", "create", "validate"},
	"create-specs":          {"resolve", "children", "create", "validate"},
	"create-plan":           {"resolve", "inspect", "create-artifact", "validate"},
	"create-tasks":          {"resolve", "inspect", "create-artifact", "validate"},
	"implement":             {"resolve", "inspect", "validate"},
	"analyze":               {"resolve", "inspect", "create-artifact", "validate"},
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
	assertSkillConformant(t, "create-knowledge-base")
	assertSkillConformant(t, "create-constitution")
}

func TestSkillsContent_ProgramFeatureSpecs(t *testing.T) {
	assertSkillConformant(t, "create-program")
	assertSkillConformant(t, "create-feature")
	assertSkillConformant(t, "create-specs")
}

func TestSkillsContent_PlanTasksImplementAnalyze(t *testing.T) {
	assertSkillConformant(t, "create-plan")
	assertSkillConformant(t, "create-tasks")
	assertSkillConformant(t, "implement")
	assertSkillConformant(t, "analyze")
}

// canonicalSkillNames is docs/architecture-specification.md §38's
// exact MVP Skill set.
var canonicalSkillNames = []string{
	"create-knowledge-base", "create-constitution", "create-program",
	"create-feature", "create-specs", "create-plan", "create-tasks",
	"implement", "analyze",
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
