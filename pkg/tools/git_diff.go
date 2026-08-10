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
	RepoDir string `json:"repo_dir"`
	Staged  bool   `json:"staged"`
	Diff    string `json:"diff"`
	Empty   bool   `json:"empty"`
}

func gitDiff(_ agent.Context, in GitDiffInput) (GitDiffOutput, error) {
	repoDir := resolveRepoDir(in.RepoDir)

	args := []string{"diff"}
	if in.Staged {
		args = append(args, "--cached")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return GitDiffOutput{}, errors.Wrapf(errors.ErrCodeToolExecFailed, err, "git diff failed — %s", strings.TrimSpace(stderr.String()))
	}

	diff := stdout.String()
	// Cap output to keep token cost reasonable.
	const maxBytes = 8000
	if len(diff) > maxBytes {
		diff = diff[:maxBytes] + "\n... [diff truncated for token safety]"
	}

	return GitDiffOutput{
		RepoDir: repoDir,
		Staged:  in.Staged,
		Diff:    diff,
		Empty:   strings.TrimSpace(diff) == "",
	}, nil
}

func GitDiffTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "git_diff",
		Description: "Returns the uncommitted git diff for the working directory, letting the agent review its own changes. Set staged=true to see staged (cached) changes.",
	}, gitDiff)
}
