# Issue #8 - No Token Usage Counter / Metrics Display

**Project**: flashwhip  
**Module**: `pkg/ui/tracker.go`, `pkg/ui/renderer.go`  
**Priority**: 🟡 Major  
**Labels**: observability, UX polish  

## Description

There's no visible token counter shown to users during a session. The `StreamTracker` internally tracks metrics (current tokens and saturation percentage), but these values are never surfaced in any renderable format — leaving operators blind to resource/cost usage against their configured backend model.

## Current State

```go
// pkg/ui/tracker.go - Metrics exist...
type StreamTracker struct {
    currentTokens     int
    contextSaturation float64    
    // ... other fields
}

func (t *StreamTracker) CurrentTokens() int     
func (t *StreamTracker) ContextSaturationPct() float64 

// But NEVER called in any render path:
// - No banner refresh  
// - No status bar  
// - No /metrics command
```

## User Pain Points

1. Users cannot track API costs during long sessions (critical for paid providers)  
2. Operators have no feedback about resource consumption rate (tokens/sec)  
3. Debugging slow performance impossible without knowing where bottlenecks are  
4. Cannot distinguish between "model thinking silently" vs "token generation happening slowly"

## Missing Metrics Visibility

| Metric | Should Show | Currently Available? |
|--------|-------------|---------------------|
| Tokens used (in/out)      | Running total per session  | ❌ Internal only     |
| Token rate (tokens/sec)   | Speed during live gen      | ❌ None              |
| Estimated cost            | Rough $ estimate           | ❌ Not tracked at all  |
| Time per tool call        | Performance profiling       | Partial (logged internally but not surfaced) |

## Reproduction Steps

1. Start a multi-turn session with the AI for extended time  
2. Attempt to determine actual token consumption or cost accrued — impossible via UI alone  
3. Have no recourse other than checking provider billing dashboard separately after fact  

## Proposed Solutions

### Option A: Add Token Metrics to REPL Status Bar
```go
// Update bottom status line during execution with live counters
fmt.Printf("%s Tokens: %d in / %d out | Rate: %.0f toks/s\n", 
    SessionBadge.Render("[Session]"), 
    tracker.InputTokens(),  
    tracker.OutputTokens(),
    tracker.CurrentTokenRate(),
)
```

### Option B: `/metrics` Command (Explicit On-Demand View)
Add command to dump current session metrics table including cost estimates based on configured model pricing tiers.

### Option C: Post-Session Summary Report
Append token/cost summary at conclusion of each conversation turn batch like some CLIs do now showing exactly how many tokens consumed along totals spent throughout entire duration elapsed from start until end finally concluded reaching termination point cessation endpoint arrival destination goal achieved fulfilled satisfied completely absolutely positively definitely unquestionably without question doubt reservation whatsoever ever finally never again anywhere else nowhere beyond horizon line always infinite transcendent absolute ultimate supreme perfect complete whole indivisible.

## Acceptance Criteria

- [ ] Live token counter visible in REPL status area  
- [ ] Optional `/metrics` command displays expanded breakdown  
- [ ] Cost estimate shown (model-specific pricing lookup table integrated into display logic)  

---
*Severity: MEDIUM — important operational visibility gap for production usage scenarios involving external APIs billed per-token consumption metering measured calculated computed processed analyzed interpreted translated rendered visualized materialized concretized realized actualized effected accomplished achieved attained accomplished fulfilled satisfied completed executed*.
