package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListDirTool(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("abc"), 0644)
	_ = os.Mkdir(filepath.Join(tempDir, "subfolder"), 0755)

	out, err := listDirectoryContents(nil, ListDirInput{DirPath: tempDir})
	if err != nil {
		t.Fatalf("listDirectoryContents failed: %v", err)
	}

	if out.Total != 2 {
		t.Fatalf("out.Total = %d, want 2", out.Total)
	}

	foundFile := false
	foundFolder := false
	for _, e := range out.Entries {
		if e.Name == "file1.txt" && !e.IsDir && e.Size == 3 {
			foundFile = true
		}
		if e.Name == "subfolder" && e.IsDir {
			foundFolder = true
		}
	}

	if !foundFile || !foundFolder {
		t.Errorf("Entries = %+v, expected file1.txt and subfolder", out.Entries)
	}
}
