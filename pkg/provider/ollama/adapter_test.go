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

func TestExtractXMLToolCalls(t *testing.T) {
	inputContent := `I will fetch the page for you.
<tool_call>
<function=web_fetch>
<parameter=url>
https://weather.com/today
</parameter>
</function>
</tool_call>`

	cleanText, toolParts := extractXMLToolCalls(inputContent)
	if cleanText != "I will fetch the page for you." {
		t.Errorf("cleanText = %q, want 'I will fetch the page for you.'", cleanText)
	}

	if len(toolParts) != 1 {
		t.Fatalf("len(toolParts) = %d, want 1", len(toolParts))
	}

	fnCall := toolParts[0].FunctionCall
	if fnCall.Name != "web_fetch" {
		t.Errorf("FunctionCall.Name = %q, want 'web_fetch'", fnCall.Name)
	}

	urlVal, ok := fnCall.Args["url"].(string)
	if !ok || urlVal != "https://weather.com/today" {
		t.Errorf("fnCall.Args['url'] = %v, want 'https://weather.com/today'", fnCall.Args["url"])
	}
}
