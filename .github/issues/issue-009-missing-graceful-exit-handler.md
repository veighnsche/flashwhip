# Issue #9 - Missing Ctrl+C Graceful Exit Handler in REPL

**Project**: flashwhip  
**Module**: `pkg/ui/repl.go`  
**Priority**: 🟡 Major  
**Labels**: robustness, UX polish  

## Description

There is no explicit handler for the interrupt signal (Ctrl+C / SIGINT) during active REPL interactions. This means:
- Long-running operations cannot be cleanly cancelled mid-stream
- Session state may get corrupted if interrupted at inopportune moments  
- Terminal cursor/line state can become permanently garbled leaving screen messy with partial renders and dangling characters still visible afterward

## Current State Analysis

```go
// In cmd/root.go main():
defer cancel()  // Handles high-level operation cancellation on exit...

// But inside REPL execution loop in pkg/ui/repl.go:
for {                                          
    input := readCommand(...)                          
    if err != nil {                                       
        // Currently only handles EOF (Ctrl+D / ^D)  
        // NO SIGINT handler mid-command processing!
        break                                              
    }
}
```

The `main()` function does register a defer/cancel for cleanup, but when users interrupt *during* active tool execution or response generation within the interactive loop itself — everything goes sideways because there's no direct signal catching mechanism designed specifically targeting that particular scenario.

## Impact Scenarios

1. **Mid-generation Ctrl+C**: Text partially rendered, cursor stuck mid-line
2. **Mid-tool-call interrupt**: Tool may be left running orphaned in background  
3. **Interrupt during context pruning**: Session state could become inconsistent  
4. **No visual "shutting down" message**: Abrupt exit leaves user confused

## Proposed Fix: Add Dedicated Signal Handling

```go
// In repl.go, add signal channel handling:
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT)

go func() {
    sig := <-sigChan
    fmt.Printf("\n\nInterrupted. Cleaning up...\n")
    tracker.Cancel()   // Cancel any in-progress operations  
    cleanupSession()   // Save partial state if needed
    os.Exit(130)       // Standard unix exit code for SIGINT
}()

// In the main loop, check context cancellation:
select {
case <-ctx.Done():
    return ctx.Err()
default:
    // continue normally
}
```

## Acceptance Criteria

- [ ] Clean Ctrl+C handling during token generation (no garbled terminal)  
- [ ] Optional: Message displayed "Interrupted. Cleaning up..." before exit
- [ ] Optional: Partial session state saved if interrupted mid-turn
- [ ] Orphaned tool processes cleaned up on interrupt  

---
*Severity: Low-to-Medium — cosmetic/user experience issue, but impacts user trust in application robustness during interactive sessions.*
