package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/adk/v2/model"
	"google.golang.org/genai"
)

func TestFetchAvailableModels(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		resp := ModelListResponse{
			Object: "list",
			Data: []ModelInfo{
				{ID: "qwen2.5-coder:7b"},
				{ID: "KAT-Coder-V2.5"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	models, err := FetchAvailableModels(ts.URL, "")
	if err != nil {
		t.Fatalf("FetchAvailableModels failed: %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("len(models) = %d, want 2", len(models))
	}

	if models[0] != "qwen2.5-coder:7b" || models[1] != "KAT-Coder-V2.5" {
		t.Errorf("Unexpected model names: %v", models)
	}
}

func TestFetchModelContextLength(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/show" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		resp := ShowModelResponse{
			Parameters: "num_ctx 32768\nstop <|im_end|>",
			ModelInfo: map[string]any{
				"general.architecture": "qwen2",
				"qwen2.context_length": float64(32768),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	ctxLen, err := FetchModelContextLength(ts.URL, "", "qwen2.5-coder:7b")
	if err != nil {
		t.Fatalf("FetchModelContextLength failed: %v", err)
	}
	if ctxLen != 32768 {
		t.Errorf("ctxLen = %d, want 32768", ctxLen)
	}
}

func TestFetchModelContextLength_Fallback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	ctxLen, _ := FetchModelContextLength(ts.URL, "", "unknown-model")
	if ctxLen != 32768 {
		t.Errorf("ctxLen fallback = %d, want 32768", ctxLen)
	}
}

func TestChatRequest_MaxTokensPayload(t *testing.T) {
	var receivedBody ChatRequest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			resp := ChatResponse{
				ID:    "test-id",
				Model: "test-model",
				Choices: []ChatChoice{
					{
						Index: 0,
						Message: ChatChoiceMessage{
							Role:    "assistant",
							Content: "Hello",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	m, err := NewModel("test-model", ts.URL, "")
	if err != nil {
		t.Fatalf("NewModel failed: %v", err)
	}

	req := &model.LLMRequest{
		Contents: []*genai.Content{
			{
				Role:  "user",
				Parts: []*genai.Part{{Text: "Hi"}},
			},
		},
	}

	ctx := context.Background()
	for _, err := range m.GenerateContent(ctx, req, false) {
		if err != nil {
			t.Fatalf("GenerateContent failed: %v", err)
		}
	}

	if receivedBody.MaxTokens != 4096 {
		t.Errorf("receivedBody.MaxTokens = %d, want 4096", receivedBody.MaxTokens)
	}
}

