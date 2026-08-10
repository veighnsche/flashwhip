package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"
	_ "modernc.org/sqlite"

	"flashwhip/pkg/errors"
)

type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Workspace string    `json:"workspace"`
	TurnCount int       `json:"turn_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID          int64     `json:"id"`
	SessionID   string    `json:"session_id"`
	Role        string    `json:"role"`
	Content     string    `json:"content"`
	JSONPayload string    `json:"json_payload,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type DB struct {
	sqlDB *sql.DB
	mu    sync.Mutex
}

var (
	defaultDB *DB
	once      sync.Once
)

// DefaultDB returns the global embedded database instance stored in ~/.flashwhip/flashwhip.db.
func DefaultDB() (*DB, error) {
	var err error
	once.Do(func() {
		home, hErr := os.UserHomeDir()
		if hErr != nil {
			err = errors.Wrap(errors.ErrCodeDBOpenFailed, "failed to get user home directory", hErr)
			return
		}

		dbDir := filepath.Join(home, ".flashwhip")
		if MkErr := os.MkdirAll(dbDir, 0755); MkErr != nil {
			err = errors.Wrap(errors.ErrCodeDBOpenFailed, "failed to create database directory", MkErr)
			return
		}

		dbPath := filepath.Join(dbDir, "flashwhip.db")
		defaultDB, err = OpenDB(dbPath)
	})
	return defaultDB, err
}

// OpenDB initializes or opens an embedded SQLite database file at dbPath.
func OpenDB(dbPath string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, errors.Wrapf(errors.ErrCodeDBOpenFailed, err, "failed to open sqlite db at %q", dbPath)
	}

	database := &DB{sqlDB: sqlDB}
	if err := database.migrate(); err != nil {
		sqlDB.Close()
		return nil, errors.Wrap(errors.ErrCodeDBMigrationFailed, "failed to run migrations", err)
	}

	return database, nil
}

