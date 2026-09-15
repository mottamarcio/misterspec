package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
)

func TestView_AgentSelection_ListsAgents(t *testing.T) {
	m := Model{
		screen:    ScreenAgentSelection,
		agentList: []agents.Adapter{fakeAdapter{id: "claude-code"}},
	}

	out := m.View()
	if out == "" {
		t.Fatal("View() is empty, want the agent list rendered")
	}
	if !strings.Contains(out, "claude-code") {
		t.Errorf("View() = %q, want it to mention %q", out, "claude-code")
	}
}

func TestView_Preview_ShowsPreviewContent(t *testing.T) {
	m := Model{
		screen: ScreenPreview,
		preview: previewContent{
			configPath:      ".misterspec/config.yaml",
			templates:       []resourceSummary{{name: "program.md.tmpl", kind: "template"}},
			skillResources:  []resourceSummary{{name: "skill-one/SKILL.md", kind: "skill"}},
			agentTargetPath: ".claude/skills",
		},
	}

	out := m.View()
	for _, want := range []string{".misterspec/config.yaml", "program.md.tmpl", "skill-one/SKILL.md", ".claude/skills"} {
		if !strings.Contains(out, want) {
			t.Errorf("View() missing %q\ngot: %s", want, out)
		}
	}
}

func TestView_Installing_ShowsProgressIndicator(t *testing.T) {
	m := Model{screen: ScreenInstalling}

	out := m.View()
	if out == "" {
		t.Error("View() is empty, want a progress indicator")
	}
}

func TestView_NonEmptyWarning_AlreadyInitialized(t *testing.T) {
	m := Model{
		screen: ScreenNonEmptyWarning,
		inspect: bootstrap.InspectResult{
			Initialized:    true,
			AgentInstalled: true,
			InstalledAgent: "claude-code",
		},
	}

	out := m.View()
	if out == "" {
		t.Fatal("View() is empty, want the already-initialized warning")
	}
	if !strings.Contains(out, "claude-code") {
		t.Errorf("View() = %q, want it to name the installed agent", out)
	}
}

func TestView_NonEmptyWarning_NonEmptyNotAProject(t *testing.T) {
	m := Model{
		screen: ScreenNonEmptyWarning,
		inspect: bootstrap.InspectResult{
			Initialized: false,
			Empty:       false,
		},
	}

	out := m.View()
	if out == "" {
		t.Fatal("View() is empty, want the non-empty-directory warning")
	}
	if strings.Contains(out, "claude-code") {
		t.Error("View() mentions an installed agent for a target that isn't even a misterspec project")
	}
}

func TestView_NonEmptyWarning_TwoCasesAreDistinct(t *testing.T) {
	already := Model{screen: ScreenNonEmptyWarning, inspect: bootstrap.InspectResult{Initialized: true}}
	nonEmpty := Model{screen: ScreenNonEmptyWarning, inspect: bootstrap.InspectResult{Initialized: false, Empty: false}}

	if already.View() == nonEmpty.View() {
		t.Error("already-initialized and non-empty warnings render identically, want two distinct messages")
	}
}

func TestView_Error_ShowsSpecificFailureAndRetrySafety(t *testing.T) {
	m := Model{
		screen:     ScreenError,
		installErr: errors.New("bootstrap: already initialized"),
	}

	out := m.View()
	if out == "" {
		t.Fatal("View() is empty, want the specific failure rendered")
	}
	if !strings.Contains(out, "already initialized") {
		t.Errorf("View() = %q, want it to name the specific failure, not a generic message", out)
	}
	if !strings.Contains(strings.ToLower(out), "safe") {
		t.Errorf("View() = %q, want it to state that re-running misterspec init is safe (FR-008)", out)
	}
}

func TestView_Error_DistinctFromGenericMessage(t *testing.T) {
	specific := Model{screen: ScreenError, installErr: errors.New("disk full")}
	other := Model{screen: ScreenError, installErr: errors.New("permission denied")}

	if specific.View() == other.View() {
		t.Error("two different failures render identically, want the specific failure text reflected each time")
	}
}

func TestView_Success_ShowsOutcomeSummary(t *testing.T) {
	m := Model{
		screen: ScreenSuccess,
		outcome: bootstrap.BootstrapOutcome{
			ProjectRoot:   "/tmp/fixture-project",
			ConfigWritten: true,
			AgentInstall:  agents.InstallResult{AdapterID: "fake-agent"},
		},
	}

	out := m.View()
	for _, want := range []string{"/tmp/fixture-project", "fake-agent"} {
		if !strings.Contains(out, want) {
			t.Errorf("View() missing %q\ngot: %s", want, out)
		}
	}
}
