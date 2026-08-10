package tools

import (
	"bufio"
	"os"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"flashwhip/pkg/errors"
)

const maxReadFileBytes = 20_000

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

	selectedLines := allLines[start-1 : end]
	content := strings.Join(selectedLines, "\n")

	truncated := false
	if len(content) > maxReadFileBytes {
		content = content[:maxReadFileBytes] + "\n... [content truncated — use start_line/end_line to paginate]"
		truncated = true
	}

	return ReadFileOutput{
		FilePath:   args.FilePath,
		Content:    content,
		Bytes:      len(content),
		TotalLines: totalLines,
		StartLine:  start,
		EndLine:    end,
		Truncated:  truncated,
	}, nil
}

func ReadFileTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "read_file",
		Description: "Reads the contents of a local file. Supports optional start_line/end_line for paginated reading of large files. Returns total_lines and a truncated flag.",
	}, readFileSnippet)
}
