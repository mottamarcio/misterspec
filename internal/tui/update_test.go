package tui

import (
	"testing"
	"testing/fstest"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
	"github.com/mottamarcio/misterspec/internal/project"
)

func fixtureSkills() fstest.MapFS {
	return fstest.MapFS{
		"skill-one/SKILL.md": {Data: []byte("# skill-one\n")},
	}
}

// --- User Story 1: ScreenAgentSelection ---

func TestUpdate_AgentSelection_CursorMovement(t *testing.T) {
	m := Model{
		screen:    ScreenAgentSelection,
		agentList: []agents.Adapter{fakeAdapter{id: "a"}, fakeAdapter{id: "b"}},
		cursor:    0,
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	mm := updated.(Model)
	if mm.cursor != 1 {
		t.Errorf("cursor after KeyDown = %d, want 1", mm.cursor)
	}

	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyUp})
	mm = updated.(Model)
	if mm.cursor != 0 {
		t.Errorf("cursor after KeyUp = %d, want 0", mm.cursor)
	}
}

func TestUpdate_AgentSelection_ConfirmBuildsPreviewAndTransitions(t *testing.T) {
	m := Model{
		screen:    ScreenAgentSelection,
		targetDir: t.TempDir(),
		skills:    fixtureSkills(),
		agentList: []agents.Adapter{fakeAdapter{id: "fake-agent", target: ".fake/skills"}},
		cursor:    0,
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm := updated.(Model)

	if mm.screen != ScreenPreview {
		t.Fatalf("screen = %v, want ScreenPreview", mm.screen)
	}
	if mm.selectedAgent == nil || mm.selectedAgent.ID() != "fake-agent" {
		t.Fatalf("selectedAgent = %v, want fake-agent", mm.selectedAgent)
	}
	if mm.preview.agentTargetPath != ".fake/skills" {
		t.Errorf("preview.agentTargetPath = %q, want %q", mm.preview.agentTargetPath, ".fake/skills")
	}
	if len(mm.preview.skillResources) != 1 {
		t.Errorf("preview.skillResources = %d entries, want 1", len(mm.preview.skillResources))
	}
	if mm.preview.configPath == "" {
		t.Error("preview.configPath is empty")
	}
	if len(mm.preview.directories) != 6 {
		t.Errorf("preview.directories = %d entries, want 6", len(mm.preview.directories))
	}
}

func TestUpdate_AgentSelection_CancelQuits(t *testing.T) {
	m := Model{
		screen:    ScreenAgentSelection,
		agentList: []agents.Adapter{fakeAdapter{id: "a"}},
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm := updated.(Model)

	if !mm.quitting {
		t.Error("quitting = false, want true")
	}
	if cmd == nil {
		t.Error("Update() returned a nil Cmd, want tea.Quit")
	}
}

// --- User Story 1: ScreenPreview ---

func TestUpdate_Preview_ConfirmTransitionsToInstalling(t *testing.T) {
	agent := fakeAdapter{id: "fake-agent", target: ".fake/skills"}
	m := Model{
		screen:        ScreenPreview,
		targetDir:     t.TempDir(),
		registry:      agents.NewRegistry(agent),
		skills:        fixtureSkills(),
		selectedAgent: agent,
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	mm := updated.(Model)

	if mm.screen != ScreenInstalling {
		t.Fatalf("screen = %v, want ScreenInstalling", mm.screen)
	}
	if cmd == nil {
		t.Fatal("Update() returned a nil Cmd, want one that calls bootstrap.Bootstrap")
	}

	msg := cmd()
	if _, ok := msg.(bootstrapResultMsg); !ok {
		t.Fatalf("cmd() produced %T, want bootstrapResultMsg", msg)
	}
}

func TestUpdate_Preview_DeclineQuitsWithoutWriting(t *testing.T) {
	target := t.TempDir()
	agent := fakeAdapter{id: "fake-agent", target: ".fake/skills"}
	m := Model{
		screen:        ScreenPreview,
		targetDir:     target,
		registry:      agents.NewRegistry(agent),
		skills:        fixtureSkills(),
		selectedAgent: agent,
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	mm := updated.(Model)

	if !mm.quitting {
		t.Error("quitting = false, want true")
	}
	if cmd != nil {
		if _, ok := cmd().(bootstrapResultMsg); ok {
			t.Fatal("Update() triggered a bootstrap.Bootstrap call after decline, want nothing written")
		}
	}

	bootstrapped, err := bootstrap.Inspect(target)
	if err != nil {
		t.Fatalf("Inspect() unexpected error: %v", err)
	}
	if bootstrapped.Initialized {
		t.Error("target was bootstrapped despite declining at the preview")
	}
}

// --- User Story 1: ScreenInstalling -> ScreenSuccess ---

func TestUpdate_BootstrapResult_SuccessTransitionsToSuccess(t *testing.T) {
	m := Model{screen: ScreenInstalling}
	outcome := bootstrap.BootstrapOutcome{ProjectRoot: "/tmp/fixture-project", ConfigWritten: true}

	updated, cmd := m.Update(bootstrapResultMsg{outcome: outcome})
	mm := updated.(Model)

	if mm.screen != ScreenSuccess {
		t.Fatalf("screen = %v, want ScreenSuccess", mm.screen)
	}
	if mm.outcome.ProjectRoot != "/tmp/fixture-project" {
		t.Errorf("outcome.ProjectRoot = %q, want %q", mm.outcome.ProjectRoot, "/tmp/fixture-project")
	}
	if !mm.quitting {
		t.Error("quitting = false, want true — Success auto-quits after rendering")
	}
	if cmd == nil {
		t.Error("Update() returned a nil Cmd, want tea.Quit")
	}
}

// --- User Story 2: ScreenInspect -> ScreenNonEmptyWarning ---

func TestUpdate_InspectResult_AlreadyInitializedGoesToWarning(t *testing.T) {
	m := Model{registry: agents.NewRegistry()}

	updated, _ := m.Update(inspectResultMsg{result: bootstrap.InspectResult{
		Initialized:    true,
		AgentInstalled: true,
		InstalledAgent: "claude-code",
	}})
	mm := updated.(Model)

	if mm.screen != ScreenNonEmptyWarning {
		t.Fatalf("screen = %v, want ScreenNonEmptyWarning", mm.screen)
	}
	if !mm.inspect.Initialized || mm.inspect.InstalledAgent != "claude-code" {
		t.Errorf("inspect = %+v, unexpected", mm.inspect)
	}
}

func TestUpdate_InspectResult_NonEmptyGoesToWarning(t *testing.T) {
	m := Model{registry: agents.NewRegistry()}

	updated, _ := m.Update(inspectResultMsg{result: bootstrap.InspectResult{
		Initialized: false,
		Empty:       false,
	}})
	mm := updated.(Model)

	if mm.screen != ScreenNonEmptyWarning {
		t.Fatalf("screen = %v, want ScreenNonEmptyWarning", mm.screen)
	}
}

func TestUpdate_InspectResult_EmptyUninitializedGoesToAgentSelection(t *testing.T) {
	m := Model{registry: agents.NewRegistry(fakeAdapter{id: "a"})}

	updated, _ := m.Update(inspectResultMsg{result: bootstrap.InspectResult{
		Initialized: false,
		Empty:       true,
	}})
	mm := updated.(Model)

	if mm.screen != ScreenAgentSelection {
		t.Fatalf("screen = %v, want ScreenAgentSelection", mm.screen)
	}
}

// --- User Story 2: ScreenNonEmptyWarning's own choices ---

func TestUpdate_NonEmptyWarning_CancelQuits(t *testing.T) {
	m := Model{screen: ScreenNonEmptyWarning}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm := updated.(Model)

	if !mm.quitting {
		t.Error("quitting = false, want true")
	}
	if cmd == nil {
		t.Error("Update() returned a nil Cmd, want tea.Quit")
	}
}

// --- User Story 3: ScreenError entry paths ---

func TestUpdate_BootstrapResult_FailureTransitionsToError(t *testing.T) {
	m := Model{screen: ScreenInstalling}

	updated, cmd := m.Update(bootstrapResultMsg{err: bootstrap.ErrAlreadyInitialized})
	mm := updated.(Model)

	if mm.screen != ScreenError {
		t.Fatalf("screen = %v, want ScreenError", mm.screen)
	}
	if mm.installErr == nil {
		t.Error("installErr is nil, want the bootstrap failure recorded")
	}
	if !mm.quitting {
		t.Error("quitting = false, want true — Error auto-quits after rendering")
	}
	if cmd == nil {
		t.Error("Update() returned a nil Cmd, want tea.Quit")
	}
}

func TestUpdate_InspectResult_GenuineErrorTransitionsToError(t *testing.T) {
	m := Model{registry: agents.NewRegistry()}

	updated, cmd := m.Update(inspectResultMsg{err: project.ErrInvalidConfiguration})
	mm := updated.(Model)

	if mm.screen != ScreenError {
		t.Fatalf("screen = %v, want ScreenError", mm.screen)
	}
	if mm.inspectErr == nil {
		t.Error("inspectErr is nil, want the Inspect failure recorded")
	}
	if !mm.quitting {
		t.Error("quitting = false, want true — Error auto-quits after rendering")
	}
	if cmd == nil {
		t.Error("Update() returned a nil Cmd, want tea.Quit")
	}
}

func TestUpdate_NonEmptyWarning_ContinueGoesToAgentSelection(t *testing.T) {
	m := Model{
		screen:   ScreenNonEmptyWarning,
		registry: agents.NewRegistry(fakeAdapter{id: "a"}),
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	mm := updated.(Model)

	if mm.screen != ScreenAgentSelection {
		t.Fatalf("screen = %v, want ScreenAgentSelection", mm.screen)
	}
	if len(mm.agentList) != 1 {
		t.Errorf("agentList = %d entries, want 1", len(mm.agentList))
	}
}
