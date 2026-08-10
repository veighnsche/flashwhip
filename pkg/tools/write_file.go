package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"flashwhip/pkg/errors"
)

type WriteFileInput struct {
	FilePath string `json:"file_path" jsonschema:"The relative or absolute local file path to write to"`
	Content  string `json:"content" jsonschema:"The code or text content to write"`
}

type WriteFileOutput struct {
	FilePath     string `json:"file_path"`
	BytesWritten int    `json:"bytes_written"`
	Message      string `json:"message"`
}

func writeFileContents(_ agent.Context, in WriteFileInput) (WriteFileOutput, error) {
	path := strings.TrimSpace(in.FilePath)
	if path == "" {
		return WriteFileOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "file_path cannot be empty")
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return WriteFileOutput{}, errors.Wrapf(errors.ErrCodeToolPermissionDenied, err, "failed to create directory %q", dir)
		}
	}

	data := []byte(in.Content)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return WriteFileOutput{}, errors.Wrapf(errors.ErrCodeToolPermissionDenied, err, "failed to write file %q", path)
	}

	return WriteFileOutput{
		FilePath:     path,
		BytesWritten: len(data),
		Message:      fmt.Sprintf("Successfully wrote %d bytes to %s", len(data), path),
	}, nil
}

func WriteFileTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "write_file",
		Description: "Creates or overwrites a local file with the provided code or text content.",
	}, writeFileContents)
}
