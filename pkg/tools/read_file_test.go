package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateUTF8(t *testing.T) {
	// 3-byte UTF-8 character '€' (0xE2, 0x82, 0xAC)
	input := "Hello € World" // len = 15 bytes
	// "Hello " is 6 bytes. '€' is 3 bytes (bytes 6, 7, 8).

	// Truncating at 7 bytes cuts '€' in half (6 bytes 'Hello ' + 1 byte 0xE2)
	got := truncateUTF8(input, 7)
	if !utf8.ValidString(got) {
		t.Errorf("truncateUTF8 produced invalid UTF-8 string: %q", got)
	}
	if got != "Hello " {
		t.Errorf("truncateUTF8(7) = %q, want %q", got, "Hello ")
	}

	// Truncating at 9 bytes includes full '€'
	gotFull := truncateUTF8(input, 9)
	if gotFull != "Hello €" {
		t.Errorf("truncateUTF8(9) = %q, want %q", gotFull, "Hello €")
	}
}

func TestReadFileTool_LineAlignedTruncation(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large.txt")

	// Create 1000 lines of text
	var lines []string
	for i := 1; i <= 1000; i++ {
		lines = append(lines, "Line "+strings.Repeat("x", 40)+" "+string(rune('A'+(i%26))))
	}
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	out, err := readFileSnippet(nil, ReadFileArgs{
		FilePath: filePath,
	})
	if err != nil {
		t.Fatalf("readFileSnippet failed: %v", err)
	}

	if out.TotalLines != 1000 {
		t.Errorf("TotalLines = %d, want 1000", out.TotalLines)
	}

	if !out.Truncated {
		t.Errorf("expected Truncated = true for large file")
	}

	// Verify output does not end mid-line (except for truncation header)
	contentLines := strings.Split(out.Content, "\n")
	lastLine := contentLines[len(contentLines)-1]
	if !strings.HasPrefix(lastLine, "... [content truncated at line") {
		t.Errorf("expected truncation footer, got: %q", lastLine)
	}

	// Verify EndLine matches the actual rendered line count
	if out.EndLine <= 0 || out.EndLine >= 1000 {
		t.Errorf("expected actual EndLine to reflect truncated line count, got %d", out.EndLine)
	}
}

func TestReadFileTool_LineRange(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "range.txt")
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	_ = os.WriteFile(filePath, []byte(content), 0644)

	out, err := readFileSnippet(nil, ReadFileArgs{
		FilePath:  filePath,
		StartLine: 2,
		EndLine:   4,
	})
	if err != nil {
		t.Fatalf("readFileSnippet failed: %v", err)
	}

	if out.StartLine != 2 || out.EndLine != 4 {
		t.Errorf("StartLine=%d, EndLine=%d, want 2 and 4", out.StartLine, out.EndLine)
	}

	expectedContent := "Line 2\nLine 3\nLine 4"
	if out.Content != expectedContent {
		t.Errorf("Content = %q, want %q", out.Content, expectedContent)
	}
	if out.Truncated {
		t.Errorf("expected Truncated = false")
	}
}
