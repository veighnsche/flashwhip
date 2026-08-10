# Issue #7 - Session and Command System Lacking Discoverability

**Project**: flashwhip  
**Module**: `pkg/ui/repl.go`, `cmd/sessions.go`  
**Priority**: 🟡 Major  
**Labels**: discoverability, UX, commands  

## Description

The slash command system exists (`/commands`, `/sessions`, `/clear`, etc.) but lacks **discoverability features**. New users stumble because they don't know what's available or how to use them.

## Current UX Friction Points

### 1. Zero Auto-Complete for Commands
```go
// In REPL.go prompt loop:
input := readCommand(...)   
switch input {
case "/commands": // Hard-coded only these values match...
    showCommands()
}
// No tab-completion, no fuzzy matching
```

User typing `/command` must know exact string; mistypes get silently ignored.

### 2. Static Help with No Contextual Depth  
The existing help system shows basic usage once — plain text, no formatting or categories.

### 3. Session System One-Dimensional  
- No filtering by time/context size metadata
- No interactive selection mode (must type session ID numerals)
- Can't delete old session history via UI

## Missing Features

| Feature | Users Expect | Current Gap |
|---------|--------------|-------------|
| Tab autocomplete for commands | `/comm` TAB → `/commands` | Complete nonexistence |
| Contextual help hierarchy | Grouped categories not walls of text | Flat single-message render only |
| Delete/filter sessions | Visually pick-and-delete without IDs | Hardcoded string-based lookup only |

## Proposed Solutions

### Solution 1: Fuzzy Finder Overlay on Prompt Line
```go
// In cmd/repl.go where readline is used currently:
rl.SetCompleter(func(line string) []string { return fuzzyMatchCommands(line) })

func fuzzyMatchCommands(input string) []string {
   // Return all slash-named commands starting with entered prefix
}
```

### Solution 2: Multi-Section Help System Redesign
Instead of flat one-page help, show contextual sections that match current context.

## Acceptance Criteria

- [ ] Tab completion works for all slash commands
- [ ] Command help broken into meaningful categories
- [ ] Interactive session selection mode available
- [ ] Delete sessions via UI confirmation dialog

---
*Severity: MEDIUM — usability gap for new and power users alike.*
