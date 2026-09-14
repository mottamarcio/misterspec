package tui

import (
	"io/fs"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
)

// Screen is one of the seven states this flow can be in
// (docs/architecture-specification.md §37, verbatim).
type Screen int

const (
	ScreenInspect Screen = iota
	ScreenNonEmptyWarning
	ScreenAgentSelection
	ScreenPreview
	ScreenInstalling
	ScreenSuccess
	ScreenError
)

// previewContent is the Preview screen's own content, assembled once
// from already-exported, read-only calls (research.md) — never a new
// deterministic operation.
type previewContent struct {
	configPath      string
	templates       []resourceSummary
	skillResources  []resourceSummary
	agentTargetPath string
}

// resourceSummary is a trimmed-down installer.Resource for display
// purposes only.
type resourceSummary struct {
	name string
	kind string
}

// Model is the Bubble Tea model for the interactive init flow — UI
// state only, no filesystem business rules of its own (§37's own
// explicit rule; data-model.md).
type Model struct {
	screen    Screen
	targetDir string
	registry  *agents.Registry
	skills    fs.FS

	inspect    bootstrap.InspectResult
	inspectErr error

	agentList     []agents.Adapter
	cursor        int
	selectedAgent agents.Adapter

	preview previewContent

	outcome    bootstrap.BootstrapOutcome
	installErr error

	quitting bool
}

// NewModel constructs the initial Model for targetDir, ready to run via
// RunInit.
func NewModel(targetDir string, registry *agents.Registry, skills fs.FS) Model {
	return Model{
		screen:    ScreenInspect,
		targetDir: targetDir,
		registry:  registry,
		skills:    skills,
	}
}

// inspectResultMsg carries bootstrap.Inspect's result back into Update.
type inspectResultMsg struct {
	result bootstrap.InspectResult
	err    error
}

// Init kicks off ScreenInspect's own work: running bootstrap.Inspect
// against targetDir, off the main render loop.
func (m Model) Init() tea.Cmd {
	targetDir := m.targetDir
	return func() tea.Msg {
		result, err := bootstrap.Inspect(targetDir)
		return inspectResultMsg{result: result, err: err}
	}
}

// RunInit runs the interactive init flow to completion: inspect the
// target, warn if it's already initialized or non-empty, let the user
// pick a registered agent, preview exactly what will be created, and —
// only on explicit confirmation — bootstrap it via bootstrap.Bootstrap
// (FR-006, FR-011). opts lets a caller (tests) redirect the program's
// I/O away from a real terminal.
//
// RunInit returns a non-nil error only for a genuine failure to run the
// interactive program itself — never to represent a user's own
// decision or a reported bootstrap failure, both of which are shown
// on-screen and end the program normally (research.md).
func RunInit(targetDir string, registry *agents.Registry, skills fs.FS, opts ...tea.ProgramOption) error {
	model := NewModel(targetDir, registry, skills)
	program := tea.NewProgram(model, opts...)
	_, err := program.Run()
	return err
}
