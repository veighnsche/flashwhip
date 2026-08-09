package ollama

import (
	"testing"
)

func TestSelectAdapter(t *testing.T) {
	tests := []struct {
		modelName string
		isQwen    bool
	}{
		{"hf.co/gbuzhf/KAT-Coder-V2.5-Dev-MTP-GGUF:UD-Q4_K_XL", true},
		{"qwen2.5-coder:7b", true},
		{"deepseek-r1:14b", true},
		{"gpt-4o", false},
	}

	for _, tt := range tests {
		adapter := SelectAdapter(tt.modelName)
		_, ok := adapter.(*QwenDeepSeekAdapter)
		if ok != tt.isQwen {
			t.Errorf("SelectAdapter(%q) isQwen = %v, want %v", tt.modelName, ok, tt.isQwen)
		}
	}
}

func TestSanitizeUTF8(t *testing.T) {
	input := "Hello\x00World\u0000 Test"
	expected := "HelloWorld Test"
	result := sanitizeUTF8(input)
	if result != expected {
		t.Errorf("sanitizeUTF8(%q) = %q, want %q", input, result, expected)
	}
}
