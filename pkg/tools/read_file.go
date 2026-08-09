package tools

import (
	"fmt"
	"os"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

type ReadFileArgs struct {
	FilePath string `json:"file_path"`
}

type ReadFileOutput struct {
	Content string `json:"content"`
	Bytes   int    `json:"bytes"`
}

func readFileSnippet(_ agent.Context, args ReadFileArgs) (ReadFileOutput, error) {
	data, err := os.ReadFile(args.FilePath)
	if err != nil {
		return ReadFileOutput{}, fmt.Errorf("failed to read file %q: %w", args.FilePath, err)
	}
	strContent := string(data)
	if len(strContent) > 4000 {
		strContent = strContent[:4000] + "\n... [content truncated]"
	}
	return ReadFileOutput{
		Content: strContent,
		Bytes:   len(data),
	}, nil
}

func ReadFileTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "read_file",
		Description: "Reads and returns the contents of a local file path.",
	}, readFileSnippet)
}
