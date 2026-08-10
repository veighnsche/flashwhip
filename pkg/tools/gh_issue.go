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

	"flashwhip/pkg/config"
	"flashwhip/pkg/errors"
)

func resolveRepoDir(dir string) string {
	dir = strings.TrimSpace(dir)
	if dir == "" || dir == "." {
		return config.GetProjectRoot()
	}
	return dir
}

// runGH executes the gh CLI in repoDir with a 30s timeout, returning raw stdout.
// Errors include the stderr output (or the exit error when none) in the message.
func runGH(repoDir string, args ...string) (string, error) {
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
		return "", errors.Wrapf(errors.ErrCodeToolExecFailed, err, "gh %s %s failed: %s", args[0], args[1], errStr)
	}
	return stdout.String(), nil
}

// GHIssueListInput holds options for listing issues using `gh issue list`.
type GHIssueListInput struct {
	RepoDir  string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	State    string `json:"state,omitempty" jsonschema:"Issue state filter: 'open' (default), 'closed', or 'all'"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum number of issues to fetch (default: 30)"`
	Author   string `json:"author,omitempty" jsonschema:"Filter issues by author username"`
	Assignee string `json:"assignee,omitempty" jsonschema:"Filter issues by assignee username"`
	Label    string `json:"label,omitempty" jsonschema:"Filter issues by label name"`
}

// GHIssueListOutput holds the result of `gh issue list`.
type GHIssueListOutput struct {
	RepoDir string `json:"repo_dir"`
	Output  string `json:"output"`
}

func ghIssueList(_ agent.Context, in GHIssueListInput) (GHIssueListOutput, error) {
	repoDir := resolveRepoDir(in.RepoDir)

	limit := in.Limit
	if limit <= 0 {
		limit = 30
	}

	args := []string{"issue", "list", "--limit", strconv.Itoa(limit)}

	if state := strings.TrimSpace(in.State); state != "" {
		args = append(args, "--state", state)
	}
	if author := strings.TrimSpace(in.Author); author != "" {
		args = append(args, "--author", author)
	}
	if assignee := strings.TrimSpace(in.Assignee); assignee != "" {
		args = append(args, "--assignee", assignee)
	}
	if label := strings.TrimSpace(in.Label); label != "" {
		args = append(args, "--label", label)
	}

	outStr, err := runGH(repoDir, args...)
	if err != nil {
		return GHIssueListOutput{}, err
	}

	const maxBytes = 8000
	if len(outStr) > maxBytes {
		outStr = outStr[:maxBytes] + "\n... [output truncated for token safety]"
	}

	return GHIssueListOutput{
		RepoDir: repoDir,
		Output:  outStr,
	}, nil
}

// GHIssueListTool returns a functiontool wrapping `gh issue list`.
func GHIssueListTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_issue_list",
		Description: "Lists issues in a GitHub repository using the gh CLI. Supports filtering by state (open/closed/all), limit, author, assignee, and label.",
	}, ghIssueList)
}

// GHIssueViewInput holds options for viewing issue details using `gh issue view`.
type GHIssueViewInput struct {
	RepoDir     string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	IssueNumber int    `json:"issue_number" jsonschema:"The issue number to view"`
	Comments    bool   `json:"comments,omitempty" jsonschema:"If true, includes issue comments in the output"`
}

// GHIssueViewOutput holds issue details.
type GHIssueViewOutput struct {
	RepoDir     string `json:"repo_dir"`
	IssueNumber int    `json:"issue_number"`
	Output      string `json:"output"`
}

func ghIssueView(_ agent.Context, in GHIssueViewInput) (GHIssueViewOutput, error) {
	if in.IssueNumber <= 0 {
		return GHIssueViewOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "issue_number must be greater than 0")
	}

	repoDir := resolveRepoDir(in.RepoDir)

	args := []string{"issue", "view", strconv.Itoa(in.IssueNumber)}
	if in.Comments {
		args = append(args, "--comments")
	}

	outStr, err := runGH(repoDir, args...)
	if err != nil {
		return GHIssueViewOutput{}, err
	}

	const maxBytes = 8000
	if len(outStr) > maxBytes {
		outStr = outStr[:maxBytes] + "\n... [output truncated for token safety]"
	}

	return GHIssueViewOutput{
		RepoDir:     repoDir,
		IssueNumber: in.IssueNumber,
		Output:      outStr,
	}, nil
}

// GHIssueViewTool returns a functiontool wrapping `gh issue view`.
func GHIssueViewTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_issue_view",
		Description: "Views details and optional comments for a specific GitHub issue by number using the gh CLI.",
	}, ghIssueView)
}

