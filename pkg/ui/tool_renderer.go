package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// getToolIcon returns a subtle icon and color for a tool.
func getToolIcon(name string) (string, lipgloss.Color) {
	switch strings.ToLower(name) {
	case "exec_command":
		return "🛠️", ToolShellColor
	case "read_file":
		return "📄", ToolFileColor
	case "write_file":
		return "📝", ToolFileColor
	case "edit_file":
		return "✏️", ToolFileColor
	case "grep_search", "file_search":
		return "🔍", ToolSearchColor
	case "web_search", "web_fetch":
		return "🌐", ToolWebColor
	case "git_diff":
		return "🔀", ToolGitColor
	case "list_dir":
		return "📁", ToolDefaultColor
	case "sys_info", "system_info":
		return "ℹ️", ToolDefaultColor
	default:
		return "⚡", ToolDefaultColor
	}
}

// shortenPath strips project root prefix or home directory to show clean relative paths.
func shortenPath(pathStr string, cwd string) string {
	pathStr = strings.TrimSpace(pathStr)
	if pathStr == "" {
		return ""
	}

	if cwd != "" {
		if rel, err := filepath.Rel(cwd, pathStr); err == nil && !strings.HasPrefix(rel, "..") {
			if rel == "." {
				return "./"
			}
			return rel
		}
	}

	return pathStr
}

// RenderCombinedToolExecution formats a completed tool call & response into a single, ultra-clean line.
// Example: "  📄 Read: go.mod (88 lines)"
// Example: "  📁 Listed: cmd/"
// Example: "  🛠️ Executed: go test ./... (0.4s)"
func RenderCombinedToolExecution(name string, args map[string]any, response map[string]any, cwd string) string {
	icon, color := getToolIcon(name)
	bullet := lipgloss.NewStyle().Foreground(color).Render(icon)

	isError := checkIsError(response)
	statusIcon := ToolOutcomeSuccessStyle.Render("✔")
	if isError {
		statusIcon = ToolOutcomeErrorStyle.Render("✖")
	}

	summary := formatActionAndOutcome(name, args, response, cwd, isError)

	return fmt.Sprintf("  %s %s %s", bullet, statusIcon, summary)
}

func checkIsError(response map[string]any) bool {
	if len(response) == 0 {
		return false
	}
	if exitCode, ok := response["exit_code"]; ok {
		if ec, isInt := toInt(exitCode); isInt && ec != 0 {
			return true
		}
	}
	if errVal, ok := response["error"]; ok && errVal != nil && fmt.Sprintf("%v", errVal) != "" {
		return true
	}
	if status, ok := response["status"]; ok && strings.ToLower(fmt.Sprintf("%v", status)) == "error" {
		return true
	}
	return false
}

