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
<function=gh_issue_view>
<parameter=issue_number>
11
</parameter>
<parameter=comments>
True
</parameter>
<parameter=repo_dir>
/Users/vince/Projects/flashwhip
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
	if fnCall.Name != "gh_issue_view" {
		t.Errorf("FunctionCall.Name = %q, want 'gh_issue_view'", fnCall.Name)
	}

	issueNum, ok := fnCall.Args["issue_number"].(int64)
	if !ok || issueNum != 11 {
		t.Errorf("fnCall.Args['issue_number'] = %v (%T), want int64 11", fnCall.Args["issue_number"], fnCall.Args["issue_number"])
	}

	commentsBool, ok := fnCall.Args["comments"].(bool)
	if !ok || !commentsBool {
		t.Errorf("fnCall.Args['comments'] = %v (%T), want bool true", fnCall.Args["comments"], fnCall.Args["comments"])
	}

	repoDir, ok := fnCall.Args["repo_dir"].(string)
	if !ok || repoDir != "/Users/vince/Projects/flashwhip" {
		t.Errorf("fnCall.Args['repo_dir'] = %v, want '/Users/vince/Projects/flashwhip'", fnCall.Args["repo_dir"])
	}
}
