package ollama

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestChatRequest_MaxTokensCustomOverride(t *testing.T) {
	var receivedBody ChatRequest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			resp := ChatResponse{
				ID:    "test-id",
				Model: "test-model",
				Choices: []ChatChoice{
					{Index: 0, Message: ChatChoiceMessage{Role: "assistant", Content: "Ok"}},
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
		Config: &genai.GenerateContentConfig{
			MaxOutputTokens: 8192,
		},
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "Hi"}}},
		},
	}

	ctx := context.Background()
	for _, err := range m.GenerateContent(ctx, req, false) {
		if err != nil {
			t.Fatalf("GenerateContent failed: %v", err)
		}
	}

	if receivedBody.MaxTokens != 8192 {
		t.Errorf("receivedBody.MaxTokens = %d, want custom 8192", receivedBody.MaxTokens)
	}
}

func TestChatRequest_MaxTokensStreamingPayload(t *testing.T) {
	var receivedBody ChatRequest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "data: {\"id\":\"1\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Chunk\"}}]}\n\ndata: [DONE]\n\n")
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
			{Role: "user", Parts: []*genai.Part{{Text: "Stream test"}}},
		},
	}

	ctx := context.Background()
	for _, err := range m.GenerateContent(ctx, req, true) {
		if err != nil {
			t.Fatalf("GenerateContent streaming failed: %v", err)
		}
	}

	if receivedBody.MaxTokens != 4096 {
		t.Errorf("receivedBody.MaxTokens in stream = %d, want 4096", receivedBody.MaxTokens)
	}
	if receivedBody.StreamOps == nil || !receivedBody.StreamOps.IncludeUsage {
		t.Errorf("expected StreamOps.IncludeUsage to be true")
	}
}

func TestChatRequest_MaxTokensRootLevelPlacement(t *testing.T) {
	var rawJSON map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			_ = json.NewDecoder(r.Body).Decode(&rawJSON)
			resp := ChatResponse{
				ID:      "test-id",
				Choices: []ChatChoice{{Index: 0, Message: ChatChoiceMessage{Role: "assistant", Content: "Done"}}},
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
			{Role: "user", Parts: []*genai.Part{{Text: "JSON raw check"}}},
		},
	}

	for _, err := range m.GenerateContent(context.Background(), req, false) {
		if err != nil {
			t.Fatalf("GenerateContent failed: %v", err)
		}
	}

	// Verify max_tokens is present at root level of JSON
	val, exists := rawJSON["max_tokens"]
	if !exists {
		t.Fatalf("max_tokens missing from root level of JSON request payload")
	}
	if floatVal, ok := val.(float64); !ok || int(floatVal) != 4096 {
		t.Errorf("rawJSON['max_tokens'] = %v, want 4096", val)
	}

	// Verify options has num_ctx
	opts, ok := rawJSON["options"].(map[string]any)
	if !ok {
		t.Fatalf("options missing or invalid in JSON request payload")
	}
	if numCtx, ok := opts["num_ctx"].(float64); !ok || int(numCtx) != 32768 {
		t.Errorf("options['num_ctx'] = %v, want 32768", opts["num_ctx"])
	}
}


