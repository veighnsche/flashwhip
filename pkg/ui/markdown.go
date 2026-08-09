package ui

import (
	"github.com/charmbracelet/glamour"
)

// RenderMarkdown converts a markdown string to stylized terminal ANSI output using glamour.
func RenderMarkdown(rawText string) string {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return rawText
	}
	out, err := renderer.Render(rawText)
	if err != nil {
		return rawText
	}
	return out
}
