package tools

import (
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"flashwhip/pkg/errors"
)

// GHPRListInput holds options for listing pull requests using `gh pr list`.
type GHPRListInput struct {
	RepoDir string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	State   string `json:"state,omitempty" jsonschema:"PR state filter: 'open' (default), 'closed', 'merged', or 'all'"`
	Limit   int    `json:"limit,omitempty" jsonschema:"Maximum number of pull requests to fetch (default: 30)"`
	Author  string `json:"author,omitempty" jsonschema:"Filter pull requests by author username"`
	Base    string `json:"base,omitempty" jsonschema:"Filter pull requests by target base branch"`
}

// GHPRListOutput holds the result of `gh pr list`.
type GHPRListOutput struct {
	RepoDir string `json:"repo_dir"`
	Output  string `json:"output"`
}

func ghPRList(_ agent.Context, in GHPRListInput) (GHPRListOutput, error) {
	repoDir := resolveRepoDir(in.RepoDir)

	limit := in.Limit
	if limit <= 0 {
		limit = 30
	}

	args := []string{"pr", "list", "--limit", strconv.Itoa(limit)}

	if state := strings.TrimSpace(in.State); state != "" {
		args = append(args, "--state", state)
	}
	if author := strings.TrimSpace(in.Author); author != "" {
		args = append(args, "--author", author)
	}
	if base := strings.TrimSpace(in.Base); base != "" {
		args = append(args, "--base", base)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = err.Error()
		}
		return GHPRListOutput{}, errors.Wrapf(errors.ErrCodeToolExecFailed, err, "gh pr list failed: %s", errStr)
	}

	outStr := stdout.String()
	const maxBytes = 8000
	if len(outStr) > maxBytes {
		outStr = outStr[:maxBytes] + "\n... [output truncated for token safety]"
	}

	return GHPRListOutput{
		RepoDir: repoDir,
		Output:  outStr,
	}, nil
}

// GHPRListTool returns a functiontool wrapping `gh pr list`.
func GHPRListTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_pr_list",
		Description: "Lists pull requests in a GitHub repository using the gh CLI. Supports filtering by state (open/closed/merged/all), limit, author, and base branch.",
	}, ghPRList)
}

// GHPRViewInput holds options for viewing pull request details using `gh pr view`.
type GHPRViewInput struct {
	RepoDir  string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	PRNumber int    `json:"pr_number" jsonschema:"The pull request number to view"`
	Comments bool   `json:"comments,omitempty" jsonschema:"If true, includes PR comments in the output"`
}

// GHPRViewOutput holds pull request details.
type GHPRViewOutput struct {
	RepoDir  string `json:"repo_dir"`
	PRNumber int    `json:"pr_number"`
	Output   string `json:"output"`
}

func ghPRView(_ agent.Context, in GHPRViewInput) (GHPRViewOutput, error) {
	if in.PRNumber <= 0 {
		return GHPRViewOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "pr_number must be greater than 0")
	}

	repoDir := resolveRepoDir(in.RepoDir)

	args := []string{"pr", "view", strconv.Itoa(in.PRNumber)}
	if in.Comments {
		args = append(args, "--comments")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = err.Error()
		}
		return GHPRViewOutput{}, errors.Wrapf(errors.ErrCodeToolExecFailed, err, "gh pr view failed: %s", errStr)
	}

	outStr := stdout.String()
	const maxBytes = 8000
	if len(outStr) > maxBytes {
		outStr = outStr[:maxBytes] + "\n... [output truncated for token safety]"
	}

	return GHPRViewOutput{
		RepoDir:  repoDir,
		PRNumber: in.PRNumber,
		Output:   outStr,
	}, nil
}

// GHPRViewTool returns a functiontool wrapping `gh pr view`.
func GHPRViewTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_pr_view",
		Description: "Views details, status checks, and optional comments for a specific GitHub pull request by number using the gh CLI.",
	}, ghPRView)
}

// GHPRCreateInput holds options for creating a PR using `gh pr create`.
type GHPRCreateInput struct {
	RepoDir string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	Title   string `json:"title" jsonschema:"Title of the pull request"`
	Body    string `json:"body" jsonschema:"Body description of the pull request"`
	Base    string `json:"base,omitempty" jsonschema:"Target base branch for the pull request (e.g. 'main')"`
	Head    string `json:"head,omitempty" jsonschema:"Head feature branch containing the changes"`
	Draft   bool   `json:"draft,omitempty" jsonschema:"If true, creates the pull request as a draft"`
}

// GHPRCreateOutput holds the result of `gh pr create`.
type GHPRCreateOutput struct {
	RepoDir string `json:"repo_dir"`
	Output  string `json:"output"`
}

