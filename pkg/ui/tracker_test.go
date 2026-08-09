package ui

import (
	"testing"
)

func TestStreamTracker_StateTransitions(t *testing.T) {
	tracker := NewStreamTracker()

	if tracker.State != StateIdle {
		t.Errorf("Initial state = %v, want StateIdle", tracker.State)
	}

	tracker.TransitionToThinking()
	if tracker.State != StateThinking {
		t.Errorf("State after TransitionToThinking = %v, want StateThinking", tracker.State)
	}
	if tracker.ThinkingStep != 1 {
		t.Errorf("ThinkingStep = %d, want 1", tracker.ThinkingStep)
	}

	tracker.TransitionToOutputting()
	if tracker.State != StateOutputting {
		t.Errorf("State after TransitionToOutputting = %v, want StateOutputting", tracker.State)
	}

	tracker.TransitionToIdle()
	if tracker.State != StateIdle {
		t.Errorf("State after TransitionToIdle = %v, want StateIdle", tracker.State)
	}
}
