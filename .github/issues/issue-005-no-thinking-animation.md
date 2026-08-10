# Issue #5 - No Animated Thinking or Processing Indicator

**Project**: flashwhip  
**Module**: `pkg/ui/runner.go`, `pkg/ui/tracker.go`
**Priority**: 🟡 Major  
**Labels**: UX, animation, user feedback  

## Description

When tool calls execute and processing occurs:
1. Only shows static text `🧠 [Thinking]` - single line  
2. Then blank space for the entire duration of processing  
3. No visual indication that work is actively happening beyond this one marker

Users get no sense of progress, speed, or "this app isn't frozen" during:
- Multi-step tool chains (sequential file reads + edit cycles)
- Web fetch operations waiting on network I/O  
- Large exec_command executions taking seconds+  

## Current Code State

```go
// pkg/ui/runner.go - thinking display:
fmt.Printf("🧠 [Thinking]\n")   // 🔴 Static line, never changes!
// Then executes tool loops silently...
for _, tc := range toolCalls {
    executeAndRender(tc)  // No feedback while executing each one
}
```

**Actual timing (observed in logs):**
- Tool loop processes → displays nothing for hundreds of milliseconds between each call
- Total time: multiple seconds with ZERO user-visible activity beyond first "Thinking" message  

## Missing Animation States

| State | What Should Happen | Current Behavior |
|-------|--------------------|------------------|
| Waiting on tool I/O         | Spinner/cursor blink while exec waits    
| Executing multi-step plan   | Sequential progress indicator showing step 1/N, then 2/N... 
| Network fetch active        | Pulse animation while waiting for response  
| Large computation pending   | Progress bar or countdown  

## Impact

- **Perception of frozen app**: Users see static "Thinking" and blank screen - assumes hung
- **Anxiety from uncertainty**: Don't know if working hard vs crashed
- **First-time user confusion**: New users may kill process mid-think expecting nothing happens  
- **Differentiation loss**: All competitive CLIs show active processing animation; flashwhip doesn't

## Proposed Solutions

### Option A: Spinner with Status Messages (Recommended MVP)

Reuse `pkg/ui/tracker.go` spinner infrastructure already partially implemented elsewhere:
```go
func startThinkingSpinner(ctx context.Context, status string) (*strings.Builder, io.Writer) {
    // Returns live-updating output that can be cleared/showing ...etc  
    go spinWithStatus(func(frame int) {
        switch {
        case frame % 4 == 0:   print("⠋")
        case frame % 4 == 1:   print("⠙") 
        case frame % 4 == 2:   print("⠹")  
        default:               print("⠸")
        }
    }, status)
}
```

### Option B: Typewriter Cursor Under Response Header  

Add blinking block cursor indicator at the thinking line that pulses while processing:
```go
fmt.Printf("🧠 [Thinking] ░\r")  // After each ms tick updates to ▓▒░ etc.  
animateCursorUntilResponseComplete(ctx) 
```

### Option C: Tool Progress Bar (When Known Steps)

For multi-step plans that have N expected tools:
```go
// Track tool calls and show progress as we go through them
for i, tc := range allToolCalls {
    fmt.Printf("  %d/%d: %s\n", i+1, len(allToolCalls), tc.Name)
    executeTool(tc)  // With mini-spinner after each step...
}
```

## Implementation Notes

- Must use ANSI escape codes for cursor movement (`\r`) without breaking terminal scrollback  
- Spinner needs cancellation hook if user hits Ctrl-C mid-processing to avoid orphan animation 
- Should auto-hide/clear once actual response rendering begins, avoiding overlap issues

## Acceptance Criteria

- [ ] Visual feedback during any tool execution/waiting phase (>200ms duration)    
- [ ] Cursor or spinner indicates active processing (not frozen/idle state)  
- [ ] Smooth cancellation on Ctrl+C without leaving zombie animation artifacts
- [ ] Status message updates to reflect current phase ("Reading file...", "Editing...", etc.)  

---
*Severity: MEDIUM - major UX polish but functional; still provides wrong signal about app health*