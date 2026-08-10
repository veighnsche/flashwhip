package middleware

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

const (
	MaxSingleToolOutputBytes     = 2000
	MaxHistoricalToolOutputBytes = 200
)

// PruneContents optimizes req.Contents to protect context window & GPU VRAM
// using default turn and size thresholds.
func PruneContents(contents []*genai.Content, maxHistoryTurns int) []*genai.Content {
	return PruneContentsWithLimits(contents, maxHistoryTurns, MaxHistoricalToolOutputBytes, MaxSingleToolOutputBytes)
}

// PruneContentsAdaptive dynamically selects pruning thresholds based on context saturation percentage.
func PruneContentsAdaptive(contents []*genai.Content, contextPct float64) []*genai.Content {
	var maxHistoryTurns, historicalToolBytes, singleToolBytes int

	switch {
	case contextPct >= 90.0:
		// Emergency Compaction: Keep only 1 recent turn, truncate historical tools to 75B
		maxHistoryTurns = 1
		historicalToolBytes = 75
		singleToolBytes = 1000
	case contextPct >= 75.0:
		// Soft Compaction: Keep last 2 turns intact, truncate historical tools to 150B
		maxHistoryTurns = 2
		historicalToolBytes = 150
		singleToolBytes = 1500
	default:
		// Standard Pruning: Keep last 5 turns intact
		maxHistoryTurns = 5
		historicalToolBytes = MaxHistoricalToolOutputBytes
		singleToolBytes = MaxSingleToolOutputBytes
	}

	return PruneContentsWithLimits(contents, maxHistoryTurns, historicalToolBytes, singleToolBytes)
}

// PruneContentsWithLimits optimizes req.Contents with explicit max turns and byte bounds.
func PruneContentsWithLimits(contents []*genai.Content, maxHistoryTurns int, historicalToolBytes int, singleToolBytes int) []*genai.Content {
	if len(contents) == 0 {
		return contents
	}

	pruned := make([]*genai.Content, len(contents))
	cutoffIndex := len(contents) - (maxHistoryTurns * 2)
	if cutoffIndex < 0 {
		cutoffIndex = 0
	}

	for i, c := range contents {
		if c == nil {
			continue
		}

		isHistorical := i < cutoffIndex
		var newParts []*genai.Part

		for _, p := range c.Parts {
			if p == nil {
				continue
			}

			// Compact historical thought/scratchpad blocks
			if p.Thought {
				if isHistorical {
					// Omit historical thought blocks to conserve tokens
					continue
				}
				newParts = append(newParts, p)
				continue
			}

			// Handle FunctionResponse pruning & capping
			if p.FunctionResponse != nil {
				maxBytes := singleToolBytes
				if isHistorical {
					maxBytes = historicalToolBytes
				}
				newResp := pruneFunctionResponse(p.FunctionResponse, maxBytes)
				newParts = append(newParts, &genai.Part{
					FunctionResponse: newResp,
				})
			} else {
				newParts = append(newParts, p)
			}
		}

		pruned[i] = &genai.Content{
			Role:  c.Role,
			Parts: newParts,
		}
	}

	return pruned
}

func pruneFunctionResponse(fr *genai.FunctionResponse, maxBytes int) *genai.FunctionResponse {
	if fr == nil || fr.Response == nil {
		return fr
	}

	respBytes, err := json.Marshal(fr.Response)
	if err != nil || len(respBytes) <= maxBytes {
		return fr
	}

	safeStr := strings.ToValidUTF8(string(respBytes), "")
	runes := []rune(safeStr)
	if len(runes) <= maxBytes {
		return fr
	}

	// Head + Tail Preservation Strategy (60% Head, 40% Tail)
	headLen := (maxBytes * 60) / 100
	if headLen < 10 {
		headLen = 10
	}
	tailLen := maxBytes - headLen
	if tailLen < 5 {
		tailLen = 5
	}

	headStr := string(runes[:headLen])
	tailStr := string(runes[len(runes)-tailLen:])
	omittedBytes := len(safeStr) - (headLen + tailLen)

	summary := fmt.Sprintf("%s\n... [%d bytes omitted for context safety] ...\n%s", headStr, omittedBytes, tailStr)
	return &genai.FunctionResponse{
		Name: fr.Name,
		Response: map[string]any{
			"summary": summary,
		},
	}
}

