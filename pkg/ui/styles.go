package ui

import (
	"fmt"
	"os"

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

// RenderBanner outputs a stylized Lipgloss header card.
func RenderBanner(modelName, baseURL string) string {
	title := TitleStyle.Render("⚡ FLASHWHIP ADK 2.0 CLI")
	modelInfo := fmt.Sprintf("%s %s", InfoLabel.Render("Model:"), InfoValue.Render(modelName))
	endpointInfo := fmt.Sprintf("%s %s", InfoLabel.Render("Endpoint:"), InfoValue.Render(baseURL))
	hint := lipgloss.NewStyle().Foreground(ThinkingColor).Render("Type 'exit', 'quit', or press Ctrl+C to exit.")

	content := fmt.Sprintf("%s\n\n%s\n%s\n\n%s", title, modelInfo, endpointInfo, hint)
	return BannerStyle.Render(content)
}

type StreamState int

const (
	StateIdle StreamState = iota
	StateThinking
	StateOutputting
)

type StreamTracker struct {
	State        StreamState
	ThinkingStep int
}

func NewStreamTracker() *StreamTracker {
	return &StreamTracker{
		State:        StateIdle,
		ThinkingStep: 0,
	}
}

func (st *StreamTracker) TransitionToThinking() {
	if st.State != StateThinking {
		st.ThinkingStep++
		if st.State == StateOutputting {
			fmt.Print("\n\n")
		}
		if st.ThinkingStep > 1 {
			fmt.Printf("\n%s\n", ThinkingBadge.Render(fmt.Sprintf("🧠 [Thinking (Step %d)]:", st.ThinkingStep)))
		} else {
			fmt.Printf("\n%s\n", ThinkingBadge.Render("🧠 [Thinking]:"))
		}
		_ = os.Stdout.Sync()
		st.State = StateThinking
	}
}

func (st *StreamTracker) TransitionToOutputting() {
	if st.State != StateOutputting {
		if st.State == StateThinking {
			fmt.Print("\n\n")
		}
		_ = os.Stdout.Sync()
		st.State = StateOutputting
	}
}

func (st *StreamTracker) TransitionToIdle() {
	if st.State != StateIdle {
		fmt.Print("\n")
		_ = os.Stdout.Sync()
		st.State = StateIdle
	}
}
