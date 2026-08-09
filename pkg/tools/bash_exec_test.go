package tools

import (
	"strings"
	"testing"
)

func TestExecCommandTool_Success(t *testing.T) {
	in := ExecCommandInput{
		Command: "echo 'Flashwhip Test'",
	}

	out, err := executeShellCommand(nil, in)
	if err != nil {
		t.Fatalf("executeShellCommand failed: %v", err)
	}

	if out.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", out.ExitCode)
	}

	if !strings.Contains(out.Stdout, "Flashwhip Test") {
		t.Errorf("Stdout = %q, want 'Flashwhip Test'", out.Stdout)
	}
}

func TestExecCommandTool_Error(t *testing.T) {
	in := ExecCommandInput{
		Command: "non_existent_command_12345",
	}

	out, err := executeShellCommand(nil, in)
	if err != nil {
		t.Fatalf("executeShellCommand failed: %v", err)
	}

	if out.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code for invalid command")
	}
}

func TestExecCommandTool_EmptyCommand(t *testing.T) {
	_, err := executeShellCommand(nil, ExecCommandInput{Command: ""})
	if err == nil {
		t.Errorf("Expected error for empty command, got nil")
	}
}