func formatActionAndOutcome(name string, args map[string]any, response map[string]any, cwd string, isError bool) string {
	switch strings.ToLower(name) {
	case "exec_command":
		cmdStr := getArgOrRespString(args, response, "command")
		if cmdStr == "" {
			cmdStr = "command"
		}
		if len(cmdStr) > 60 {
			cmdStr = cmdStr[:57] + "..."
		}
		action := fmt.Sprintf("Executed: %s", ToolNameStyle.Render(fmt.Sprintf("%q", cmdStr)))

		if isError {
			errMsg := "failed"
			if exitCode, ok := response["exit_code"]; ok {
				errMsg = fmt.Sprintf("failed (exit %v)", exitCode)
			}
			return fmt.Sprintf("%s (%s)", action, ToolOutcomeErrorStyle.Render(errMsg))
		}

		var meta []string
		if dur, ok := response["duration_ms"]; ok {
			if ms, ok := toInt(dur); ok {
				if ms >= 1000 {
					meta = append(meta, fmt.Sprintf("%.1fs", float64(ms)/1000.0))
				} else {
					meta = append(meta, fmt.Sprintf("%dms", ms))
				}
			}
		}
		if len(meta) > 0 {
			return fmt.Sprintf("%s (%s)", action, ToolMutedStyle.Render(strings.Join(meta, ", ")))
		}
		return action

	case "read_file":
		filePath := getArgOrRespString(args, response, "file_path", "path")
		if filePath == "" {
			filePath = "(file)"
		}
		path := shortenPath(filePath, cwd)
		action := fmt.Sprintf("Read: %s", ToolNameStyle.Render(path))

		var meta []string
		start, hasStart := toInt(getArgOrRespVal(args, response, "start_line"))
		end, hasEnd := toInt(getArgOrRespVal(args, response, "end_line"))
		totalLines, hasTotal := toInt(getArgOrRespVal(args, response, "total_lines"))

		if hasStart && hasEnd && start > 0 && end >= start {
			returnedLines := end - start + 1
			if hasTotal && totalLines > 0 && returnedLines != totalLines {
				meta = append(meta, fmt.Sprintf("L%d-%d, %d of %d lines", start, end, returnedLines, totalLines))
			} else {
				meta = append(meta, fmt.Sprintf("L%d-%d, %d lines", start, end, returnedLines))
			}
		} else if hasTotal && totalLines > 0 {
			meta = append(meta, fmt.Sprintf("%d lines", totalLines))
		}

		if len(meta) > 0 {
			return fmt.Sprintf("%s (%s)", action, ToolMutedStyle.Render(strings.Join(meta, ", ")))
		}
		return action

	case "write_file":
		filePath := getArgOrRespString(args, response, "file_path", "path")
		path := shortenPath(filePath, cwd)
		return fmt.Sprintf("Wrote: %s", ToolNameStyle.Render(path))

	case "edit_file":
		filePath := getArgOrRespString(args, response, "file_path", "path")
		path := shortenPath(filePath, cwd)
		return fmt.Sprintf("Edited: %s", ToolNameStyle.Render(path))

	case "list_dir":
		dirPath := getArgOrRespString(args, response, "dir_path", "path")
		path := shortenPath(dirPath, cwd)
		if path == "" {
			path = "./"
		}
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		return fmt.Sprintf("Listed: %s", ToolNameStyle.Render(path))

	case "grep_search":
		query := getArgOrRespString(args, response, "query", "pattern")
		action := fmt.Sprintf("Searched: %s", ToolNameStyle.Render(fmt.Sprintf("%q", query)))
		if matches, ok := response["total_matches"]; ok {
			return fmt.Sprintf("%s (%s)", action, ToolMutedStyle.Render(fmt.Sprintf("%v matches", matches)))
		}
		return action

	case "file_search":
		pattern := getArgOrRespString(args, response, "pattern", "query")
		action := fmt.Sprintf("Searched files: %s", ToolNameStyle.Render(fmt.Sprintf("%q", pattern)))
		if count, ok := response["count"]; ok {
			return fmt.Sprintf("%s (%s)", action, ToolMutedStyle.Render(fmt.Sprintf("%v files", count)))
		}
		return action

	case "web_search":
		query := getArgOrRespString(args, response, "query")
		return fmt.Sprintf("Searched web: %s", ToolNameStyle.Render(fmt.Sprintf("%q", query)))

	case "web_fetch":
		url := getArgOrRespString(args, response, "url")
		return fmt.Sprintf("Fetched web: %s", ToolNameStyle.Render(url))

	case "git_diff":
		return fmt.Sprintf("Ran: %s", ToolNameStyle.Render("git diff"))

	default:
		target := getArgOrRespString(args, response, "command", "file_path", "path", "query", "url")
		if target != "" {
			return fmt.Sprintf("%s: %s", ToolNameStyle.Render(name), ToolTargetStyle.Render(target))
		}
		return ToolNameStyle.Render(name)
	}
}

func getArgOrRespString(args map[string]any, response map[string]any, keys ...string) string {
	if s := getArgString(args, keys...); s != "" {
		return s
	}
	if s := getArgString(response, keys...); s != "" {
		return s
	}
	return ""
}

func getArgOrRespVal(args map[string]any, response map[string]any, key string) any {
	if args != nil {
		if v, ok := args[key]; ok && v != nil {
			return v
		}
	}
	if response != nil {
		if v, ok := response[key]; ok && v != nil {
			return v
		}
	}
	return nil
}

func getArgString(args map[string]any, keys ...string) string {
	if len(args) == 0 {
		return ""
	}
	for _, k := range keys {
		if v, ok := args[k]; ok && v != nil {
			vStr := strings.TrimSpace(fmt.Sprintf("%v", v))
			if vStr != "" {
				return vStr
			}
		}
	}
	return ""
}

// RenderToolCall and RenderToolResult wrappers for compatibility
func RenderToolCall(name string, args map[string]any) string {
	return RenderCombinedToolExecution(name, args, nil, "")
}

func RenderToolResult(name string, response map[string]any) string {
	return ""
}

func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	default:
		var i int
		_, err := fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &i)
		return i, err == nil
	}
}
