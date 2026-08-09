package tools

import (
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
