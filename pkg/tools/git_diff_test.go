package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initGitRepo sets up a minimal git repo in dir so git commands succeed.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
}

func TestGitDiffTool_NoChanges(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	// Initial commit so HEAD exists.
	file := filepath.Join(dir, "hello.go")
	_ = os.WriteFile(file, []byte("package main\n"), 0644)
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	_ = cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = dir
	_ = cmd.Run()

	out, err := gitDiff(nil, GitDiffInput{RepoDir: dir})
	if err != nil {
		t.Fatalf("gitDiff failed: %v", err)
	}
	if !out.Empty {
		t.Errorf("expected empty diff, got: %s", out.Diff)
	}
}

func TestGitDiffTool_WithChanges(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	file := filepath.Join(dir, "hello.go")
	_ = os.WriteFile(file, []byte("package main\n"), 0644)
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	_ = cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = dir
	_ = cmd.Run()

	// Make an unstaged change.
	_ = os.WriteFile(file, []byte("package main\n\nfunc main() {}\n"), 0644)

	out, err := gitDiff(nil, GitDiffInput{RepoDir: dir})
	if err != nil {
		t.Fatalf("gitDiff failed: %v", err)
	}
	if out.Empty {
		t.Error("expected non-empty diff")
	}
	if out.Diff == "" {
		t.Error("expected diff content")
	}
}

func TestGitDiffTool_Staged(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	file := filepath.Join(dir, "hello.go")
	_ = os.WriteFile(file, []byte("package main\n"), 0644)
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	_ = cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = dir
	_ = cmd.Run()

	// Stage a change.
	_ = os.WriteFile(file, []byte("package main\n\nfunc main() {}\n"), 0644)
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	_ = cmd.Run()

	out, err := gitDiff(nil, GitDiffInput{RepoDir: dir, Staged: true})
	if err != nil {
		t.Fatalf("gitDiff staged failed: %v", err)
	}
	if out.Empty {
		t.Error("expected non-empty staged diff")
	}
}

func TestGitDiffTool_TruncationListsAllFiles(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	// Two tracked files so the file list has multiple entries.
	for _, name := range []string{"hello.go", "second.go"} {
		p := filepath.Join(dir, name)
		_ = os.WriteFile(p, []byte("package main\n"), 0644)
		cmd := exec.Command("git", "add", ".")
		cmd.Dir = dir
		_ = cmd.Run()
	}
	cmd := exec.Command("git", "commit", "-m", "init")
	cmd.Dir = dir
	_ = cmd.Run()

	// Large changes to force truncation well past the 8000-byte cap.
	big := "package main\n\n"
	for i := 0; i < 400; i++ {
		big += fmt.Sprintf("// line %d %s\n", i, strings.Repeat("x", 80))
	}
	for _, name := range []string{"hello.go", "second.go"} {
		_ = os.WriteFile(filepath.Join(dir, name), []byte(big), 0644)
	}

	out, err := gitDiff(nil, GitDiffInput{RepoDir: dir})
	if err != nil {
		t.Fatalf("gitDiff failed: %v", err)
	}
	if out.Empty {
		t.Error("expected non-empty diff")
	}
	if !strings.Contains(out.Diff, "[diff truncated for token safety]") {
		t.Error("expected diff truncation marker in output")
	}
	if len(out.Files) != 2 {
		t.Fatalf("expected 2 files in file list, got %d: %v", len(out.Files), out.Files)
	}
	for _, want := range []string{"hello.go", "second.go"} {
		if !strings.Contains(out.Diff, "  - "+want) {
			t.Errorf("expected %q in modified-files header of output", want)
		}
	}
}
