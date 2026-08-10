package ui

import (
	"strings"
	"testing"
)

func TestShortenPath(t *testing.T) {
	cwd := "/Users/vince/Projects/flashwhip"

	if got := shortenPath("/Users/vince/Projects/flashwhip/cmd", cwd); got != "cmd" {
		t.Errorf("shortenPath() = %q, want %q", got, "cmd")
	}
	if got := shortenPath("/Users/vince/Projects/flashwhip/go.mod", cwd); got != "go.mod" {
		t.Errorf("shortenPath() = %q, want %q", got, "go.mod")
	}
}

func TestRenderCombinedToolExecution(t *testing.T) {
	cwd := "/Users/vince/Projects/flashwhip"

	t.Run("list_dir single line", func(t *testing.T) {
		args := map[string]any{"dir_path": "/Users/vince/Projects/flashwhip/cmd"}
		resp := map[string]any{"status": "success"}

		out := RenderCombinedToolExecution("list_dir", args, resp, cwd)
		if !strings.Contains(out, "📁") || !strings.Contains(out, "Listed: cmd/") {
			t.Errorf("RenderCombinedToolExecution list_dir = %q", out)
		}
		if strings.Contains(out, "dir_path=") {
			t.Errorf("RenderCombinedToolExecution list_dir should not contain raw dir_path=, got %q", out)
		}
	})

	t.Run("read_file single line", func(t *testing.T) {
		args := map[string]any{"file_path": "/Users/vince/Projects/flashwhip/go.mod"}
		resp := map[string]any{"total_lines": 88, "bytes": 4096}

		out := RenderCombinedToolExecution("read_file", args, resp, cwd)
		if !strings.Contains(out, "📄") || !strings.Contains(out, "Read: go.mod (88 lines)") {
			t.Errorf("RenderCombinedToolExecution read_file = %q", out)
		}
	})

	t.Run("read_file line range slice with total lines", func(t *testing.T) {
		args := map[string]any{"file_path": "/Users/vince/Projects/flashwhip/pkg/tools/web_fetch.go", "start_line": 140, "end_line": 165}
		resp := map[string]any{"total_lines": 213, "start_line": 140, "end_line": 165}

		out := RenderCombinedToolExecution("read_file", args, resp, cwd)
		expected := "Read: pkg/tools/web_fetch.go (L140-165, 26 of 213 lines)"
		if !strings.Contains(out, expected) {
			t.Errorf("RenderCombinedToolExecution read_file range = %q, want substring %q", out, expected)
		}
	})

	t.Run("read_file line range slice without total lines", func(t *testing.T) {
		args := map[string]any{"file_path": "/Users/vince/Projects/flashwhip/test.txt", "start_line": 10, "end_line": 20}
		resp := map[string]any{"start_line": 10, "end_line": 20}

		out := RenderCombinedToolExecution("read_file", args, resp, cwd)
		expected := "Read: test.txt (L10-20, 11 lines)"
		if !strings.Contains(out, expected) {
			t.Errorf("RenderCombinedToolExecution read_file range = %q, want substring %q", out, expected)
		}
	})

	t.Run("exec_command success single line", func(t *testing.T) {
		args := map[string]any{"command": "go test ./..."}
		resp := map[string]any{"exit_code": 0, "duration_ms": 400}

		out := RenderCombinedToolExecution("exec_command", args, resp, cwd)
		if !strings.Contains(out, "🛠️") || !strings.Contains(out, "Executed: \"go test ./...\" (400ms)") {
			t.Errorf("RenderCombinedToolExecution exec_command = %q", out)
		}
	})

	t.Run("exec_command error single line", func(t *testing.T) {
		args := map[string]any{"command": "go build ./..."}
		resp := map[string]any{"exit_code": 1, "stderr": "syntax error"}

		out := RenderCombinedToolExecution("exec_command", args, resp, cwd)
		if !strings.Contains(out, "✖") || !strings.Contains(out, "failed (exit 1)") {
			t.Errorf("RenderCombinedToolExecution exec_command error = %q", out)
		}
	})
}
