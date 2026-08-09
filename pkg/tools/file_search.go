package tools

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

type FileSearchInput struct {
	Pattern string `json:"pattern" jsonschema:"The glob filename pattern to match (e.g. '*.go', '*.json')"`
	RootDir string `json:"root_dir,omitempty" jsonschema:"Optional root directory to search in (defaults to '.')"`
}

type FileSearchOutput struct {
	Pattern string   `json:"pattern"`
	Matches []string `json:"matches"`
	Count   int      `json:"count"`
}

func searchFiles(_ agent.Context, in FileSearchInput) (FileSearchOutput, error) {
	pattern := strings.TrimSpace(in.Pattern)
	if pattern == "" {
		return FileSearchOutput{}, fmt.Errorf("pattern cannot be empty")
	}

	rootDir := in.RootDir
	if rootDir == "" {
		rootDir = "."
	}

	var matches []string
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, wErr error) error {
		if wErr != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		matched, mErr := filepath.Match(pattern, d.Name())
		if mErr == nil && matched {
			matches = append(matches, path)
		}
		return nil
	})

	if err != nil {
		return FileSearchOutput{}, fmt.Errorf("file search failed: %w", err)
	}

	return FileSearchOutput{
		Pattern: pattern,
		Matches: matches,
		Count:   len(matches),
	}, nil
}

func FileSearchTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "file_search",
		Description: "Searches for files matching glob patterns (e.g. '*.go', '*.json') across project subdirectories.",
	}, searchFiles)
}
