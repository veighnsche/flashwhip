package ollama

import (
	"fmt"
	"sync"
)

// Usage tracks cumulative token consumption across conversation turns.
type Usage struct {
	mu            sync.Mutex
	promptTokens  int
	completionTok int
	totalTokens   int
}

// NewUsage creates a new usage tracker.
func NewUsage() *Usage {
	return &Usage{}
}

// Record adds tokens from a single response to the cumulative totals.
func (u *Usage) Record(prompt, completion, total int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.promptTokens += prompt
	u.completionTok += completion
	u.totalTokens += total
}

// Get returns the current cumulative token counts.
func (u *Usage) Get() (prompt, completion, total int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.promptTokens, u.completionTok, u.totalTokens
}

// EstimatePercentage calculates usage as a percentage of context_length.
// Returns -1 if context_length is zero (unknown).
func (u *Usage) EstimatePercentage(contextLength int) float64 {
	u.mu.Lock()
	defer u.mu.Unlock()
	if contextLength == 0 {
		return -1
	}
	return float64(u.totalTokens) / float64(contextLength) * 100
}

func (u *Usage) String() string {
	prompt, completion, total := u.Get()
	return fmt.Sprintf("tokens: %d (prompt: %d,completion: %d)", total, prompt, completion)
}

// Reset zeroes all counters.
func (u *Usage) Reset() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.promptTokens = 0
	u.completionTok = 0
	u.totalTokens = 0
}
