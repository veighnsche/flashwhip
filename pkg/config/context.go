package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
	sb.WriteString(fmt.Sprintf("**Working directory**: `%s`\n", abs))
	sb.WriteString(fmt.Sprintf("**Runtime**: `%s/%s`, Go `%s`, started `%s`\n\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), time.Now().Format(time.RFC3339)))

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

	sb.WriteString("\n## Tool Usage Guidelines\nWhen multiple tools could solve a problem, follow this priority:\n1. **Explore first**: `file_search` → `grep_search` → `list_dir` to map the codebase before making changes\n2. **Read before edit**: Always use `read_file` on files you intend to modify—even if you think you know their contents\n3. **Execute with care**: Use `exec_command` for go test, go build, git operations; avoid destructive commands without explicit user confirmation\n4. **Review your work**: After any file changes, run `git_diff` to verify and explain what changed\n5. **External research**: Prefer `web_search` for factual info; use `web_fetch` only when you need full page content\n\n## Workflow Patterns\n- **Debugging**: Reproduce first (read relevant code) → search for error pattern → propose minimal fix → test incrementally\n- **Feature work**: Understand existing architecture (`file_search` + `grep_search`) → plan interface changes → implement → verify with tests or Builds\n- **Refactoring**: Map all usages with `grep_search` → identify safe boundaries → make targeted edits → run check suite\n- **One-liner answers**: Respond directly without tool calls unless the question requires file inspection\n\n## Safety Rules (Non-Negotiable)\n- Never delete files without explicit user instruction\n- Don't mutate files outside the current project context unless told\n- If a tool fails, surface the error cleanly—don't retry aggressively or hide failures\n- When uncertain about project structure, ask before assuming\n")

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