func ghPRCreate(_ agent.Context, in GHPRCreateInput) (GHPRCreateOutput, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return GHPRCreateOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "title cannot be empty")
	}

	repoDir := resolveRepoDir(in.RepoDir)

	args := []string{"pr", "create", "--title", title, "--body", strings.TrimSpace(in.Body)}

	if base := strings.TrimSpace(in.Base); base != "" {
		args = append(args, "--base", base)
	}
	if head := strings.TrimSpace(in.Head); head != "" {
		args = append(args, "--head", head)
	}
	if in.Draft {
		args = append(args, "--draft")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = err.Error()
		}
		return GHPRCreateOutput{}, errors.Wrapf(errors.ErrCodeToolExecFailed, err, "gh pr create failed: %s", errStr)
	}

	return GHPRCreateOutput{
		RepoDir: repoDir,
		Output:  strings.TrimSpace(stdout.String()),
	}, nil
}

// GHPRCreateTool returns a functiontool wrapping `gh pr create`.
func GHPRCreateTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_pr_create",
		Description: "Creates a new GitHub pull request with title, body, target base branch, feature head branch, and draft flag using the gh CLI.",
	}, ghPRCreate)
}

// GHPRCommentInput holds options for adding a comment to a PR using `gh pr comment`.
type GHPRCommentInput struct {
	RepoDir  string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	PRNumber int    `json:"pr_number" jsonschema:"The pull request number to comment on"`
	Body     string `json:"body" jsonschema:"Comment text body"`
}

// GHPRCommentOutput holds the result of `gh pr comment`.
type GHPRCommentOutput struct {
	RepoDir  string `json:"repo_dir"`
	PRNumber int    `json:"pr_number"`
	Output   string `json:"output"`
}

func ghPRComment(_ agent.Context, in GHPRCommentInput) (GHPRCommentOutput, error) {
	if in.PRNumber <= 0 {
		return GHPRCommentOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "pr_number must be greater than 0")
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return GHPRCommentOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "body cannot be empty")
	}

	repoDir := resolveRepoDir(in.RepoDir)

	args := []string{"pr", "comment", strconv.Itoa(in.PRNumber), "--body", body}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = err.Error()
		}
		return GHPRCommentOutput{}, errors.Wrapf(errors.ErrCodeToolExecFailed, err, "gh pr comment failed: %s", errStr)
	}

	return GHPRCommentOutput{
		RepoDir:  repoDir,
		PRNumber: in.PRNumber,
		Output:   strings.TrimSpace(stdout.String()),
	}, nil
}

// GHPRCommentTool returns a functiontool wrapping `gh pr comment`.
func GHPRCommentTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_pr_comment",
		Description: "Adds a comment to a specific GitHub pull request by number using the gh CLI.",
	}, ghPRComment)
}

// GHPRCloseInput holds options for closing a PR using `gh pr close`.
type GHPRCloseInput struct {
	RepoDir      string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	PRNumber     int    `json:"pr_number" jsonschema:"The pull request number to close"`
	DeleteBranch bool   `json:"delete_branch,omitempty" jsonschema:"If true, deletes the local and remote branch after closing"`
}

// GHPRCloseOutput holds the result of `gh pr close`.
type GHPRCloseOutput struct {
	RepoDir  string `json:"repo_dir"`
	PRNumber int    `json:"pr_number"`
	Output   string `json:"output"`
}

func ghPRClose(_ agent.Context, in GHPRCloseInput) (GHPRCloseOutput, error) {
	if in.PRNumber <= 0 {
		return GHPRCloseOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "pr_number must be greater than 0")
	}

	repoDir := resolveRepoDir(in.RepoDir)

	args := []string{"pr", "close", strconv.Itoa(in.PRNumber)}
	if in.DeleteBranch {
		args = append(args, "--delete-branch")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = err.Error()
		}
		return GHPRCloseOutput{}, errors.Wrapf(errors.ErrCodeToolExecFailed, err, "gh pr close failed: %s", errStr)
	}

	return GHPRCloseOutput{
		RepoDir:  repoDir,
		PRNumber: in.PRNumber,
		Output:   strings.TrimSpace(stdout.String()),
	}, nil
}

// GHPRCloseTool returns a functiontool wrapping `gh pr close`.
func GHPRCloseTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_pr_close",
		Description: "Closes a specific GitHub pull request by number using the gh CLI. Supports optional branch deletion.",
	}, ghPRClose)
}

// GHPRMergeInput holds options for merging a PR using `gh pr merge`.
type GHPRMergeInput struct {
	RepoDir      string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	PRNumber     int    `json:"pr_number" jsonschema:"The pull request number to merge"`
	MergeMethod  string `json:"merge_method,omitempty" jsonschema:"Merge strategy: 'squash' (default), 'merge', or 'rebase'"`
	DeleteBranch bool   `json:"delete_branch,omitempty" jsonschema:"If true, deletes the local and remote feature branch after merge"`
}

