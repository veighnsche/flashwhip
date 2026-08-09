package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
