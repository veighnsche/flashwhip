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

func TestGrepSearchRegex(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.go")
	content := []byte("package main\nfunc Hello() {}\nfunc World() {}\nfunc Greet() {}\n")
	_ = os.WriteFile(file1, content, 0644)

	// Test regex search - match functions starting with capital letter followed by any letters
	out, err := grepSearchCodebase(nil, GrepSearchInput{Query: "^func [A-Z][a-zA-Z]+\\(\\)", RootDir: tempDir, IsRegex: true})
	if err != nil {
		t.Fatalf("grepSearchCodebase with regex failed: %v", err)
	}

	if out.TotalMatches != 3 {
		t.Fatalf("out.TotalMatches = %d, want 3", out.TotalMatches)
	}
}

func TestGrepSearchInvalidRegex(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.go")
	_ = os.WriteFile(file1, []byte("test"), 0644)

	_, err := grepSearchCodebase(nil, GrepSearchInput{Query: "[invalid", RootDir: tempDir, IsRegex: true})
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}

// TestGrepSearchAlternative demonstrates regex alternation support.
// Note: Go regexp (RE2) does not support backreferences (\1, \2), so this test uses | instead.
func TestGrepSearchAlternative(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.go")
	content := []byte("hello world\nfoo bar\nhello foo\n")
	_ = os.WriteFile(file1, content, 0644)

	// Match lines containing either "world" or "bar"
	out, err := grepSearchCodebase(nil, GrepSearchInput{Query: `world|bar`, RootDir: tempDir, IsRegex: true})
	if err != nil {
		t.Fatalf("grepSearchCodebase with alternative failed: %v", err)
	}

	if out.TotalMatches != 2 {
		t.Fatalf("out.TotalMatches = %d, want 2 (should match 'hello world' and 'foo bar')", out.TotalMatches)
	}
}
