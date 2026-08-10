package ui_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"flashwhip/pkg/provider/ollama"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
	"google.golang.org/genai"
)

// TestStreamToolCallDeliveredOnce pins the invariant that the runner relies on
// in ExecuteStreamLoop (runner.go): every tool call is streamed exactly once as
// a Partial delivery and then folded into the turn-complete content once more.
// The runner records/persists tool calls only on the Partial delivery, so any
// extra partial-emission regresses history dedup and stall detection.
func TestStreamToolCallDeliveredOnce(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/show":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"parameters":"num_ctx 32768","model_info":{"qwen2.context_length":32768}}`)
		case "/v1/chat/completions":
			w.Header().Set("Content-Type", "text/event-stream")
			buf := new(strings.Builder)
			io.Copy(buf, r.Body)
			if strings.Contains(buf.String(), `"tool"`) {
				fmt.Fprint(w, "data: {\"id\":\"2\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Done.\"}}]}\n\n")
			} else {
				fmt.Fprint(w, "data: {\"id\":\"1\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Let me inspect go.mod\"}}]}\n\n")
				fmt.Fprint(w, "data: {\"id\":\"2\",\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"call_1\",\"type\":\"function\",\"function\":{\"name\":\"dummy\",\"arguments\":\"{\\\"x\\\":1}\"}}]}}]}\n\n")
			}
			fmt.Fprint(w, "data: {\"id\":\"3\",\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":5,\"total_tokens\":15}}\n\n")
			fmt.Fprint(w, "data: [DONE]\n\n")
			return
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	m, err := ollama.NewModel("test-qwen", ts.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	dummy, err := functiontool.New(functiontool.Config{Name: "dummy", Description: "noop"}, func(_ agent.Context, in map[string]any) (map[string]any, error) {
		return map[string]any{"ok": true}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	a, err := llmagent.New(llmagent.Config{Name: "test", Model: m, Tools: []tool.Tool{dummy}})
	if err != nil {
		t.Fatal(err)
	}
	sessSvc := session.InMemoryService()
	r, err := runner.New(runner.Config{AppName: "app", Agent: a, SessionService: sessSvc, AutoCreateSession: true})
	if err != nil {
		t.Fatal(err)
	}

	userMsg := &genai.Content{Role: "user", Parts: []*genai.Part{{Text: "read go.mod"}}}

	partialCalls := 0
	finalCalls := 0
	for ev, err := range r.Run(context.Background(), "user", "s1", userMsg, agent.RunConfig{StreamingMode: agent.StreamingModeSSE}) {
		if err != nil {
			t.Fatal(err)
		}
		if ev == nil || ev.Content == nil {
			continue
		}
		for _, p := range ev.Content.Parts {
			if p.FunctionCall != nil {
				if ev.Partial {
					partialCalls++
				} else {
					finalCalls++
				}
			}
		}
	}

	if partialCalls != 1 {
		t.Errorf("streamed tool call partial deliveries = %d, want exactly 1", partialCalls)
	}
	if finalCalls != 1 {
		t.Errorf("turn-complete tool call deliveries = %d, want exactly 1", finalCalls)
	}

	// The ADK session must store each tool call exactly once.
	resp, err := sessSvc.Get(context.Background(), &session.GetRequest{AppName: "app", UserID: "user", SessionID: "s1"})
	if err != nil {
		t.Fatal(err)
	}
	storedCalls := 0
	for ev := range resp.Session.Events().All() {
		if ev.Content == nil {
			continue
		}
		for _, p := range ev.Content.Parts {
			if p.FunctionCall != nil {
				storedCalls++
			}
		}
	}
	if storedCalls != 1 {
		t.Errorf("stored function calls = %d, want exactly 1", storedCalls)
	}
}
