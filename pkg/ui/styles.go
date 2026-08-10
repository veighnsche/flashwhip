package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	PrimaryColor   = lipgloss.Color("#7D56F4")
	SecondaryColor = lipgloss.Color("#04B575")
	ThinkingColor  = lipgloss.Color("#8A8A8A")
	BadgeColor     = lipgloss.Color("#FF5F87")

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(PrimaryColor).
			Padding(0, 1)

	Bold = lipgloss.NewStyle().Bold(true)

	BannerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2).
			MarginBottom(1)

	InfoLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Bold(true)

	InfoValue = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	ThinkingBadge = lipgloss.NewStyle().
			Foreground(ThinkingColor).
			Bold(true).
			Italic(true)

	ToolCallBadge = lipgloss.NewStyle().
			Foreground(BadgeColor).
			Bold(true)

	ToolResultBadge = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	AssistantBadge = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	// Tool UI compact inline styles
	ToolShellColor   = lipgloss.Color("#00D7FF")
	ToolFileColor    = lipgloss.Color("#3B82F6")
	ToolSearchColor  = lipgloss.Color("#10B981")
	ToolWebColor     = lipgloss.Color("#A855F7")
	ToolGitColor     = lipgloss.Color("#F59E0B")
	ToolDefaultColor = lipgloss.Color("#EC4899")
	SuccessColor     = lipgloss.Color("#10B981")
	ErrorColor       = lipgloss.Color("#EF4444")
	MutedColor       = lipgloss.Color("#6B7280")

	ToolNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB")).
			Bold(true)

	ToolTargetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	ToolOutcomeSuccessStyle = lipgloss.NewStyle().
				Foreground(SecondaryColor)

	ToolOutcomeErrorStyle = lipgloss.NewStyle().
				Foreground(ErrorColor).
				Bold(true)

	ToolMutedStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	ToolErrorBadge = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)
)


