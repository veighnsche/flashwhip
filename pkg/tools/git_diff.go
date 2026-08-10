package tools

import (
	"bytes"
	"os/exec"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"flashwhip/pkg/errors"
)

type GitDiffInput struct {
	RepoDir string `json:"repo_dir,omitempty" jsonschema:"Optional path to the git repository root (defaults to project root)"`
	Staged  bool   `json:"staged,omitempty" jsonschema:"If true, shows staged (cached) changes instead of unstaged working-tree changes"`
}

type GitDiffOutput struct {
	RepoDir string   `json:"repo_dir"`
	Staged  bool     `json:"staged"`
	Files   []string `json:"files"`
	Diff    string   `json:"diff"`
	Empty   bool     `json:"empty"`
}

func gitDiff(_ agent.Context, in GitDiffInput) (GitDiffOutput, error) {
	repoDir := resolveRepoDir(in.RepoDir)

	diffArgs := []string{"diff"}
	nameArgs := []string{"diff", "--name-only"}
	if in.Staged {
		diffArgs = append(diffArgs, "--cached")
		nameArgs = append(nameArgs, "--cached")
	}

	cmd := exec.Command("git", diffArgs...)
	cmd.Dir = repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return GitDiffOutput{}, errors.Wrapf(errors.ErrCodeToolExecFailed, err, "git diff failed — %s", strings.TrimSpace(stderr.String()))
	}

	// Fetch the complete list of modified files regardless of diff size so
	// the agent never misses a changed file due to truncation.
	nameCmd := exec.Command("git", nameArgs...)
	nameCmd.Dir = repoDir
	var nameOut, nameErr bytes.Buffer
	nameCmd.Stdout = &nameOut
	nameCmd.Stderr = &nameErr
	if err := nameCmd.Run(); err != nil {
		return GitDiffOutput{}, errors.Wrapf(errors.ErrCodeToolExecFailed, err, "git diff --name-only failed — %s", strings.TrimSpace(nameErr.String()))
	}

	files := []string{}
	for _, line := range strings.Split(strings.TrimSpace(nameOut.String()), "\n") {
		if strings.TrimSpace(line) != "" {
			files = append(files, line)
		}
	}

	raw := stdout.String()
	diff := raw
	truncated := false
	// Cap output to keep token cost reasonable.
	const maxBytes = 8000
	if len(diff) > maxBytes {
		diff = truncateUTF8(diff, maxBytes) + "\n... [diff truncated for token safety]"
		truncated = true
	}

	// Prepend the full modified-file list so the agent can discover every
	// changed file even when the diff body is truncated.
	var b strings.Builder
	if len(files) > 0 {
		b.WriteString("Modified files:\n")
		for _, f := range files {
			b.WriteString("  - ")
			b.WriteString(f)
			b.WriteString("\n")
		}
	} else {
		b.WriteString("No modified files.\n")
	}
	if truncated {
		b.WriteString("Note: diff output was truncated for token safety; the file list above is complete.\n")
	}
	b.WriteString("\n")
	b.WriteString(diff)

	return GitDiffOutput{
		RepoDir: repoDir,
		Staged:  in.Staged,
		Files:   files,
		Diff:    b.String(),
		Empty:   strings.TrimSpace(raw) == "",
	}, nil
}

func GitDiffTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "git_diff",
		Description: "Returns the uncommitted git diff for the working directory, preceded by a complete list of modified files, letting the agent review its own changes. Set staged=true to see staged (cached) changes.",
	}, gitDiff)
}
