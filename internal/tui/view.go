package tui

import (
	"fmt"
	"strings"
)

// View renders the current screen. Per-screen rendering below is
// filled in incrementally by each user story — see
// specs/010-interactive-init-tui/contracts/tui.md's transition table.
func (m Model) View() string {
	switch m.screen {
	case ScreenInspect:
		return "Inspecting " + m.targetDir + "...\n"
	case ScreenNonEmptyWarning:
		return m.viewNonEmptyWarning()
	case ScreenAgentSelection:
		return m.viewAgentSelection()
	case ScreenPreview:
		return m.viewPreview()
	case ScreenInstalling:
		return m.viewInstalling()
	case ScreenSuccess:
		return m.viewSuccess()
	case ScreenError:
		return m.viewError()
	}
	return ""
}

// viewAgentSelection (User Story 1) lists every registered agent, with
// the current cursor position highlighted.
func (m Model) viewAgentSelection() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select a coding agent:") + "\n\n")
	for i, a := range m.agentList {
		cursor := "  "
		line := fmt.Sprintf("%s (%s)", a.ID(), a.Name())
		if i == m.cursor {
			cursor = "> "
			line = selectedStyle.Render(line)
		}
		b.WriteString(cursor + line + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("[Enter to select, Esc to cancel]") + "\n")
	return b.String()
}

// viewPreview (User Story 1) shows exactly what a confirmed install
// would create — the same content bootstrap.Bootstrap then performs
// (FR-004, FR-005).
func (m Model) viewPreview() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("This will create:") + "\n\n")
	b.WriteString("  " + m.preview.configPath + "\n")
	for _, dir := range m.preview.directories {
		b.WriteString("  " + dir + "/\n")
	}
	agentID := ""
	if m.selectedAgent != nil {
		agentID = m.selectedAgent.ID()
	}
	b.WriteString(fmt.Sprintf("  %s (%d Skills for %s)\n", m.preview.agentTargetPath, len(m.preview.skillResources), agentID))
	for _, r := range m.preview.skillResources {
		b.WriteString("    " + r.name + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("Proceed? [y/N]") + "\n")
	return b.String()
}

// viewInstalling (User Story 1) shows a simple progress indicator while
// bootstrap.Bootstrap runs.
func (m Model) viewInstalling() string {
	return titleStyle.Render("Installing...") + "\n"
}

// viewSuccess (User Story 1) reports exactly what was created, matching
// the Preview screen's own content (FR-007).
func (m Model) viewSuccess() string {
	var b strings.Builder
	b.WriteString(successStyle.Render("✓ Project bootstrapped.") + "\n\n")
	b.WriteString("  " + m.outcome.ProjectRoot + "\n")
	if m.outcome.ConfigWritten {
		b.WriteString("  Configuration written\n")
	}
	b.WriteString(fmt.Sprintf("  %d directories scaffolded\n", len(m.outcome.DirectoriesScaffolded)))
	b.WriteString(fmt.Sprintf("  %d Skills installed for %s\n", len(m.outcome.AgentInstall.Outcomes), m.outcome.AgentInstall.AdapterID))
	return b.String()
}

// viewNonEmptyWarning (User Story 2) renders two distinct messages —
// already initialized (naming the installed agent, if any) versus
// merely non-empty — never a generic one.
func (m Model) viewNonEmptyWarning() string {
	var b strings.Builder
	if m.inspect.Initialized {
		if m.inspect.AgentInstalled {
			b.WriteString(warningStyle.Render(fmt.Sprintf("Already a misterspec project (agent: %s).", m.inspect.InstalledAgent)) + "\n\n")
		} else {
			b.WriteString(warningStyle.Render("Already a misterspec project (no agent installed).") + "\n\n")
		}
	} else {
		b.WriteString(warningStyle.Render("Directory is not empty (not a misterspec project).") + "\n\n")
	}
	b.WriteString(dimStyle.Render("[C]ontinue anyway    [Esc] Cancel") + "\n")
	return b.String()
}

// viewError (User Story 3) names the specific failure — whichever of
// installErr (a Bootstrap failure) or inspectErr (a genuine Inspect
// failure) is set — and states that re-running misterspec init is
// always safe (FR-008; 007-project-bootstrap's own inspect-then-reject
// guarantee means a partial project is still detectable and rerunnable,
// never silently corrupted).
func (m Model) viewError() string {
	var b strings.Builder
	b.WriteString(errorStyle.Render("✗ Something went wrong.") + "\n\n")

	err := m.installErr
	if err == nil {
		err = m.inspectErr
	}
	if err != nil {
		b.WriteString("  " + err.Error() + "\n\n")
	}

	b.WriteString(dimStyle.Render("It is safe to run misterspec init again.") + "\n")
	return b.String()
}
