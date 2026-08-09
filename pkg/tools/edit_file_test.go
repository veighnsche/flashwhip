package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEditFileTool(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "code.go")
	originalContent := "package main\n\nfunc Hello() {\n\tfmt.Println(\"Old\")\n}"
	_ = os.WriteFile(targetFile, []byte(originalContent), 0644)

	in := EditFileInput{
		FilePath:           targetFile,
		TargetContent:      "fmt.Println(\"Old\")",
		ReplacementContent: "fmt.Println(\"New\")",
	}

	out, err := editFileContents(nil, in)
	if err != nil {
		t.Fatalf("editFileContents failed: %v", err)
	}

	if !out.Success {
		t.Errorf("Expected out.Success == true, got false")
	}

	data, _ := os.ReadFile(targetFile)
	expectedContent := "package main\n\nfunc Hello() {\n\tfmt.Println(\"New\")\n}"
	if string(data) != expectedContent {
		t.Errorf("Edited content = %q, want %q", string(data), expectedContent)
	}
}

func TestEditFileTool_TargetNotFound(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "code.go")
	_ = os.WriteFile(targetFile, []byte("hello"), 0644)

	in := EditFileInput{
		FilePath:           targetFile,
		TargetContent:      "missing_string",
		ReplacementContent: "new_string",
	}

	out, err := editFileContents(nil, in)
	if err != nil {
		t.Fatalf("editFileContents returned error: %v", err)
	}

	if out.Success {
		t.Errorf("Expected out.Success == false when target content is missing")
	}
}
