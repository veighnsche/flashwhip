package middleware

import (
	"strings"
	"testing"

	"google.golang.org/genai"
)

func TestPruneContents(t *testing.T) {
	// Create 8 turns (16 content objects)
	contents := []*genai.Content{
		// Historical turn 1 (index 0, 1)
		{Role: "user", Parts: []*genai.Part{{Text: "Prompt 1"}}},
		{Role: "model", Parts: []*genai.Part{
			{Text: "Scratchpad reasoning for turn 1", Thought: true},
			{FunctionResponse: &genai.FunctionResponse{
				Name:     "web_fetch",
				Response: map[string]any{"data": strings.Repeat("A", 300)},
			}},
			{Text: "Final answer for turn 1"},
		}},
		// Historical turn 2 (index 2, 3)
		{Role: "user", Parts: []*genai.Part{{Text: "Prompt 2"}}},
		{Role: "model", Parts: []*genai.Part{
			{Text: "Final answer for turn 2"},
		}},
		// Historical turn 3 (index 4, 5)
		{Role: "user", Parts: []*genai.Part{{Text: "Prompt 3"}}},
		{Role: "model", Parts: []*genai.Part{
			{Text: "Final answer for turn 3"},
		}},
		// Recent turn 4 (index 6, 7) - within maxHistoryTurns=1
		{Role: "user", Parts: []*genai.Part{{Text: "Prompt 4"}}},
		{Role: "model", Parts: []*genai.Part{
			{Text: "Current reasoning", Thought: true},
			{FunctionResponse: &genai.FunctionResponse{
				Name:     "web_fetch",
				Response: map[string]any{"data": strings.Repeat("B", 3000)}, // Exceeds 2,000 limit
			}},
			{Text: "Current answer"},
		}},
	}

	pruned := PruneContents(contents, 1) // Keep last 1 turn intact

	// 1. Verify historical thought block in index 1 was stripped
	for _, p := range pruned[1].Parts {
		if p.Thought {
			t.Errorf("Historical thought block in turn 1 was not stripped")
		}
	}

	// 2. Verify historical tool response in index 1 was summarized to <= 200 bytes
	for _, p := range pruned[1].Parts {
		if p.FunctionResponse != nil {
			summary, ok := p.FunctionResponse.Response["summary"].(string)
			if !ok || !strings.Contains(summary, "truncated") {
				t.Errorf("Expected truncated summary, got %v", p.FunctionResponse.Response)
			}
		}
	}

	// 3. Verify recent thought block in index 7 was PRESERVED
	foundRecentThought := false
	for _, p := range pruned[7].Parts {
		if p.Thought && p.Text == "Current reasoning" {
			foundRecentThought = true
		}
	}
	if !foundRecentThought {
		t.Errorf("Recent thought block in turn 4 was lost")
	}

	// 4. Verify giant tool output in recent turn 4 was capped at 2,000 bytes
	for _, p := range pruned[7].Parts {
		if p.FunctionResponse != nil {
			summary, ok := p.FunctionResponse.Response["summary"].(string)
			if !ok || len(summary) > 2200 {
				t.Errorf("Giant tool response in recent turn was not capped, len = %d", len(summary))
			}
		}
	}
}
