package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BuildProjectContext returns a compact, human-readable block describing the
// active project directory. It is injected into the system instruction so the
// agent knows where it is operating from the very first turn.
func BuildProjectContext(projectRoot string) string {
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ""
		}
		projectRoot = cwd
	}

	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		abs = projectRoot
	}

	var sb strings.Builder
	sb.WriteString("\n\n---\n")
	sb.WriteString("## Active Project Context\n\n")
	sb.WriteString(fmt.Sprintf("**Working directory**: `%s`\n\n", abs))

	// Git branch (best-effort, silent on failure)
	if branch := gitBranch(abs); branch != "" {
		sb.WriteString(fmt.Sprintf("**Git branch**: `%s`\n\n", branch))
	}

	// Shallow directory tree (depth 1, skip .git / node_modules / vendor)
	sb.WriteString("**Top-level directory layout**:\n```\n")
	entries, err := os.ReadDir(abs)
	if err == nil {
		for _, e := range entries {
			name := e.Name()
			if name == ".git" {
				continue
			}
			if e.IsDir() {
				sb.WriteString(fmt.Sprintf("  %s/\n", name))
			} else {
				sb.WriteString(fmt.Sprintf("  %s\n", name))
			}
		}
	}
	sb.WriteString("```\n\n")

	sb.WriteString("**Coding harness rules**:\n")
	sb.WriteString("- Always read a file before writing or editing it.\n")
	sb.WriteString("- Prefer small, targeted edits over full rewrites.\n")
	sb.WriteString("- After making changes, call `git_diff` to review your own edits.\n")
	sb.WriteString("- Use `grep_search` and `file_search` to orient yourself before assuming file locations.\n")
	sb.WriteString("- Do not delete files unless explicitly instructed.\n")
	sb.WriteString("---")

	return sb.String()
}

// gitBranch returns the current git branch name for the given directory,
// or an empty string if the directory is not a git repo or git is unavailable.
func gitBranch(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	branch := strings.TrimSpace(string(out))
	if branch == "HEAD" {
		// Detached HEAD — try to get the short hash instead
		cmd2 := exec.Command("git", "rev-parse", "--short", "HEAD")
		cmd2.Dir = dir
		out2, err2 := cmd2.Output()
		if err2 == nil {
			return strings.TrimSpace(string(out2)) + " (detached)"
		}
	}
	return branch
}
