package tools

import (
	"google.golang.org/adk/v2/tool"

	"flashwhip/pkg/errors"
)

// DefaultTools returns the standard set of built-in function tools for Flashwhip.
func DefaultTools() ([]tool.Tool, error) {
	sysTool, err := SystemInfoTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize sysTool", err)
	}

	fileTool, err := ReadFileTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize fileTool", err)
	}

	writeTool, err := WriteFileTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize writeTool", err)
	}

	editTool, err := EditFileTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize editTool", err)
	}

	execTool, err := ExecCommandTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize execTool", err)
	}

	searchTool, err := WebSearchTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize web_search tool", err)
	}

	fetchTool, err := WebFetchTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize web_fetch tool", err)
	}

	listDirTool, err := ListDirTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize list_dir tool", err)
	}

	fileSearchTool, err := FileSearchTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize file_search tool", err)
	}

	grepSearchTool, err := GrepSearchTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize grep_search tool", err)
	}

	gitDiffTool, err := GitDiffTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize git_diff tool", err)
	}

	ghIssueListTool, err := GHIssueListTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_issue_list tool", err)
	}

	ghIssueViewTool, err := GHIssueViewTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_issue_view tool", err)
	}

	ghIssueCreateTool, err := GHIssueCreateTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_issue_create tool", err)
	}

	ghIssueCommentTool, err := GHIssueCommentTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_issue_comment tool", err)
	}

	ghIssueCloseTool, err := GHIssueCloseTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_issue_close tool", err)
	}

	ghPRListTool, err := GHPRListTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_pr_list tool", err)
	}

	ghPRViewTool, err := GHPRViewTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_pr_view tool", err)
	}

	ghPRCreateTool, err := GHPRCreateTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_pr_create tool", err)
	}

	ghPRCommentTool, err := GHPRCommentTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_pr_comment tool", err)
	}

	ghPRCloseTool, err := GHPRCloseTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_pr_close tool", err)
	}

	ghPRMergeTool, err := GHPRMergeTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_pr_merge tool", err)
	}

	ghPRDiffTool, err := GHPRDiffTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_pr_diff tool", err)
	}

	ghPRChecksTool, err := GHPRChecksTool()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeToolInitFailed, "failed to initialize gh_pr_checks tool", err)
	}

	return []tool.Tool{
		sysTool,
		fileTool,
		writeTool,
		editTool,
		execTool,
		searchTool,
		fetchTool,
		listDirTool,
		fileSearchTool,
		grepSearchTool,
		gitDiffTool,
		ghIssueListTool,
		ghIssueViewTool,
		ghIssueCreateTool,
		ghIssueCommentTool,
		ghIssueCloseTool,
		ghPRListTool,
		ghPRViewTool,
		ghPRCreateTool,
		ghPRCommentTool,
		ghPRCloseTool,
		ghPRMergeTool,
		ghPRDiffTool,
		ghPRChecksTool,
	}, nil
}

type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetToolDescriptions returns metadata for all registered tools.
func GetToolDescriptions() ([]ToolInfo, error) {
	toolsList, err := DefaultTools()
	if err != nil {
		return nil, err
	}

	var descriptions []ToolInfo
	for _, t := range toolsList {
		if t != nil {
			descriptions = append(descriptions, ToolInfo{
				Name:        t.Name(),
				Description: t.Description(),
			})
		}
	}
	return descriptions, nil
}
