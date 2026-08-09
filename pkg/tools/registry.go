package tools

import (
	"fmt"

	"google.golang.org/adk/v2/tool"
)

// DefaultTools returns the standard set of built-in function tools for Flashwhip.
func DefaultTools() ([]tool.Tool, error) {
	sysTool, err := SystemInfoTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sysTool: %w", err)
	}

	fileTool, err := ReadFileTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize fileTool: %w", err)
	}

	writeTool, err := WriteFileTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize writeTool: %w", err)
	}

	editTool, err := EditFileTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize editTool: %w", err)
	}

	execTool, err := ExecCommandTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize execTool: %w", err)
	}

	searchTool, err := WebSearchTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize web_search tool: %w", err)
	}

	fetchTool, err := WebFetchTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize web_fetch tool: %w", err)
	}

	listDirTool, err := ListDirTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize list_dir tool: %w", err)
	}

	fileSearchTool, err := FileSearchTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize file_search tool: %w", err)
	}

	grepSearchTool, err := GrepSearchTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize grep_search tool: %w", err)
	}

	gitDiffTool, err := GitDiffTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git_diff tool: %w", err)
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
