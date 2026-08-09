package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileSearchTool(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "sub")
	_ = os.MkdirAll(subDir, 0755)

	_ = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(subDir, "helper.go"), []byte("package sub"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Title"), 0644)

	out, err := searchFiles(nil, FileSearchInput{Pattern: "*.go", RootDir: tempDir})
	if err != nil {
		t.Fatalf("searchFiles failed: %v", err)
	}

	if out.Count != 2 {
		t.Fatalf("out.Count = %d, want 2", out.Count)
	}
}
