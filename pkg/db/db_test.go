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

	if len(msgs) != 2 {
		t.Fatalf("len(msgs) = %d, want 2", len(msgs))
	}

	if msgs[0].Role != "user" || msgs[0].Content != userPrompt {
		t.Errorf("msgs[0] = %+v, want role user and prompt %q", msgs[0], userPrompt)
	}

	if msgs[1].Role != "assistant" || msgs[1].Content != assistantText {
		t.Errorf("msgs[1] = %+v, want role assistant and text %q", msgs[1], assistantText)
	}

	// 4. Verify updated_at is populated and after created_at
	if session.UpdatedAt.Before(session.CreatedAt) {
		t.Errorf("session.UpdatedAt (%v) is before CreatedAt (%v)", session.UpdatedAt, session.CreatedAt)
	}

	// 5. Test ListSessions
	sessions, err := database.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(sessions))
	}

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("DB file %q does not exist", dbPath)
	}
}
