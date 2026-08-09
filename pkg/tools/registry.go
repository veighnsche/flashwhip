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

	searchTool, err := WebSearchTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize web_search tool: %w", err)
	}

	fetchTool, err := WebFetchTool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize web_fetch tool: %w", err)
	}

	return []tool.Tool{
		sysTool,
		fileTool,
		searchTool,
		fetchTool,
	}, nil
}
