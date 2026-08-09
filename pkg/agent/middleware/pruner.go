package middleware

import (
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

// PruneContents prunes historical tool results in req.Contents to prevent context window bloat during multi-turn chat.
func PruneContents(contents []*genai.Content, maxHistoryTurns int) []*genai.Content {
	if len(contents) <= maxHistoryTurns*2 {
		return contents
	}

	pruned := make([]*genai.Content, len(contents))
	cutoffIndex := len(contents) - (maxHistoryTurns * 2)

	for i, c := range contents {
		if c == nil {
			continue
		}

		// For turns older than cutoff, prune long FunctionResponse contents
		if i < cutoffIndex {
			var newParts []*genai.Part
			for _, p := range c.Parts {
				if p == nil {
					continue
				}
				if p.FunctionResponse != nil {
					newResp := pruneFunctionResponse(p.FunctionResponse)
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
		} else {
			pruned[i] = c
		}
	}

	return pruned
}

func pruneFunctionResponse(fr *genai.FunctionResponse) *genai.FunctionResponse {
	if fr == nil || fr.Response == nil {
		return fr
	}

	respBytes, err := json.Marshal(fr.Response)
	if err != nil || len(respBytes) <= 200 {
		return fr
	}

	summary := fmt.Sprintf("%s... [historical output truncated for context safety]", string(respBytes[:200]))
	return &genai.FunctionResponse{
		Name: fr.Name,
		Response: map[string]any{
			"summary": summary,
		},
	}
}
