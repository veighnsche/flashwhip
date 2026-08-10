# Flashwhip UI Issues Index

This directory contains all issues identified from the Flashwhip UI architecture audit.

## Priority Overview

| ID | Title | Priority | Module | Status |
|----|-------|----------|--------|--------|
| #1 | Dead Streaming Buffer (Zero Text Until Done) | 🔴 Critical | runner.go | Open |
| #2 | Dead Code: RenderToolResult() Always Returns Empty | 🔴 Critical | tool_renderer.go | Open |
| #3 | Context Saturation Completely Invisible to Users | 🟡 Major | tracker.go | Open |
| #4 | Static Banner Never Updates After Startup | 🟡 Major | banner.go | Open |
| #5 | No Animated Thinking/Processing Indicator | 🟡 Major | runner.go | Open |
| #6 | Error Feedback Invisible and Undifferentiated | 🔴 Critical | renderer.go | Open |
| #7 | Session/Command System Lacking Discoverability | 🟡 Major | repl.go | Open |
| #8 | No Live Token Counter / Metrics Display | 🟡 Major | tracker.go | Open |
| #9 | Missing Ctrl+C Graceful Exit Handler in REPL | 🟡 Major | repl.go | Open |
| #10 | No Export Conversations Feature | 🟢 Nice-to-have | db.go | Open |

---

## Issue Details

### 🔴 Critical (Impact User Trust/Cores Functionality)

#### [Issue #1: Dead Streaming Buffer](./issue-001-dead-streaming-buffer.md)
Text buffers all tokens silently, renders complete response at once via glamour when done. No streaming feedback during generation — "dead air" perception.

**Fix**: Incremental markdown rendering / typewriter effect as tokens arrive.

---

#### [Issue #2: Dead Code in RenderToolResult](./issue-002-dead-render-tool-result.md)
`RenderToolResult()` always returns empty string. Tool result formatting is dead code — never called, always renders nothing. Affects all text/string-based tool outputs (file reads, search results, command output).

**Fix**: Implement proper rendering or reuse `RenderCombinedToolExecution()`.

---

#### [Issue #6: Error Feedback Invisible](./issue-006-error-feedback-invisible.md)
Tool failures show no color differentiation — only metadata (exit code) visible. No error banner on context pruning. Users can't distinguish "success with empty output" from "tool actually errored."

**Fix**: Add red highlighting for failures, surface context warnings, centralize error display.

---

### 🟡 Major (Significant UX Gaps)

#### [Issue #3: Context Saturation Invisible](./issue-003-invisible-context-saturation.md)
Tracker maintains saturation percentage (`tracker.ContextSaturationPct()`) but UI never displays it. Users have no idea when pruning will occur or why context disappears.

**Fix**: Show live saturation in REPL status bar + warning banners at 75%/90% thresholds.

---

#### [Issue #4: Static Banner Never Updates](./issue-004-static-banner.md)
Banner shows config values (model, context size) only at startup. Stale if model/session changes mid-session or saturation evolves.

**Fix**: Refresh banner on session state transitions; consider compact status bar approach.

---

#### [Issue #5: No Thinking Animation](./issue-005-no-thinking-animation.md)
Only shows static `🧠 [Thinking]` text — no spinner during tool execution. Users perceive app as frozen between tool calls or during network I/O.

**Fix**: Add spinner animation with status messages (reuse existing tracker infrastructure).

---

#### [Issue #7: Command Discoverability](./issue-007-session-command-discoverability.md)
Slash commands exist but have no tab-completion, contextual help is flat text dump, session system has no interactive selection.

**Fix**: Add fuzzy command completion + categorized help + interactive session picker.

---

#### [Issue #8: No Token Counter](./issue-008-no-live-token-counter.md)
Token metrics exist internally but never shown to users. Can't track API costs or diagnose performance issues.

**Fix**: Add `/metrics` command and/or live token counter in REPL status bar.

---

#### [Issue #9: Missing Graceful Exit](./issue-009-missing-graceful-exit-handler.md)
No SIGINT handler during active REPL interactions. Ctrl+C leaves garbled terminal state, no "shutting down" message, potential session corruption.

**Fix**: Add dedicated signal handling with cleanup + status message.

---

### 🟢 Nice-to-Have (Feature Requests)

#### [Issue #10: No Export Feature](./issue-010-no-export-feature.md)
No ability to export or share conversation sessions. Valuable conversations trapped in SQLite database inaccessible for human consumption.

**Fix**: Add `/export session_id --format md|json` command with file output options.

---

## Quick Wins vs Effort Estimate

| Issue | Effort | Impact | Recommended First? |
|-------|--------|--------|---------------------|
| #2 RenderToolResult | Small | High | **Yes** — dead code, quick fix |
| #6 Error Feedback | Medium | High | **Yes** — color differentiation |
| #3 Context Saturation | Small | High | **Yes** — reuse existing tracker |
| #1 Streaming Buffer | Large | High | Consider after fixes above |
| #5 Thinking Animation | Small | Medium | Yes — easy visual polish |
| Others | Varies | Medium-Low | Future iterations |

---

*Created: UI Architecture Audit - Flashwhip v1.0*  
*Date: 2024-08-10*

