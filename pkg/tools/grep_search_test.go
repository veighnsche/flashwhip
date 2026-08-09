package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGrepSearchTool(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.go")
	_ = os.WriteFile(file1, []byte("package main\n\nfunc TargetFunction() {}\n"), 0644)

	out, err := grepSearchCodebase(nil, GrepSearchInput{Query: "TargetFunction", RootDir: tempDir})
	if err != nil {
		t.Fatalf("grepSearchCodebase failed: %v", err)
	}

	if out.TotalMatches != 1 {
		t.Fatalf("out.TotalMatches = %d, want 1", out.TotalMatches)
	}

	if out.Matches[0].LineNumber != 3 {
		t.Errorf("LineNumber = %d, want 3", out.Matches[0].LineNumber)
	}
}
