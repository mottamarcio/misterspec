package tui

import (
	"io/fs"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
	"github.com/mottamarcio/misterspec/internal/installer"
)

// bootstrapResultMsg carries bootstrap.Bootstrap's result back into
// Update, once ScreenInstalling's command completes.
type bootstrapResultMsg struct {
	outcome bootstrap.BootstrapOutcome
	err     error
}

// Update is the Bubble Tea update loop. Cancelling (Ctrl+C) is honored
// globally, before any per-screen handling (FR-009). Per-screen
// transition logic below is filled in incrementally by each user
// story — see specs/010-interactive-init-tui/contracts/tui.md's
// transition table.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyCtrlC {
		m.quitting = true
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case inspectResultMsg:
		return m.handleInspectResult(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	case bootstrapResultMsg:
		return m.handleBootstrapResult(msg)
	}

	return m, nil
}

// handleInspectResult stores bootstrap.Inspect's result and routes to
// the correct next screen — a genuine Inspect error goes to
// ScreenError (User Story 3); an already-initialized or non-empty
// target goes to ScreenNonEmptyWarning (User Story 2); otherwise
// straight to ScreenAgentSelection (User Story 1).
func (m Model) handleInspectResult(msg inspectResultMsg) (tea.Model, tea.Cmd) {
	m.inspect = msg.result
	m.inspectErr = msg.err

	switch {
	case msg.err != nil:
		m.screen = ScreenError
		m.quitting = true
		return m, tea.Quit
	case msg.result.Initialized, !msg.result.Empty:
		m.screen = ScreenNonEmptyWarning
	default:
		m.screen = ScreenAgentSelection
		m.agentList = m.registry.List()
	}
	return m, nil
}

// handleKey dispatches a key press to the current screen's own
// handling.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenNonEmptyWarning:
		return m.updateNonEmptyWarning(msg)
	case ScreenAgentSelection:
		return m.updateAgentSelection(msg)
	case ScreenPreview:
		return m.updatePreview(msg)
	case ScreenSuccess, ScreenError:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

// updateAgentSelection (User Story 1): cursor movement between
// registered agents; Enter selects the current one, builds the Preview
// (research.md's read-only composition), and transitions to
// ScreenPreview; Esc cancels.
func (m Model) updateAgentSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "down", "j":
		if m.cursor < len(m.agentList)-1 {
			m.cursor++
		}
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		if len(m.agentList) == 0 {
			return m, nil
		}
		m.selectedAgent = m.agentList[m.cursor]
		m.preview = buildPreview(m.skills, m.selectedAgent)
		m.screen = ScreenPreview
	case "esc":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

// buildPreview assembles the Preview screen's content from
// already-exported, read-only calls only — no new deterministic
// operation (research.md).
func buildPreview(skills fs.FS, agent agents.Adapter) previewContent {
	return previewContent{
		configPath:      ".misterspec/config.yaml",
		templates:       toResourceSummaries(installer.List()),
		skillResources:  toResourceSummaries(installer.ListFS(skills, ".", "skill")),
		agentTargetPath: agent.TargetPath(),
	}
}

func toResourceSummaries(resources []installer.Resource) []resourceSummary {
	out := make([]resourceSummary, 0, len(resources))
	for _, r := range resources {
		out = append(out, resourceSummary{name: r.Name, kind: r.Kind})
	}
	return out
}

// updatePreview (User Story 1): confirm triggers bootstrap.Bootstrap as
// a tea.Cmd and transitions to ScreenInstalling (FR-006); decline quits
// with nothing written (FR-005).
func (m Model) updatePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.screen = ScreenInstalling
		targetDir := m.targetDir
		agentID := m.selectedAgent.ID()
		registry := m.registry
		skills := m.skills
		return m, func() tea.Msg {
			outcome, err := bootstrap.Bootstrap(targetDir, agentID, registry, skills)
			return bootstrapResultMsg{outcome: outcome, err: err}
		}
	case "n", "esc":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

// handleBootstrapResult (success path — User Story 1; failure path —
// User Story 3) routes ScreenInstalling's outcome to ScreenSuccess or
// ScreenError, and quits immediately once that screen's own final View
// has rendered — a keypress-driven "any key to continue" would
// otherwise race this async result's own arrival for no benefit
// (research.md-style refinement made during implementation; see
// contracts/tui.md's reconciliation).
func (m Model) handleBootstrapResult(msg bootstrapResultMsg) (tea.Model, tea.Cmd) {
	m.quitting = true
	if msg.err != nil {
		m.installErr = msg.err
		m.screen = ScreenError
		return m, tea.Quit
	}
	m.outcome = msg.outcome
	m.screen = ScreenSuccess
	return m, tea.Quit
}

// updateNonEmptyWarning (User Story 2): Cancel quits with nothing
// written; Continue proceeds to ScreenAgentSelection — the underlying
// bootstrap.Bootstrap rejection for an already-initialized target is
// still enforced downstream regardless (Constitution Principle VIII;
// User Story 3's Error screen surfaces it clearly if reached).
func (m Model) updateNonEmptyWarning(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		m.screen = ScreenAgentSelection
		m.agentList = m.registry.List()
	case "esc":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}
