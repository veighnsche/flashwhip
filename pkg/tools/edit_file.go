package tools

import (
	"fmt"
	"os"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"flashwhip/pkg/errors"
)

type EditFileInput struct {
	FilePath           string `json:"file_path" jsonschema:"The local file path to edit"`
	TargetContent      string `json:"target_content" jsonschema:"The exact string or code block to replace. Must be unique within the file."`
	ReplacementContent string `json:"replacement_content" jsonschema:"The replacement string or code block"`
}

type EditFileOutput struct {
	FilePath    string `json:"file_path"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Occurrences int    `json:"occurrences,omitempty"`
}

func editFileContents(_ agent.Context, in EditFileInput) (EditFileOutput, error) {
	path := strings.TrimSpace(in.FilePath)
	if path == "" {
		return EditFileOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "file_path cannot be empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return EditFileOutput{}, errors.Wrapf(errors.ErrCodeToolFileNotFound, err, "failed to read file %q", path)
	}

	original := string(data)
	count := strings.Count(original, in.TargetContent)

	if count == 0 {
		return EditFileOutput{
			FilePath:    path,
			Success:     false,
			Occurrences: 0,
			Message:     fmt.Sprintf("target_content not found in %s — verify the exact whitespace and content", path),
		}, nil
	}

	if count > 1 {
		return EditFileOutput{
			FilePath:    path,
			Success:     false,
			Occurrences: count,
			Message:     fmt.Sprintf("target_content matched %d times in %s — provide a longer, more specific string that appears exactly once", count, path),
		}, nil
	}

	updated := strings.Replace(original, in.TargetContent, in.ReplacementContent, 1)
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return EditFileOutput{}, errors.Wrapf(errors.ErrCodeToolPermissionDenied, err, "failed to save edited file %q", path)
	}

	return EditFileOutput{
		FilePath:    path,
		Success:     true,
		Occurrences: 1,
		Message:     fmt.Sprintf("Successfully edited %s", path),
	}, nil
}

func EditFileTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "edit_file",
		Description: "Performs precise search-and-replace editing on a file. target_content must appear exactly once; returns an occurrences count and error if it is missing or ambiguous.",
	}, editFileContents)
}
