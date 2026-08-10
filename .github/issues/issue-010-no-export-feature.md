# Issue #10 - No Export Conversations Feature

**Project**: flashwhip  
**Module**: `cmd/`, `pkg/db/`
**Priority**: 🟢 Nice-to-have  
**Labels**: feature request, UX enhancement  

## Description

There is no ability to export or share conversation sessions with users. Once a session ends or the terminal is closed, all conversational context is either:
- Lost (if not persisted)  
- Trapped in SQLite database inaccessible for human consumption  
- Impossible to reproduce/review without rerunning exact same prompts  

This prevents users from:
- Sharing useful AI-generated content with colleagues  
- Archiving valuable conversations for reference later  
- Using flashwhip outputs in documentation/code comments  
- Reproducing debugging sessions externally

## Current State

```go
// pkg/db/db.go - Sessions are persisted...
func (db *Database) SaveSession(sessionID string, messages []genai.Content) error
func (db *Database) GetSession(sessionID string) (*Session, error)

// ...but no export/render functionality exists anywhere in codebase.
```

## User-Requested Capabilities

| Export Format | Use Case | Priority |
|---------------|----------|----------|
| Markdown (.md)| Documentation, sharing on GitHub | **High**  
| JSON (.json)  | Programmatic processing, API integration    | Medium  
| Plain text    | Quick copy-paste into other tools           | Low    

## Reproduction Steps

1. Have a valuable conversation with flashwhip about something useful
2. Want to share that conversation or use parts of it elsewhere  
3. Realize there's no `/export` command available  

**Current workaround:** None — must manually transcribe from terminal scrollback (if still accessible)

## Proposed Solutions

### Option A: `/export session_id --format markdown` Command
Add new CLI subcommands to export sessions in various formats:
```bash
flashwhip export <session_id> [--format md|json|txt] [--output path/to/file]
```

This would pull from the SQLite database and render conversations into human-readable formats.

### Option B: Inline Export During Session
Add ability to copy/paste current conversation to clipboard or write file directly while active:
```go
// In commands.go handler for /export  
case "/export":
    // Show menu or accept path argument
    exportCurrentSessionToMarkdown(filepath)
```

### Option C: Dedicated Export Command With Session Selection UI
Interactive mode where user can browse and select session(s) to export via the REPl interface using the existing session management infrastructure already present.

## Implementation Considerations

- Must handle multi-role conversations properly (user vs assistant alternating)
- Should preserve tool call metadata when possible for full fidelity  
- Exported markdown should be nicely formatted with syntax highlighting cues where applicable
- Support appending to existing files vs overwriting (with confirmation dialogs)

## Acceptance Criteria

- [ ] `/export session_id` command available in REPL
- [ ] Markdown format export works and renders cleanly  
- [ ] JSON export option preserves full structure including tool calls/metadata
- [ ] Multiple sessions can be concatenated into single export file
- [ ] Exported files save to user-specified or default location with helpful status message confirmation

---
*Severity: Low — feature gap, not a bug but significantly limits usability for professional/production workflows where session archival and sharing matters.*

