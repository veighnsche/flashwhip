# Issue #1 - Zero Streaming Transparency "Dead Air" Experience

**Project**: flashwhip  
**Module**: `pkg/ui` / `pkg/ui/runner.go`  
**Priority**: 🔴 Critical  
**Labels**: UX, streaming, performance  

## Description

The current implementation buffers **all response tokens silently**, then renders the complete text at once via glamour when generation finishes. During this time:

- No text appears on screen (entirely blank)
- User perceives nothing happening ("dead air")
- Only `🧠 [Thinking]` shows *before* tool calls begin, but during actual response generation there's zero visual feedback

## Current Code Path

```go
// pkg/ui/runner.go lines 80-120 (simplified)
for _, part := range parts {
    if textPart != nil {
        textBuf.WriteString(textPart.Text)     // Silently buffers tokens
    }
    if toolCallPart != nil {
        renderToolCall(toolCallPart)          // Shows metadata
        executeTools()
    }
}

// Final render - blocks until ALL turns complete:
if textBuf.Len() > 0 {
    fmt.Print(RenderMarkdown(textBuf.String()))   // 🔴 BLOCKING, delayed output
}
```

## Reproduction Steps

1. Run `flashwhip chat` with a complex prompt requiring multiple tool calls + response generation  
2. Observe blank terminal during model's initial "thinking" phase (tool loops execute)  
3. See complete silence until all content is ready - then flash of rendered markdown  

**Expected**: Text should appear incrementally as tokens stream in, even mid-response  
**Actual**: User waits with nothing on screen, perceives application frozen or unresponsive

## Impact

- **Perceived performance**: App feels "dead" for 10s+ during generation
- **User trust**: Users may think the app hung and kill it prematurely
- **Competitive gap**: All modern conversational AIs stream text character-by-character

## Proposed Fix

Implement streaming rendering so text reveals progressively:

### Option A: Simple Typewriter (Recommended MVP)
```go
// Incremental render at token boundaries instead of buffering all
for _, part := range parts {
    if textPart != nil {
        // Render incrementally using existing glamour renderer after each chunk
        currentText = append(currentText, textPart.Text)
        fmt.Print(RenderMarkdownToBuffer(currentText))  // Live update
    }
}
```

### Option B: Glamour with Streaming Writer Interface
Wrap glamour's output to support partial writes + maintain markdown state across chunks.

### Considerations
- Must handle markdown rendering atomically per complete paragraph (avoid mid-word breaks)
- Cursor position management when interleaving tool calls between responses  
- Memory should not grow unboundedly - flush/trim completed markdown after display

## Acceptance Criteria

- [ ] Text appears within **< 200ms** of first token generation
- [ ] Markdown renders incrementally without visual tearing/flickering
- [ ] Tool call interleaving still works correctly between partial response chunks  
- [ ] Scrollback reflects full streaming history (not just final render)

---
*Created: UI Architecture Audit - Flashwhip v1.0*