func (d *DB) migrate() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		workspace TEXT NOT NULL DEFAULT '',
		turn_count INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		json_payload TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
	`
	if _, err := d.sqlDB.Exec(schema); err != nil {
		return errors.Wrap(errors.ErrCodeDBMigrationFailed, "schema execution failed", err)
	}

	// Migrations for pre-existing tables
	_, _ = d.sqlDB.Exec("ALTER TABLE sessions ADD COLUMN turn_count INTEGER NOT NULL DEFAULT 0;")
	_, _ = d.sqlDB.Exec("ALTER TABLE sessions ADD COLUMN workspace TEXT NOT NULL DEFAULT '';")
	_, _ = d.sqlDB.Exec("ALTER TABLE messages ADD COLUMN json_payload TEXT;")

	return nil
}

// SaveContent saves a structured genai.Content object (including text & tool calls) and updates session metadata.
func (d *DB) SaveContent(sessionID string, content *genai.Content) error {
	if content == nil || sessionID == "" {
		return nil
	}

	if content.Role == "assistant" {
		content.Role = "model"
	}

	textSummary := summarizeContent(content)

	jsonBytes, err := json.Marshal(content)
	jsonPayload := ""
	if err == nil {
		jsonPayload = string(jsonBytes)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	pwd, _ := os.Getwd()

	var exists bool
	err = d.sqlDB.QueryRow("SELECT EXISTS(SELECT 1 FROM sessions WHERE id = ?)", sessionID).Scan(&exists)
	if err != nil {
		return errors.Wrap(errors.ErrCodeDBQueryFailed, "failed to check session existence", err)
	}

	if !exists {
		// Defer title assignment: only the first user message with real text becomes
		// the session title. Tool-call/model messages (which have no user text) get an
		// empty title that is backfilled on the first user message instead of a useless
		// "[model message]" placeholder.
		title := ""
		if content.Role == "user" && textSummary != "" {
			title = truncateTitle(textSummary)
		}
		turnIncrement := 0
		if content.Role == "assistant" || content.Role == "model" {
			turnIncrement = 1
		}
		_, err = d.sqlDB.Exec("INSERT INTO sessions (id, title, workspace, turn_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", sessionID, title, pwd, turnIncrement, now, now)
		if err != nil {
			return errors.Wrapf(errors.ErrCodeDBSaveFailed, err, "failed to create session %q", sessionID)
		}
	} else {
		if content.Role == "assistant" || content.Role == "model" {
			_, err = d.sqlDB.Exec("UPDATE sessions SET updated_at = ?, turn_count = turn_count + 1 WHERE id = ?", now, sessionID)
		} else {
			// Backfill a deferred/placeholder title on the first user message with real text.
			if textSummary != "" {
				_, err = d.sqlDB.Exec("UPDATE sessions SET title = ?, updated_at = ? WHERE id = ? AND (title = '' OR title LIKE '[% message]')", truncateTitle(textSummary), now, sessionID)
			}
			if err == nil {
				_, err = d.sqlDB.Exec("UPDATE sessions SET updated_at = ? WHERE id = ?", now, sessionID)
			}
		}
		if err != nil {
			return errors.Wrapf(errors.ErrCodeDBSaveFailed, err, "failed to update session timestamp for %q", sessionID)
		}
	}

	_, err = d.sqlDB.Exec("INSERT INTO messages (session_id, role, content, json_payload, created_at) VALUES (?, ?, ?, ?, ?)", sessionID, content.Role, textSummary, jsonPayload, now)
	if err != nil {
		return errors.Wrapf(errors.ErrCodeDBSaveFailed, err, "failed to insert message for session %q", sessionID)
	}

	return nil
}

// truncateTitle shortens a session title to at most 60 runes (rune-safe, avoiding
// splitting a UTF-8 rune in half), appending an ellipsis when truncated.
func truncateTitle(s string) string {
	const maxRunes = 60
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes]) + "..."
}

// SaveMessage stores a text-only message (user or assistant) and updates the session's updated_at timestamp.
func (d *DB) SaveMessage(sessionID, role, content string) error {
	c := &genai.Content{
		Role: role,
		Parts: []*genai.Part{
			{Text: content},
		},
	}
	return d.SaveContent(sessionID, c)
}

// summarizeContent collapses a content's non-thought text parts into a single
// display string, falling back to a role placeholder when there is no text.
func summarizeContent(c *genai.Content) string {
	var textParts []string
	for _, p := range c.Parts {
		if p == nil {
			continue
		}
		if p.Text != "" && !p.Thought {
			textParts = append(textParts, p.Text)
		}
	}
	textSummary := strings.Join(textParts, " ")
	if textSummary == "" && len(c.Parts) > 0 {
		textSummary = fmt.Sprintf("[%s message]", c.Role)
	}
	return textSummary
}

// ReplaceSessionGenAIContents replaces stored session messages in SQLite with a pruned content set.
func (d *DB) ReplaceSessionGenAIContents(sessionID string, contents []*genai.Content) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.sqlDB.Begin()
	if err != nil {
		return errors.Wrapf(errors.ErrCodeDBSaveFailed, err, "failed to start tx for session replacement %q", sessionID)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.Exec("DELETE FROM messages WHERE session_id = ?", sessionID)
	if err != nil {
		return errors.Wrapf(errors.ErrCodeDBSaveFailed, err, "failed to clear messages for session %q", sessionID)
	}

	now := time.Now()
	for _, content := range contents {
		if content == nil {
			continue
		}
		textSummary := summarizeContent(content)

		jsonBytes, err := json.Marshal(content)
		jsonPayload := ""
		if err == nil {
			jsonPayload = string(jsonBytes)
		}

		_, err = tx.Exec("INSERT INTO messages (session_id, role, content, json_payload, created_at) VALUES (?, ?, ?, ?, ?)",
			sessionID, content.Role, textSummary, jsonPayload, now)
		if err != nil {
			return errors.Wrapf(errors.ErrCodeDBSaveFailed, err, "failed to re-insert pruned message for session %q", sessionID)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrapf(errors.ErrCodeDBSaveFailed, err, "failed to commit transaction for session %q", sessionID)
	}

	return nil
}

// GetSession returns session info and all text messages for sessionID.
func (d *DB) GetSession(sessionID string) (*Session, []Message, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var s Session
	err := d.sqlDB.QueryRow("SELECT id, title, workspace, turn_count, created_at, updated_at FROM sessions WHERE id = ?", sessionID).Scan(&s.ID, &s.Title, &s.Workspace, &s.TurnCount, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, nil, errors.Wrapf(errors.ErrCodeDBSessionNotFound, err, "failed to get session %q", sessionID)
	}

	rows, err := d.sqlDB.Query("SELECT id, session_id, role, content, json_payload, created_at FROM messages WHERE session_id = ? ORDER BY id ASC", sessionID)
	if err != nil {
		return nil, nil, errors.Wrapf(errors.ErrCodeDBQueryFailed, err, "failed to query messages for session %q", sessionID)
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		var jsonPayload sql.NullString
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &jsonPayload, &m.CreatedAt); err != nil {
			return nil, nil, errors.Wrap(errors.ErrCodeDBQueryFailed, "row scan failed", err)
		}
		if jsonPayload.Valid {
			m.JSONPayload = jsonPayload.String
		}
		msgs = append(msgs, m)
	}

	return &s, msgs, nil
}

// GetSessionGenAIContents reconstructs GenAI Content history from SQLite for session continuation.
func (d *DB) GetSessionGenAIContents(sessionID string) ([]*genai.Content, error) {
	_, msgs, err := d.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	var contents []*genai.Content
	for _, m := range msgs {
		if m.JSONPayload != "" {
			var gc genai.Content
			if unmarshalErr := json.Unmarshal([]byte(m.JSONPayload), &gc); unmarshalErr == nil {
				contents = append(contents, &gc)
				continue
			}
		}

		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role: role,
			Parts: []*genai.Part{
				{Text: m.Content},
			},
		})
	}
	return contents, nil
}

// ListSessions returns all recorded sessions sorted by updated_at descending.
func (d *DB) ListSessions() ([]Session, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	rows, err := d.sqlDB.Query("SELECT id, title, turn_count, created_at, updated_at FROM sessions ORDER BY updated_at DESC")
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeDBQueryFailed, "failed to list sessions", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.Title, &s.TurnCount, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, errors.Wrap(errors.ErrCodeDBQueryFailed, "row scan failed", err)
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

// FormatRelativeTime formats a duration into human readable relative time.
func FormatRelativeTime(t time.Time) string {
	diff := time.Since(t)
	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "yesterday"
	}
	return fmt.Sprintf("%d days ago", days)
}

// Close closes the underlying SQL database.
func (d *DB) Close() error {
	if err := d.sqlDB.Close(); err != nil {
		return errors.Wrap(errors.ErrCodeDBCloseFailed, "failed to close database", err)
	}
	return nil
}
