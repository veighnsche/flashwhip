package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
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
