package tools

import (
	"os"
	"os/exec"
	"path/filepath"
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
