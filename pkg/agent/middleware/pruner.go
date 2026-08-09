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

// PruneContents optimizes req.Contents to protect context window & GPU VRAM:
// 1. Caps any single immediate tool response exceeding 2,000 bytes.
// 2. Compacts historical thought blocks (Thought == true) older than maxHistoryTurns.
// 3. Summarizes tool outputs older than maxHistoryTurns down to 200 bytes.
func PruneContents(contents []*genai.Content, maxHistoryTurns int) []*genai.Content {
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
				maxBytes := MaxSingleToolOutputBytes
				if isHistorical {
					maxBytes = MaxHistoricalToolOutputBytes
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
	if len(runes) > maxBytes {
		safeStr = string(runes[:maxBytes])
	}

	summary := fmt.Sprintf("%s... [truncated for context safety]", safeStr)
	return &genai.FunctionResponse{
		Name: fr.Name,
		Response: map[string]any{
			"summary": summary,
		},
	}
}
