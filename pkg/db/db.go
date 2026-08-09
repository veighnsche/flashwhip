package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"
	_ "modernc.org/sqlite"
)

type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	TurnCount int       `json:"turn_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
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
			err = fmt.Errorf("failed to get user home directory: %w", hErr)
			return
		}

		dbDir := filepath.Join(home, ".flashwhip")
		if MkErr := os.MkdirAll(dbDir, 0755); MkErr != nil {
			err = fmt.Errorf("failed to create database directory: %w", MkErr)
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
		return nil, fmt.Errorf("failed to open sqlite db at %q: %w", dbPath, err)
	}

	database := &DB{sqlDB: sqlDB}
	if err := database.migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
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
		turn_count INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
	`
	_, err := d.sqlDB.Exec(schema)
	return err
}

// SaveMessage stores a text-only message (user or assistant) and updates the session's updated_at timestamp.
func (d *DB) SaveMessage(sessionID, role, content string) error {
	content = strings.TrimSpace(content)
	if content == "" || sessionID == "" {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	// Ensure session exists
	var exists bool
	err := d.sqlDB.QueryRow("SELECT EXISTS(SELECT 1 FROM sessions WHERE id = ?)", sessionID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}

	if !exists {
		title := content
		if len(title) > 60 {
			title = title[:60] + "..."
		}
		turnIncrement := 0
		if role == "assistant" || role == "model" {
			turnIncrement = 1
		}
		_, err = d.sqlDB.Exec("INSERT INTO sessions (id, title, turn_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", sessionID, title, turnIncrement, now, now)
		if err != nil {
			return fmt.Errorf("failed to create session %q: %w", sessionID, err)
		}
	} else {
		if role == "assistant" || role == "model" {
			_, err = d.sqlDB.Exec("UPDATE sessions SET updated_at = ?, turn_count = turn_count + 1 WHERE id = ?", now, sessionID)
		} else {
			_, err = d.sqlDB.Exec("UPDATE sessions SET updated_at = ? WHERE id = ?", now, sessionID)
		}
		if err != nil {
			return fmt.Errorf("failed to update session timestamp for %q: %w", sessionID, err)
		}
	}

	_, err = d.sqlDB.Exec("INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)", sessionID, role, content, now)
	if err != nil {
		return fmt.Errorf("failed to insert message for session %q: %w", sessionID, err)
	}

	return nil
}

// GetSession returns session info and all text messages for sessionID.
func (d *DB) GetSession(sessionID string) (*Session, []Message, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var s Session
	err := d.sqlDB.QueryRow("SELECT id, title, turn_count, created_at, updated_at FROM sessions WHERE id = ?", sessionID).Scan(&s.ID, &s.Title, &s.TurnCount, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get session %q: %w", sessionID, err)
	}

	rows, err := d.sqlDB.Query("SELECT id, session_id, role, content, created_at FROM messages WHERE session_id = ? ORDER BY id ASC", sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query messages for session %q: %w", sessionID, err)
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, nil, err
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
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.Title, &s.TurnCount, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
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
	return d.sqlDB.Close()
}
