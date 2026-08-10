package tools

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"flashwhip/pkg/errors"
)

type GrepSearchInput struct {
	Query       string `json:"query" jsonschema:"The text query string to search for"`
	RootDir     string `json:"root_dir,omitempty" jsonschema:"Optional root directory to search in (defaults to '.')"`
	FilePattern string `json:"file_pattern,omitempty" jsonschema:"Optional filename pattern filter (e.g. '*.go')"`
	IsRegex     bool   `json:"is_regex,omitempty" jsonschema:"If true, treat query as a regex pattern (default: false)"`
}

type GrepMatch struct {
	FilePath   string `json:"file_path"`
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
}

type GrepSearchOutput struct {
	Query        string      `json:"query"`
	Matches      []GrepMatch `json:"matches"`
	TotalMatches int         `json:"total_matches"`
}

func grepSearchCodebase(_ agent.Context, in GrepSearchInput) (GrepSearchOutput, error) {
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return GrepSearchOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "query cannot be empty")
	}

	rootDir := in.RootDir
	if rootDir == "" {
		rootDir = "."
	}

	var re *regexp.Regexp
	if in.IsRegex {
		var compileErr error
		re, compileErr = regexp.Compile(query)
		if compileErr != nil {
			return GrepSearchOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, fmt.Sprintf("invalid regex pattern: %v", compileErr))
		}
	}

	var matches []GrepMatch
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, wErr error) error {
		if wErr != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		if in.FilePattern != "" {
			matched, _ := filepath.Match(in.FilePattern, d.Name())
			if !matched {
				return nil
			}
		}

		// Skip binary files
		if isBinaryExt(d.Name()) {
			return nil
		}

		file, oErr := os.Open(path)
		if oErr != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 1
		for scanner.Scan() {
			line := scanner.Text()
			match := false
			if re != nil {
				match = re.MatchString(line)
			} else {
				match = strings.Contains(line, query)
			}
			if match {
				matches = append(matches, GrepMatch{
					FilePath:   path,
					LineNumber: lineNum,
					Content:    strings.TrimSpace(line),
				})
				if len(matches) >= 100 {
					return filepath.SkipAll
				}
			}
			lineNum++
		}
		if err := scanner.Err(); err != nil {
			return nil
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return GrepSearchOutput{}, errors.Wrap(errors.ErrCodeToolFileNotFound, "grep search failed", err)
	}

	return GrepSearchOutput{
		Query:        query,
		Matches:      matches,
		TotalMatches: len(matches),
	}, nil
}

func isBinaryExt(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".ico", ".pdf", ".zip", ".tar", ".gz", ".exe", ".dll", ".so", ".dylib", ".db", ".sqlite":
		return true
	default:
		return false
	}
}

func GrepSearchTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "grep_search",
		Description: "Searches codebase source files line-by-line for a string query, returning file paths, line numbers, and matching lines.",
	}, grepSearchCodebase)
}
