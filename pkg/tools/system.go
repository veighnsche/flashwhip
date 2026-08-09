package tools

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

type SystemInfoInput struct{}

type SystemInfoOutput struct {
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	GoVersion string `json:"go_version"`
	Time      string `json:"current_time"`
}

func getSystemInfo(_ agent.Context, _ SystemInfoInput) (SystemInfoOutput, error) {
	return SystemInfoOutput{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		Time:      time.Now().Format(time.RFC3339),
	}, nil
}

func SystemInfoTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "get_system_info",
		Description: "Returns details about the host operating system, architecture, and current time.",
	}, getSystemInfo)
}

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
