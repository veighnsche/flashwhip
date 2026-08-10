# Issue #3 - Context Saturation Completely Invisible to Users

**Project**: flashwhip  
**Module**: `pkg/ui/tracker.go`, `pkg/config`  
**Priority**: 🟡 Major  
**Labels**: state management, error handling, UX  

## Description

The tracker maintains context saturation tracking (`tracker.ContextSaturationPct()`) BUT **never surfaces this information to the user anywhere in the UI**. Users have no idea when they're approaching limits or why pruning silently occurs mid-conversation.

## Current Code Evidence

```go
// pkg/ui/tracker.go — TRACKING exists...
func (t *StreamTracker) SetSaturation(pct float64) { ... }
func (t *StreamTracker) ContextSaturationPct() float64 { return t.contextSaturation }
func (t *StreamTracker) CurrentTokens() int           { return t.currentTokens }

// ...but UI NEVER CALLS THESE to display anything!
```

**Proof the UI ignores it:**
- `RenderCombinedToolExecution` does NOT check saturation  
- REPL prompt output doesn't warn about context near limit
- No warning banner when thresholds breach (*75%, 90%, 100%*)

## Reproduction Scenario

1. Start long conversation with multiple turns involving large files/tools 
2. Trigger session pruning at `MaxContextTurns` threshold (happens automatically)
3. See no prior indication in UI that limits were hit — just context disappearing silently
4. Confusion when assistant seems to have "forgotten" earlier parts

## Current Missing Indicators

| Saturation Level | Visible Warning? | Expected Behavior |
|------------------|------------------|-------------------|
| 0-60% | None (correct)  
| 75% | ❌ Silent threshold hit point  
| 90%+ | ❌ No danger indicator before auto-prune   |
| At `MaxContextTurns` | ❌ Hidden from user entirely  

## Impact

Users cannot:
- Anticipate when to clear context manually (`/clear`)
- Distinguish intentional pruning vs. bugs  
- Understand why assistant "forgot" earlier instructions mid-session
- Make informed decisions about session boundaries

## Proposed Solutions

### 1. Live Saturation Counter (Recommended)
Display current saturation in the **REPL prompt area**, update during tool loops:
```go
// In REPL prompt string or bottom status bar  
fmt.Printf("%s Context: %d/%d turns (%d%%)\n", 
    ConfigBadge.Render("[Config]"),
    tracker.CurrentTurnCount(),    
    tracker.MaxContextTurns(),
    int(tracker.ContextSaturationPct()),
)
```

### 2. Warning Banners (Threshold Triggers)
Show transient notifications at key saturation points:
- **75%**: "⚠️ Context reaching capacity"  
- **90%**: "🔴 Nearly full context — consider /clear"
- Post-prune: "ℹ️ Pruned conversation history to free context"

### 3. Session Info in Banner Update
Refresh banner values periodically (not just startup):
```go
func RefreshBannerDisplay(tracker *StreamTracker, cfg *config.Config) {
    // Instead of one-time print at startup: only show live updates
    fmt.Printf("Context usage: %d tokens / saturation %.1f%%\n", 
        tracker.CurrentTokens(), tracker.ContextSaturationPct())
}
```

## Acceptance Criteria

- [ ] Saturation displayed as percentage near session context limits  
- [ ] Progress bar or counter visible alongside current turn count in REPL interface  
- [ ] Clear textual warning at 75%+ saturation thresholds  
- [ ] Post-prune notification confirms what occurred (user was not blind to it)
- [ ] `/status` command shows full context metrics on demand

---

*Severity: HIGH — breaks user trust through silent state transitions*  
*Related: #4 Static Banner Never Updates*

