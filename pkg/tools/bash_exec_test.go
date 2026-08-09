package tools

import (
	"os"
	"path/filepath"
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

func TestExecCommandTool_Cwd(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subfolder")
	_ = os.MkdirAll(subDir, 0755)

	in := ExecCommandInput{
		Command: "pwd",
		Cwd:     subDir,
	}

	out, err := executeShellCommand(nil, in)
	if err != nil {
		t.Fatalf("executeShellCommand with Cwd failed: %v", err)
	}

	if out.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", out.ExitCode)
	}

	if !strings.Contains(out.Stdout, "subfolder") {
		t.Errorf("Stdout = %q, want containing 'subfolder'", out.Stdout)
	}
}

func TestExecCommandTool_DestructivePattern(t *testing.T) {
	if !isDestructiveCommand("rm -rf /") {
		t.Errorf("Expected 'rm -rf /' to be marked as destructive")
	}

	if !isDestructiveCommand("git reset --hard HEAD~1") {
		t.Errorf("Expected 'git reset --hard' to be marked as destructive")
	}

	if isDestructiveCommand("go test ./...") {
		t.Errorf("Expected 'go test ./...' to NOT be marked as destructive")
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
