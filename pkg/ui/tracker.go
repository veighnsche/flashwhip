package ui

import (
	"fmt"
	"os"
	"strings"

	"flashwhip/pkg/config"
	"flashwhip/pkg/db"
	"flashwhip/pkg/provider/ollama"
)

type StreamState int

const (
	StateIdle StreamState = iota
	StateThinking
	StateOutputting
)

type StreamTracker struct {
	State        StreamState
	ThinkingStep int
	SessionID    string
	BaseURL      string
	APIKey       string
	ModelName    string
}

func NewStreamTracker() *StreamTracker {
	return &StreamTracker{
		State:        StateIdle,
		ThinkingStep: 0,
	}
}

func NewStreamTrackerWithConfig(sessionID string, cfg *config.Config) *StreamTracker {
	st := NewStreamTracker()
	st.SessionID = sessionID
	if cfg != nil {
		st.BaseURL = cfg.BaseURL
		st.APIKey = cfg.APIKey
		st.ModelName = cfg.ModelName
	}
	return st
}

// RenderUsageBar builds a visual token usage progress bar string.
func RenderUsageBar(sessionID, baseURL, apiKey, modelName string) string {
	ctxLen, _ := ollama.FetchModelContextLength(baseURL, apiKey, modelName)
	if ctxLen <= 0 {
		ctxLen = 32768
	}

	totalChars := 0
	database, err := db.DefaultDB()
	if err == nil && sessionID != "" {
		contents, cErr := database.GetSessionGenAIContents(sessionID)
		if cErr == nil {
			for _, c := range contents {
				for _, p := range c.Parts {
					totalChars += len(p.Text)
				}
			}
		}
	}

	estTokens := totalChars / 4
	pct := float64(estTokens) / float64(ctxLen) * 100
	if pct > 100 {
		pct = 100
	}

	barWidth := 15
	filled := int((pct / 100.0) * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	return fmt.Sprintf("[%s] %.1f%% (%d/%d tokens)", bar, pct, estTokens, ctxLen)
}

func (st *StreamTracker) TransitionToThinking() {
	if st.State != StateThinking {
		st.ThinkingStep++
		if st.State == StateOutputting {
			fmt.Print("\n\n")
		}

		if st.ModelName != "" {
			usageBarStr := RenderUsageBar(st.SessionID, st.BaseURL, st.APIKey, st.ModelName)
			fmt.Printf("\n%s %s\n", ToolCallBadge.Render("📊 Context Usage:"), InfoValue.Render(usageBarStr))
		}

		if st.ThinkingStep > 1 {
			fmt.Printf("%s\n", ThinkingBadge.Render(fmt.Sprintf("🧠 [Thinking (Step %d)]:", st.ThinkingStep)))
		} else {
			fmt.Printf("%s\n", ThinkingBadge.Render("🧠 [Thinking]:"))
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
