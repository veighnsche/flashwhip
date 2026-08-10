# Issue #4 - Static Banner Never Updates After Startup

**Project**: flashwhip  
**Module**: `pkg/ui/banner.go`, `cmd/root.go`
**Priority**: 🟡 Major  
**Labels**: UI, state consistency, banner
  
## Description

The startup banner displays configuration values (model name, context size) **only once at application launch**. These become stale if:
- User changes model/session mid-campaign with flags `-m`, `--url`  
- Context saturation metrics evolve over time  
- Configuration updates dynamically  

This creates a misleading UI that presents outdated info as current state.

## Current Code Location

```go
// cmd/root.go (main execution path) - banner rendered ONCE at startup
banner := RenderBanner(cfg.Model, ctxSize, cwdPath, branchInfo, maxTurns)   
fmt.Println(banner)  // 🔴 One-time render - never refreshed!
```

## Evidence of Staleness Problem

1. Users pass `--model different-model` or `--url new-ollama-url` on CLI
2. Banner shows original cold-start values from first execution  
3. No banner refresh occurs to reflect actual current config/state  
4. Similar pattern everywhere: config is read fresh BUT display happens upfront only

## Missing Dynamic Updates

| Element | Initial Value | Should Update When... | Current State |
|---------|---------------|-----------------------|---------------|
| Model name   | Cold-start model  | User changes via flag/session switch | ❌ Stale forever   |
| Context size | Static ctxSize    | Session turns accumulate, saturation grows  | ❌ Fixed number    |
| CWD path     | Startup working dir | User runs `cd` in terminal (if supported)  | ❌ Never updates  |
| Branch info  | First git hash   | New commits added to repo                   | ❌ Out of date    |

## Impact Severity

- **Low frequency but high confusion**: Users don't notice immediately, get suspicious when things behave differently  
- **Trust degradation**: UI shows values that "should be" current are actually wrong
- **Debugging impediment**: User reports model not working - they're looking at outdated banner model name

## Proposed Solutions

### Option A: Session-Level Banner Refresh (Recommended)
Refresh banner display periodically AND on state transitions:
```go
func renderDynamicBanner(ctx context.Context, tracker *StreamTracker, cfg *config.Config, cwd string) {
    // Called each session start + at saturation warnings
    fmt.Println(renderFreshBanner(cfg.Model, tracker.CurrentTokens()/1000, cwd))  
}

// In the REPL loop - update on key events:
case ... when model switches or context thresholds hit:
    renderDynamicBanner(...)
```

### Option B: Status Bar Approach (Better UX)
Move away from big banner to compact header that updates inline:
```go
fmt.Printf("🚀 Flashwhip | Model: %s | Context: %dk tokens | PID: %d\n", 
    cfg.Model, tracker.CurrentTokens()/1000, os.Getpid())
// Update this line at regular intervals instead
```

## Acceptance Criteria

- [ ] Banner/replacement header displays current model throughout session  
- [ ] Saturation context token count updates live as turns process  
- [ ] User-visible state changes (model switch via flags) reflect in display  
- [ ] No visual tear or re-rendering artifacts during any updates  

---
*Severity: HIGH — user trust in UI accuracy*
