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
	Usage        *ollama.Usage
}

func NewStreamTracker() *StreamTracker {
	return &StreamTracker{
		State:        StateIdle,
		ThinkingStep: 0,
	}
}

func NewStreamTrackerWithConfig(sessionID string, cfg *config.Config, usage *ollama.Usage) *StreamTracker {
	st := NewStreamTracker()
	st.SessionID = sessionID
	st.Usage = usage
	if cfg != nil {
		st.BaseURL = cfg.BaseURL
		st.APIKey = cfg.APIKey
		st.ModelName = cfg.ModelName
	}
	return st
}

// RenderUsageBar builds a visual token usage progress bar string and saturation alerts.
func (st *StreamTracker) RenderUsageBar() string {
	ctxLen, _ := ollama.FetchModelContextLength(st.BaseURL, st.APIKey, st.ModelName)
	if ctxLen <= 0 {
		ctxLen = 32768
	}

	tokens := 0

	// 1. Try real token counts from last LLM prompt turn
	if st.Usage != nil {
		_, _, lastTotal := st.Usage.LastContextTokens()
		if lastTotal > 0 {
			tokens = lastTotal
		}
	}

	// 2. Fallback estimation if no LLM response turn has completed yet (e.g. Turn 1)
	if tokens == 0 {
		totalChars := 0
		database, err := db.DefaultDB()
		if err == nil && st.SessionID != "" {
			contents, cErr := database.GetSessionGenAIContents(st.SessionID)
			if cErr == nil {
				for _, c := range contents {
					for _, p := range c.Parts {
						totalChars += len(p.Text)
					}
				}
			}
		}
		// Base estimate includes system prompt + tool schemas (~1800 tokens) + history
		tokens = 1800 + (totalChars / 4)
	}

	pct := float64(tokens) / float64(ctxLen) * 100
	if pct > 100 {
		pct = 100
	}

	barWidth := 15
	filled := int((pct / 100.0) * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	var alertStr string
	if pct >= 90.0 {
		alertStr = fmt.Sprintf(" 🚨 [Context Almost Full (%.1f%%) — Run /compact]", pct)
	} else if pct >= 75.0 {
		alertStr = fmt.Sprintf(" ⚠️ [Context Filling Up (%.1f%%)]", pct)
	}

	return fmt.Sprintf("[%s] %.1f%% (%d/%d tokens)%s", bar, pct, tokens, ctxLen, alertStr)
}

func (st *StreamTracker) TransitionToThinking() {
	if st.State != StateThinking {
		st.ThinkingStep++
		if st.State == StateOutputting {
			fmt.Print("\n\n")
		}

		if st.ModelName != "" {
			usageBarStr := st.RenderUsageBar()
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
