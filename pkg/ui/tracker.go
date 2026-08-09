package ui

import (
	"fmt"
	"os"
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
