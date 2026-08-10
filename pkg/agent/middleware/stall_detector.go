package middleware

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ToolCallRecord stores normalized signature for a single tool call.
type ToolCallRecord struct {
	Name          string
	CanonicalArgs string
	Signature     string
}

// LoopPattern describes the type of stalled execution loop detected.
type LoopPattern string

const (
	PatternConsecutive LoopPattern = "consecutive"
	PatternOscillating LoopPattern = "oscillating"
)

// LoopInfo contains details when a stall loop is detected.
type LoopInfo struct {
	ToolName         string
	ConsecutiveCount int
	Pattern          LoopPattern
	Signature        string
}

// StallDetector tracks tool call history during a session to detect repetitive loops.
type StallDetector struct {
	mu             sync.Mutex
	maxConsecutive int
	history        []ToolCallRecord
}

// NewStallDetector initializes a StallDetector with a given max consecutive repeat threshold (default 3).
func NewStallDetector(maxConsecutive int) *StallDetector {
	if maxConsecutive <= 0 {
		maxConsecutive = 3
	}
	return &StallDetector{
		maxConsecutive: maxConsecutive,
		history:        make([]ToolCallRecord, 0),
	}
}

// CanonicalSignature returns a deterministic signature for a tool name and arguments map.
func CanonicalSignature(name string, args map[string]any) string {
	cleanName := strings.TrimSpace(strings.ToLower(name))
	canonicalArgs := CanonicalizeArgs(args)
	return fmt.Sprintf("%s:%s", cleanName, canonicalArgs)
}

// CanonicalizeArgs produces a stable JSON string with recursively sorted map keys.
func CanonicalizeArgs(v any) string {
	if v == nil {
		return "{}"
	}
	normalized := sortMapKeys(v)
	b, err := json.Marshal(normalized)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func sortMapKeys(v any) any {
	switch val := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sortedMap := make(map[string]any, len(val))
		for _, k := range keys {
			sortedMap[k] = sortMapKeys(val[k])
		}
		return sortedMap
	case []any:
		sortedSlice := make([]any, len(val))
		for i, item := range val {
			sortedSlice[i] = sortMapKeys(item)
		}
		return sortedSlice
	default:
		return val
	}
}

// RecordCall registers a tool call and returns whether a loop/stall pattern was detected.
func (d *StallDetector) RecordCall(name string, args map[string]any) (bool, LoopInfo) {
	d.mu.Lock()
	defer d.mu.Unlock()

	sig := CanonicalSignature(name, args)
	rec := ToolCallRecord{
		Name:          name,
		CanonicalArgs: CanonicalizeArgs(args),
		Signature:     sig,
	}
	d.history = append(d.history, rec)

	n := len(d.history)

	// 1. Check for consecutive identical tool calls
	if n >= d.maxConsecutive {
		consecutive := true
		targetSig := d.history[n-1].Signature
		for i := n - d.maxConsecutive; i < n; i++ {
			if d.history[i].Signature != targetSig {
				consecutive = false
				break
			}
		}
		if consecutive {
			return true, LoopInfo{
				ToolName:         name,
				ConsecutiveCount: d.maxConsecutive,
				Pattern:          PatternConsecutive,
				Signature:        sig,
			}
		}
	}

	// 2. Check for alternating 2-tool cycle (A, B, A, B, A, B) over 6 steps
	const cycleLen = 6
	if n >= cycleLen {
		h := d.history[n-cycleLen:]
		if h[0].Signature == h[2].Signature && h[2].Signature == h[4].Signature &&
			h[1].Signature == h[3].Signature && h[3].Signature == h[5].Signature &&
			h[0].Signature != h[1].Signature {
			return true, LoopInfo{
				ToolName:         name,
				ConsecutiveCount: cycleLen,
				Pattern:          PatternOscillating,
				Signature:        sig,
			}
		}
	}

	return false, LoopInfo{}
}

// Reset clears the detector history.
func (d *StallDetector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.history = d.history[:0]
}
