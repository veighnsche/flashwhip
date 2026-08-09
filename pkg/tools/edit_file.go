package tools

import (
	"fmt"
	"os"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

type EditFileInput struct {
	FilePath           string `json:"file_path" jsonschema:"The local file path to edit"`
	TargetContent      string `json:"target_content" jsonschema:"The exact string or code block to replace"`
	ReplacementContent string `json:"replacement_content" jsonschema:"The replacement string or code block"`
}

type EditFileOutput struct {
	FilePath string `json:"file_path"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

func editFileContents(_ agent.Context, in EditFileInput) (EditFileOutput, error) {
	path := strings.TrimSpace(in.FilePath)
	if path == "" {
		return EditFileOutput{}, fmt.Errorf("file_path cannot be empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return EditFileOutput{}, fmt.Errorf("failed to read file %q: %w", path, err)
	}

	original := string(data)
	if !strings.Contains(original, in.TargetContent) {
		return EditFileOutput{
			FilePath: path,
			Success:  false,
			Message:  fmt.Sprintf("target_content not found in %s", path),
		}, nil
	}

	updated := strings.Replace(original, in.TargetContent, in.ReplacementContent, 1)
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return EditFileOutput{}, fmt.Errorf("failed to save edited file %q: %w", path, err)
	}

	return EditFileOutput{
		FilePath: path,
		Success:  true,
		Message:  fmt.Sprintf("Successfully edited %s", path),
	}, nil
}

func EditFileTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "edit_file",
		Description: "Performs precise search-and-replace block editing on an existing local file.",
	}, editFileContents)
}
