# Issue #6 - Error Feedback is Invisible and Undifferentiated

**Project**: flashwhip  
**Module**: `pkg/ui/renderer.go`, `pkg/ui/tool_renderer.go`    
**Priority**: 🔴 Critical (usability) / 🟡 Major (diagnosability)  
**Labels**: debugging, error handling, user experience

## Description

When tools fail or produce errors:
1. Only the tool call metadata shows ("Executing tool:", exit code 0/1 etc.)
2. No **color differentiation** between successful vs failed tool calls in terminal view   
3. Tool error messages are not surfaced/connected to parent response rendering  
4. REPL session has no persistent record or banner surface of errors that occurred   
5. There's zero visual distinction distinguishing "tool succeeded but output empty" from "tool actually errored out with something meaningful"

## Current Code Path - Silent Failures

```go
// In runner.go tool execution loop (simplified):
result := executeTool(tc)
fmt.Printf("↳ %s: code=%d\n", tc.Name, result.ExitCode)  // 🔴 No error highlighting!
if result.ExitCode != 0 { 
    // Error was silently swallowed here currently - nothing special happens!
}
```

**Evidence of dead/error handling:** The errors package (`/pkg/errors`) has proper codes mapped but UI doesn't use them for visual feedback. Only `exit code` metadata visible, not actual failure reason formatted nicely.

## Missing Error Visibility States

| Scenario | Should Show | Actually Shows | Gap |
|----------|-------------|----------------|-----|
| Exec command exits nonzero (e.g., file not found) | Red highlight with error message    | Silent exit=1 only   ❌ No context  |
| ReadFile missing path     | Clear "file does not exist" formatting | Empty/silent failure output 
| WebFetch blocked          | Timeout/network failure visual banner    | Just empty text/nothing    
| Agent tool returns err map | Error message extracted and displayed prominently  | Raw `map` rendered badly/ignored  

## Reproduction Steps in Real Session

1. Run: `/flashwhip chat "Read the non-existent file /does-not-exist.txt"`
2. Observable **nothing visually signals error** - user thinks it succeeded  
3. Tool returns but UI never surfaces actual reason why output is empty  
4. Confusion ensues for hours debugging when actually just error being hidden

## Specific Current Code Issues

### 1. Missing Red Highlighting on Failure Tool Calls
```go
// Currently always uses neutral RenderSuccess() 
fmt.Printf("%s: code=%d\n", tc.Name, result.ExitCode)  
// Should check and use different renderer for failures!
```

### 2. No Error Metadata Displayed  
Tool errors containing meaningful message strings ARE available from `pkg/errors` but invisible to REPL view - nothing extracts them from the error map into terminal output.

### 3. Silent Context Pruning Mistaken for Error Behavior 
When context hits MaxContextTurns, pruning silently occurs AND user can't tell if their request was **processed successfully with forgotten history** vs **blocked due to failure**. Zero notification differentiates these states.

## User Experience Impact Ranking

```
User sees blank response → "Did it work?"
                         → asks AI again "Why didn't that file read display"
                         → gets re-explained when actually errored silently  
                         → frustration, lost trust in tool reliability
```

## Recommended Fixes (Priority Ordered)

### Fix 1: Color-Differentiate Tool States Immediately (P0 Fix)
Add red error styling for nonzero exit codes and error strings from tools:
```go
if result.ExitCode != 0 || containsError(result.Metadata) {
    fmt.Printf("%s [FAILED]: code=%d", ConfigBadge.Render(tc.Name), result.ExitCode)  
    renderToolErrors(result.Errors, config.GetRedColor())  // New helper function to display 
    continue  // don't just silently proceed...
}

func RenderToolErrors(errors []string, colorStyle lipgloss.Style) string {
   return fmt.Sprintf("\n  %s\n", strings.Join(errors, "\n  "))
}
```

### Fix 2: Surface Context Warning Clearly on Prune/Max Turns
When auto-pruning occurs mid-turn (at `tracker.MaxContextTurns()`):
- Show large banner "⚠️ CONTEXT PRUNED — previous messages dropped"  
- Don't silently continue conversation appearance, acknowledge explicitly  

### Fix 3: Centralized Error Status Rendering for Consistency

**Extract error display into shared function:** 
```go
func RenderToolOutcome(tc *genai.ToolCall, result interface{}) {
   // Unified formatter applies style based on success/failure/etc.
}
```

## Acceptance Criteria Checklist

- [ ] Tool failure/exit != 0 renders prominently with red highlighting (clear visual failure)  
- [ ] Error message metadata actually displays in terminal output when present 
- [ ] Distinguish between "ran successfully, empty result" vs "[ERROR] failed because..." - BOTH must surface visually  
- [ ] Context pruning explicitly acknowledged to user via noticeable notification text/banner  
- [ ] New errors_cmd.go `/errors` subcommand surfaces runtime errors encountered during session (debug use) without relying on silent terminal state tracking

---
*Severity note: This is P0 for usability but has no crash/security implications - bugs remain discoverable through investigation at least.* 

