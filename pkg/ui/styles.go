package ui

import (
	"fmt"

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

	AssistantBadge = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)
)

// RenderBanner outputs a stylized Lipgloss header card.
func RenderBanner(modelName, baseURL string) string {
	title := TitleStyle.Render("⚡ FLASHWHIP ADK 2.0 CLI")
	modelInfo := fmt.Sprintf("%s %s", InfoLabel.Render("Model:"), InfoValue.Render(modelName))
	endpointInfo := fmt.Sprintf("%s %s", InfoLabel.Render("Endpoint:"), InfoValue.Render(baseURL))
	hint := lipgloss.NewStyle().Foreground(ThinkingColor).Render("Type 'exit', 'quit', or press Ctrl+C to exit.")

	content := fmt.Sprintf("%s\n\n%s\n%s\n\n%s", title, modelInfo, endpointInfo, hint)
	return BannerStyle.Render(content)
}