// GHPRMergeOutput holds the result of `gh pr merge`.
type GHPRMergeOutput struct {
	RepoDir  string `json:"repo_dir"`
	PRNumber int    `json:"pr_number"`
	Output   string `json:"output"`
}

func ghPRMerge(_ agent.Context, in GHPRMergeInput) (GHPRMergeOutput, error) {
	if in.PRNumber <= 0 {
		return GHPRMergeOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "pr_number must be greater than 0")
	}

	repoDir := resolveRepoDir(in.RepoDir)

	args := []string{"pr", "merge", strconv.Itoa(in.PRNumber)}

	method := strings.ToLower(strings.TrimSpace(in.MergeMethod))
	switch method {
	case "merge":
		args = append(args, "--merge")
	case "rebase":
		args = append(args, "--rebase")
	default:
		args = append(args, "--squash")
	}

	if in.DeleteBranch {
		args = append(args, "--delete-branch")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = err.Error()
		}
		return GHPRMergeOutput{}, errors.Wrapf(errors.ErrCodeToolExecFailed, err, "gh pr merge failed: %s", errStr)
	}

	return GHPRMergeOutput{
		RepoDir:  repoDir,
		PRNumber: in.PRNumber,
		Output:   strings.TrimSpace(stdout.String()),
	}, nil
}

// GHPRMergeTool returns a functiontool wrapping `gh pr merge`.
func GHPRMergeTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_pr_merge",
		Description: "Merges a GitHub pull request by number using the gh CLI. Supports squash, merge, and rebase strategies.",
	}, ghPRMerge)
}

// GHPRDiffInput holds options for viewing code diff of a PR using `gh pr diff`.
type GHPRDiffInput struct {
	RepoDir  string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	PRNumber int    `json:"pr_number" jsonschema:"The pull request number to view diff for"`
}

// GHPRDiffOutput holds the PR code diff.
type GHPRDiffOutput struct {
	RepoDir  string `json:"repo_dir"`
	PRNumber int    `json:"pr_number"`
	Diff     string `json:"diff"`
}

func ghPRDiff(_ agent.Context, in GHPRDiffInput) (GHPRDiffOutput, error) {
	if in.PRNumber <= 0 {
		return GHPRDiffOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "pr_number must be greater than 0")
	}

	repoDir := resolveRepoDir(in.RepoDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", "pr", "diff", strconv.Itoa(in.PRNumber))
	cmd.Dir = repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = err.Error()
		}
		return GHPRDiffOutput{}, errors.Wrapf(errors.ErrCodeToolExecFailed, err, "gh pr diff failed: %s", errStr)
	}

	outStr := stdout.String()
	const maxBytes = 8000
	if len(outStr) > maxBytes {
		outStr = outStr[:maxBytes] + "\n... [diff truncated for token safety]"
	}

	return GHPRDiffOutput{
		RepoDir:  repoDir,
		PRNumber: in.PRNumber,
		Diff:     outStr,
	}, nil
}

// GHPRDiffTool returns a functiontool wrapping `gh pr diff`.
func GHPRDiffTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_pr_diff",
		Description: "Views the unmerged code diff for a specific GitHub pull request by number using the gh CLI.",
	}, ghPRDiff)
}

// GHPRChecksInput holds options for viewing status checks using `gh pr checks`.
type GHPRChecksInput struct {
	RepoDir  string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	PRNumber int    `json:"pr_number" jsonschema:"The pull request number to view status checks for"`
}

// GHPRChecksOutput holds the status checks output.
type GHPRChecksOutput struct {
	RepoDir  string `json:"repo_dir"`
	PRNumber int    `json:"pr_number"`
	Output   string `json:"output"`
}

func ghPRChecks(_ agent.Context, in GHPRChecksInput) (GHPRChecksOutput, error) {
	if in.PRNumber <= 0 {
		return GHPRChecksOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "pr_number must be greater than 0")
	}

	repoDir := resolveRepoDir(in.RepoDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", "pr", "checks", strconv.Itoa(in.PRNumber))
	cmd.Dir = repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = err.Error()
		}
		return GHPRChecksOutput{}, errors.Wrapf(errors.ErrCodeToolExecFailed, err, "gh pr checks failed: %s", errStr)
	}

	return GHPRChecksOutput{
		RepoDir:  repoDir,
		PRNumber: in.PRNumber,
		Output:   strings.TrimSpace(stdout.String()),
	}, nil
}

// GHPRChecksTool returns a functiontool wrapping `gh pr checks`.
func GHPRChecksTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_pr_checks",
		Description: "Views CI build and test status checks for a specific GitHub pull request by number using the gh CLI.",
	}, ghPRChecks)
}
