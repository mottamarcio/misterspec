package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
	warningStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	errorStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	successStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
)
