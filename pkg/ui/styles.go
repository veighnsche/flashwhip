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
)
