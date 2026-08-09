package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileTool(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "subDir", "test_file.txt")
	content := "Hello World from WriteFileTool test!"

	in := WriteFileInput{
		FilePath: targetFile,
		Content:  content,
	}

	out, err := writeFileContents(nil, in)
	if err != nil {
		t.Fatalf("writeFileContents failed: %v", err)
	}

	if out.BytesWritten != len(content) {
		t.Errorf("BytesWritten = %d, want %d", out.BytesWritten, len(content))
	}

	data, rErr := os.ReadFile(targetFile)
	if rErr != nil {
		t.Fatalf("ReadFile failed: %v", rErr)
	}

	if string(data) != content {
		t.Errorf("File content = %q, want %q", string(data), content)
	}
}

func TestWriteFileTool_EmptyPath(t *testing.T) {
	_, err := writeFileContents(nil, WriteFileInput{FilePath: ""})
	if err == nil {
		t.Errorf("Expected error for empty file path, got nil")
	}
}
