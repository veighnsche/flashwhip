package tools

import (
	"testing"

	"flashwhip/pkg/config"
)

func TestGHIssueToolsInstantiation(t *testing.T) {
	tests := []struct {
		name     string
		toolFunc func() (interface{}, error)
		expected string
	}{
		{
			name: "gh_issue_list",
			toolFunc: func() (interface{}, error) {
				tl, err := GHIssueListTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_issue_list",
		},
		{
			name: "gh_issue_view",
			toolFunc: func() (interface{}, error) {
				tl, err := GHIssueViewTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_issue_view",
		},
		{
			name: "gh_issue_create",
			toolFunc: func() (interface{}, error) {
				tl, err := GHIssueCreateTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_issue_create",
		},
		{
			name: "gh_issue_comment",
			toolFunc: func() (interface{}, error) {
				tl, err := GHIssueCommentTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_issue_comment",
		},
		{
			name: "gh_issue_close",
			toolFunc: func() (interface{}, error) {
				tl, err := GHIssueCloseTool()
				if err != nil {
					return nil, err
				}
				return tl.Name(), nil
			},
			expected: "gh_issue_close",
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

func TestResolveRepoDir_DefaultsToProjectRoot(t *testing.T) {
	customRoot := "/tmp/test-project-root"
	config.SetProjectRoot(customRoot)

	tests := []struct {
		input    string
		expected string
	}{
		{input: "", expected: customRoot},
		{input: "  ", expected: customRoot},
		{input: ".", expected: customRoot},
		{input: "/custom/path", expected: "/custom/path"},
	}

	for _, tt := range tests {
		got := resolveRepoDir(tt.input)
		if got != tt.expected {
			t.Errorf("resolveRepoDir(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

