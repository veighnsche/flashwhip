package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"flashwhip/pkg/db"
)

// RenderSessionList builds a stylized Lipgloss table card for conversation sessions.
func RenderSessionList(sessions []db.Session) string {
	if len(sessions) == 0 {
		return ThinkingBadge.Render("No stored conversation sessions found in ~/.flashwhip/flashwhip.db")
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(PrimaryColor).
		Padding(0, 1)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2)

	idStyle := lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true)
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	turnStyle := lipgloss.NewStyle().Foreground(BadgeColor).Bold(true)
	timeStyle := lipgloss.NewStyle().Foreground(ThinkingColor).Italic(true)

	var sb strings.Builder
	sb.WriteString(headerStyle.Render("⚡ FLASHWHIP CONVERSATION SESSIONS"))
	sb.WriteString("\n\n")

	sb.WriteString(fmt.Sprintf("%-20s %-45s %-8s %-15s\n",
		InfoLabel.Render("SESSION ID"),
		InfoLabel.Render("TITLE"),
		InfoLabel.Render("TURNS"),
		InfoLabel.Render("UPDATED"),
	))
	sb.WriteString(lipgloss.NewStyle().Foreground(ThinkingColor).Render(strings.Repeat("─", 92)))
	sb.WriteString("\n")

	for _, s := range sessions {
		title := s.Title
		if len(title) > 42 {
			title = title[:42] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-20s %-45s %-8s %-15s\n",
			idStyle.Render(s.ID),
			titleStyle.Render(title),
			turnStyle.Render(fmt.Sprintf("%d", s.TurnCount)),
			timeStyle.Render(db.FormatRelativeTime(s.UpdatedAt)),
		))
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(ThinkingColor).Render("Tip: Resume any session using './flashwhip chat --session <id>' or '/load <id>' in REPL."))

	return cardStyle.Render(sb.String())
}
