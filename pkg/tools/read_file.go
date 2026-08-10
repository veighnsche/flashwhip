package tools

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"flashwhip/pkg/config"
	"flashwhip/pkg/errors"
)

// truncateUTF8 safely truncates string s to at most maxBytes without splitting UTF-8 multi-byte runes.
func truncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	b := []byte(s[:maxBytes])
	for len(b) > 0 && !utf8.Valid(b) {
		b = b[:len(b)-1]
	}
	return string(b)
}

type ReadFileArgs struct {
	FilePath  string `json:"file_path" jsonschema:"The local file path to read"`
	StartLine int    `json:"start_line,omitempty" jsonschema:"Optional 1-indexed line to start reading from (inclusive). Omit to read from the beginning."`
	EndLine   int    `json:"end_line,omitempty" jsonschema:"Optional 1-indexed line to stop reading at (inclusive). Omit to read to the end of file or until the byte cap is reached."`
}

type ReadFileOutput struct {
	FilePath   string `json:"file_path"`
	Content    string `json:"content"`
	Bytes      int    `json:"bytes"`
	TotalLines int    `json:"total_lines"`
	StartLine  int    `json:"start_line,omitempty"`
	EndLine    int    `json:"end_line,omitempty"`
	Truncated  bool   `json:"truncated"`
}

func readFileSnippet(_ agent.Context, args ReadFileArgs) (ReadFileOutput, error) {
	f, err := os.Open(args.FilePath)
	if err != nil {
		return ReadFileOutput{}, errors.Wrapf(errors.ErrCodeToolFileNotFound, err, "failed to open file %q", args.FilePath)
	}
	defer f.Close()

	// Collect all lines so we can report total_lines and do line-range slicing.
	var allLines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return ReadFileOutput{}, errors.Wrapf(errors.ErrCodeToolFileNotFound, err, "failed to read file %q", args.FilePath)
	}

	totalLines := len(allLines)

	// Resolve line range (1-indexed, inclusive).
	start := 1
	end := totalLines
	if args.StartLine > 0 {
		start = args.StartLine
	}
	if args.EndLine > 0 {
		end = args.EndLine
	}
	// Clamp to valid range.
	if start < 1 {
		start = 1
	}
	if end > totalLines {
		end = totalLines
	}
	if start > end {
		return ReadFileOutput{
			FilePath:   args.FilePath,
			TotalLines: totalLines,
			StartLine:  start,
			EndLine:    end,
		}, nil
	}

	startIdx := start - 1
	endIdx := end

	var selectedLines []string
	currentBytes := 0
	actualEndLine := start - 1
	truncated := false

	for i := startIdx; i < endIdx; i++ {
		line := allLines[i]
		lineBytes := len(line)
		needed := lineBytes
		if len(selectedLines) > 0 {
			needed++ // newline delimiter
		}

		if currentBytes+needed > config.MaxFileReadBytes {
			if len(selectedLines) == 0 {
				// If even the first line exceeds MaxFileReadBytes, truncate safely at UTF-8 boundary
				truncatedLine := truncateUTF8(line, config.MaxFileReadBytes)
				selectedLines = append(selectedLines, truncatedLine)
				currentBytes = len(truncatedLine)
				actualEndLine = i + 1
			}
			truncated = true
			break
		}

		selectedLines = append(selectedLines, line)
		currentBytes += needed
		actualEndLine = i + 1
	}

	content := strings.Join(selectedLines, "\n")
	if truncated {
		content += fmt.Sprintf("\n... [content truncated at line %d of %d — use start_line=%d to paginate]", actualEndLine, totalLines, actualEndLine+1)
	}

	return ReadFileOutput{
		FilePath:   args.FilePath,
		Content:    content,
		Bytes:      len(content),
		TotalLines: totalLines,
		StartLine:  start,
		EndLine:    actualEndLine,
		Truncated:  truncated,
	}, nil
}

func ReadFileTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "read_file",
		Description: "Reads the contents of a local file. Supports optional start_line/end_line for paginated reading of large files. Returns total_lines and a truncated flag.",
	}, readFileSnippet)
}
