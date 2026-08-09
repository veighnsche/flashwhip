package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

type ExecCommandInput struct {
	Command string `json:"command" jsonschema:"The shell command line to execute (e.g. 'go test ./...', 'git status')"`
	Cwd     string `json:"cwd,omitempty" jsonschema:"Optional relative or absolute working directory path to execute the command in"`
}

type ExecCommandOutput struct {
	Command       string `json:"command"`
	Cwd           string `json:"cwd,omitempty"`
	Stdout        string `json:"stdout"`
	Stderr        string `json:"stderr"`
	ExitCode      int    `json:"exit_code"`
	DurationMs    int64  `json:"duration_ms"`
	IsDestructive bool   `json:"is_destructive,omitempty"`
}

func isDestructiveCommand(cmd string) bool {
	lower := strings.ToLower(cmd)
	dangerousPatterns := []string{
		"rm -rf /", "rm -r /", "git reset --hard", "git clean -fdx",
		"drop database", "drop table", "mkfs", "dd if=", "> /dev/sd",
	}
	for _, p := range dangerousPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func executeShellCommand(_ agent.Context, in ExecCommandInput) (ExecCommandOutput, error) {
	cmdStr := strings.TrimSpace(in.Command)
	if cmdStr == "" {
		return ExecCommandOutput{}, fmt.Errorf("command cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	startTime := time.Now()
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	if strings.TrimSpace(in.Cwd) != "" {
		cmd.Dir = strings.TrimSpace(in.Cwd)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	durationMs := time.Since(startTime).Milliseconds()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	stdout := strings.TrimSpace(stdoutBuf.String())
	stderr := strings.TrimSpace(stderrBuf.String())

	if len(stdout) > 3000 {
		stdout = stdout[:3000] + "\n... [stdout truncated for token safety]"
	}
	if len(stderr) > 1000 {
		stderr = stderr[:1000] + "\n... [stderr truncated for token safety]"
	}

	return ExecCommandOutput{
		Command:       cmdStr,
		Cwd:           in.Cwd,
		Stdout:        stdout,
		Stderr:        stderr,
		ExitCode:      exitCode,
		DurationMs:    durationMs,
		IsDestructive: isDestructiveCommand(cmdStr),
	}, nil
}

func ExecCommandTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "exec_command",
		Description: "Executes a shell command line (e.g. go test, git diff, npm test) in an optional working directory (cwd) and returns stdout/stderr.",
	}, executeShellCommand)
}
