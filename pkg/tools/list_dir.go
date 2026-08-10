package tools

import (
	"os"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"flashwhip/pkg/errors"
)

type ListDirInput struct {
	DirPath string `json:"dir_path,omitempty" jsonschema:"Relative or absolute directory path to inspect (defaults to project root)"`
}

type DirEntryInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size_bytes"`
}

type ListDirOutput struct {
	DirPath string         `json:"dir_path"`
	Entries []DirEntryInfo `json:"entries"`
	Total   int            `json:"total_entries"`
}

func listDirectoryContents(_ agent.Context, in ListDirInput) (ListDirOutput, error) {
	dirPath := resolveRepoDir(in.DirPath)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return ListDirOutput{}, errors.Wrapf(errors.ErrCodeToolFileNotFound, err, "failed to list directory %q", dirPath)
	}

	var results []DirEntryInfo
	for _, e := range entries {
		info, iErr := e.Info()
		size := int64(0)
		if iErr == nil {
			size = info.Size()
		}
		results = append(results, DirEntryInfo{
			Name:  e.Name(),
			IsDir: e.IsDir(),
			Size:  size,
		})
	}

	return ListDirOutput{
		DirPath: dirPath,
		Entries: results,
		Total:   len(results),
	}, nil
}

func ListDirTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "list_dir",
		Description: "Lists all files, subdirectories, and sizes inside a target directory path.",
	}, listDirectoryContents)
}