// GHIssueCreateInput holds options for creating an issue using `gh issue create`.
type GHIssueCreateInput struct {
	RepoDir   string   `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	Title     string   `json:"title" jsonschema:"Title of the issue"`
	Body      string   `json:"body" jsonschema:"Body description of the issue"`
	Labels    []string `json:"labels,omitempty" jsonschema:"Optional list of label names"`
	Assignees []string `json:"assignees,omitempty" jsonschema:"Optional list of assignee usernames"`
}

// GHIssueCreateOutput holds the result of `gh issue create`.
type GHIssueCreateOutput struct {
	RepoDir string `json:"repo_dir"`
	Output  string `json:"output"`
}

func ghIssueCreate(_ agent.Context, in GHIssueCreateInput) (GHIssueCreateOutput, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return GHIssueCreateOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "title cannot be empty")
	}

	repoDir := resolveRepoDir(in.RepoDir)

	args := []string{"issue", "create", "--title", title, "--body", strings.TrimSpace(in.Body)}

	for _, lbl := range in.Labels {
		if l := strings.TrimSpace(lbl); l != "" {
			args = append(args, "--label", l)
		}
	}
	for _, asgn := range in.Assignees {
		if a := strings.TrimSpace(asgn); a != "" {
			args = append(args, "--assignee", a)
		}
	}

	outStr, err := runGH(repoDir, args...)
	if err != nil {
		return GHIssueCreateOutput{}, err
	}

	return GHIssueCreateOutput{
		RepoDir: repoDir,
		Output:  strings.TrimSpace(outStr),
	}, nil
}

// GHIssueCreateTool returns a functiontool wrapping `gh issue create`.
func GHIssueCreateTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_issue_create",
		Description: "Creates a new GitHub issue with title, body, labels, and assignees using the gh CLI.",
	}, ghIssueCreate)
}

// GHIssueCommentInput holds options for adding a comment to an issue using `gh issue comment`.
type GHIssueCommentInput struct {
	RepoDir     string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	IssueNumber int    `json:"issue_number" jsonschema:"The issue number to comment on"`
	Body        string `json:"body" jsonschema:"Comment text body"`
}

// GHIssueCommentOutput holds the result of `gh issue comment`.
type GHIssueCommentOutput struct {
	RepoDir     string `json:"repo_dir"`
	IssueNumber int    `json:"issue_number"`
	Output      string `json:"output"`
}

func ghIssueComment(_ agent.Context, in GHIssueCommentInput) (GHIssueCommentOutput, error) {
	if in.IssueNumber <= 0 {
		return GHIssueCommentOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "issue_number must be greater than 0")
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return GHIssueCommentOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "body cannot be empty")
	}

	repoDir := resolveRepoDir(in.RepoDir)

	args := []string{"issue", "comment", strconv.Itoa(in.IssueNumber), "--body", body}

	outStr, err := runGH(repoDir, args...)
	if err != nil {
		return GHIssueCommentOutput{}, err
	}

	return GHIssueCommentOutput{
		RepoDir:     repoDir,
		IssueNumber: in.IssueNumber,
		Output:      strings.TrimSpace(outStr),
	}, nil
}

// GHIssueCommentTool returns a functiontool wrapping `gh issue comment`.
func GHIssueCommentTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_issue_comment",
		Description: "Adds a comment to a specific GitHub issue by number using the gh CLI.",
	}, ghIssueComment)
}

// GHIssueCloseInput holds options for closing an issue using `gh issue close`.
type GHIssueCloseInput struct {
	RepoDir     string `json:"repo_dir,omitempty" jsonschema:"Optional path to git repository root (defaults to project root)"`
	IssueNumber int    `json:"issue_number" jsonschema:"The issue number to close"`
	Reason      string `json:"reason,omitempty" jsonschema:"Optional closing reason: 'completed' (default) or 'not planned'"`
}

// GHIssueCloseOutput holds the result of `gh issue close`.
type GHIssueCloseOutput struct {
	RepoDir     string `json:"repo_dir"`
	IssueNumber int    `json:"issue_number"`
	Output      string `json:"output"`
}

func ghIssueClose(_ agent.Context, in GHIssueCloseInput) (GHIssueCloseOutput, error) {
	if in.IssueNumber <= 0 {
		return GHIssueCloseOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "issue_number must be greater than 0")
	}

	repoDir := resolveRepoDir(in.RepoDir)

	args := []string{"issue", "close", strconv.Itoa(in.IssueNumber)}
	if reason := strings.TrimSpace(in.Reason); reason != "" {
		args = append(args, "--reason", reason)
	}

	outStr, err := runGH(repoDir, args...)
	if err != nil {
		return GHIssueCloseOutput{}, err
	}

	return GHIssueCloseOutput{
		RepoDir:     repoDir,
		IssueNumber: in.IssueNumber,
		Output:      strings.TrimSpace(outStr),
	}, nil
}

// GHIssueCloseTool returns a functiontool wrapping `gh issue close`.
func GHIssueCloseTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "gh_issue_close",
		Description: "Closes a specific GitHub issue by number using the gh CLI. Supports optional closing reason ('completed' or 'not planned').",
	}, ghIssueClose)
}
