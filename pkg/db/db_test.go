package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteDB(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	database, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer database.Close()

	sessionID := "test-session-1"

	// 1. Save user prompt
	userPrompt := "What is the weather in Amsterdam?"
	if err := database.SaveMessage(sessionID, "user", userPrompt); err != nil {
		t.Fatalf("SaveMessage user failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	// 2. Save final assistant text response (no thinking, no tool calls)
	assistantText := "The current temperature in Amsterdam is 22°C with clear skies."
	if err := database.SaveMessage(sessionID, "assistant", assistantText); err != nil {
		t.Fatalf("SaveMessage assistant failed: %v", err)
	}

	// 3. Verify session info and messages
	session, msgs, err := database.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if session.ID != sessionID {
		t.Errorf("session.ID = %q, want %q", session.ID, sessionID)
	}

	if session.TurnCount != 1 {
		t.Errorf("session.TurnCount = %d, want 1", session.TurnCount)
	}

	if len(msgs) != 2 {
		t.Fatalf("len(msgs) = %d, want 2", len(msgs))
	}

	// 4. Verify GenAI contents reconstruction
	genaiContents, err := database.GetSessionGenAIContents(sessionID)
	if err != nil {
		t.Fatalf("GetSessionGenAIContents failed: %v", err)
	}

	if len(genaiContents) != 2 {
		t.Fatalf("len(genaiContents) = %d, want 2", len(genaiContents))
	}

	if genaiContents[0].Role != "user" || genaiContents[0].Parts[0].Text != userPrompt {
		t.Errorf("genaiContents[0] = %+v, want role user", genaiContents[0])
	}

	if genaiContents[1].Role != "model" || genaiContents[1].Parts[0].Text != assistantText {
		t.Errorf("genaiContents[1] = %+v, want role model", genaiContents[1])
	}

	// 5. Test FormatRelativeTime
	relTime := FormatRelativeTime(session.UpdatedAt)
	if relTime != "just now" {
		t.Errorf("FormatRelativeTime = %q, want 'just now'", relTime)
	}

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("DB file %q does not exist", dbPath)
	}
}

func TestReplaceSessionGenAIContents(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_replace.db")

	database, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer database.Close()

	sessionID := "test-replace-session"
	if err := database.SaveMessage(sessionID, "user", "Hello"); err != nil {
		t.Fatalf("SaveMessage user failed: %v", err)
	}
	if err := database.SaveMessage(sessionID, "assistant", "Hi there, long text payload here..."); err != nil {
		t.Fatalf("SaveMessage assistant failed: %v", err)
	}

	contents, err := database.GetSessionGenAIContents(sessionID)
	if err != nil || len(contents) != 2 {
		t.Fatalf("Initial contents check failed: %v, len=%d", err, len(contents))
	}

	// Replace with pruned content (1 content item)
	pruned := contents[1:]
	if err := database.ReplaceSessionGenAIContents(sessionID, pruned); err != nil {
		t.Fatalf("ReplaceSessionGenAIContents failed: %v", err)
	}

	newContents, err := database.GetSessionGenAIContents(sessionID)
	if err != nil {
		t.Fatalf("GetSessionGenAIContents after replace failed: %v", err)
	}
	if len(newContents) != 1 {
		t.Fatalf("Expected 1 content item after replace, got %d", len(newContents))
	}
}

