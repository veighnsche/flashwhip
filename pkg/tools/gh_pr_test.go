package tools

import (
	"testing"
)

func TestGHPRToolsInstantiation(t *testing.T) {
	tests := []struct {
		name     string
		toolFunc func() (interface{}, error)
		expected string
	}{
		{
			name: "gh_pr_list",
			toolFunc: func() (interface{}, error) {
				tl, err := GHPRListTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_pr_list",
		},
		{
			name: "gh_pr_view",
			toolFunc: func() (interface{}, error) {
				tl, err := GHPRViewTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_pr_view",
		},
		{
			name: "gh_pr_create",
			toolFunc: func() (interface{}, error) {
				tl, err := GHPRCreateTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_pr_create",
		},
		{
			name: "gh_pr_comment",
			toolFunc: func() (interface{}, error) {
				tl, err := GHPRCommentTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_pr_comment",
		},
		{
			name: "gh_pr_close",
			toolFunc: func() (interface{}, error) {
				tl, err := GHPRCloseTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_pr_close",
		},
		{
			name: "gh_pr_merge",
			toolFunc: func() (interface{}, error) {
				tl, err := GHPRMergeTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_pr_merge",
		},
		{
			name: "gh_pr_diff",
			toolFunc: func() (interface{}, error) {
				tl, err := GHPRDiffTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_pr_diff",
		},
		{
			name: "gh_pr_checks",
			toolFunc: func() (interface{}, error) {
				tl, err := GHPRChecksTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_pr_checks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.toolFunc()
			if err != nil {
				t.Fatalf("unexpected error creating %s tool: %v", tt.name, err)
			}
			if got != tt.expected {
				t.Errorf("expected tool name '%s', got '%v'", tt.expected, got)
			}
		})
	}
}
