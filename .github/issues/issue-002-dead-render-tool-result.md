# Issue #2 - Dead Code / Broken RenderToolResult Function

**Project**: flashwhip  
**Module**: `pkg/ui/tool_renderer.go`  
**Priority**: 🔴 Critical  
**Labels**: bug, UI, tool rendering  

## Description

The function `RenderToolResult()` **always returns an empty string**, making it impossible to display any tool execution result output. This appears to be either stubbed out or accidentally deleted during refactoring.

## Current Code (pkg/ui/tool_renderer.go:251-256)

```go
// RenderToolResult and RenderToolCall wrappers for compatibility
func RenderToolResult(name string, response map[string]any) string {
    return ""  // 🔴 ALWAYS returns empty!
}
```

## Reproduction Steps

1. Execute any `flashwhip chat` prompt that triggers tool calls with string results (e.g., `read_file`, `file_search`)  
2. Observe the tool call metadata appears (`Executing tool: ...`) in output  
3. Notice **no result content** renders below it - the section is silent/invisible

## Affected Code Paths

- All places calling `RenderToolResult()` across the UI package return empty strings
- The function exists but does nothing - no side effects, just dead code

## Expected vs Actual

| Aspect | Expected | Actual |
|--------|----------|--------|
| Function behavior | Formats & returns tool result renderable string  
| Display in REPL | Tool output should appear after "Executing tool" metadata block
| Current state | Silent empty return value → no visible tool results |

## Impact Assessment

- **Severity: P0** - Tool outputs are completely invisible to users despite being computed correctly internally
- All text/string-based tool responses (file reads, search results, command output) become unreadable in REPL view
- Users cannot validate whether tools succeeded beyond the "Executing..." message itself
- Debugging conversation flows is impossible without seeing what tools returned

## Proposed Fix Options

### Option A: Implement Proper Rendering
```go
func RenderToolResult(name string, response map[string]any) string {
    if result, ok := response["result"].(string); ok && result != "" {
        // Format as code block or styled text depending on tool type
        return fmt.Sprintf("```\n%s\n```", result)
    } else if output, ok := response["output"]; ok {
        // Handle various output formats (json, error structs, etc.)  
        return formatFormattedOutput(output)
    }
    return "[no content]"
}
```

### Option B: Reuse Existing Tool Formatting Pipeline
Leverage `RenderCombinedToolExecution()` which already handles this correctly for batch tool calls.

## Acceptance Criteria

- [ ] All string-based tool outputs render to terminal (`read_file`, `exec_command` text output, etc.)
- [ ] JSON/tool structured results format with proper indentation
- [ ] Empty/error responses show appropriate placeholder message  
- [ ] No regression in existing test coverage for tool rendering paths

---

*Related Issue: #3 (Static Banner)*
