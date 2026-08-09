package tools

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"strings"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

type WebFetchInput struct {
	URL string `json:"url" jsonschema:"The web page HTTP or HTTPS URL to fetch"`
}

type WebFetchOutput struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Content    string `json:"content"`
}

func fetchWebPage(_ agent.Context, in WebFetchInput) (WebFetchOutput, error) {
	rawURL := strings.TrimSpace(in.URL)
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := nurl.Parse(rawURL)
	if err != nil {
		return WebFetchOutput{}, fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return WebFetchOutput{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return WebFetchOutput{}, fmt.Errorf("failed to fetch URL %q: %w", parsedURL.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return WebFetchOutput{
			URL:        parsedURL.String(),
			StatusCode: resp.StatusCode,
			Content:    fmt.Sprintf("HTTP %d error: %s", resp.StatusCode, string(bodyBytes[:minVal(len(bodyBytes), 500)])),
		}, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return WebFetchOutput{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// 1. Stage 2: Readability filtering (strips ads, headers, nav, footers)
	var title string
	var htmlToConvert string

	article, rErr := readability.FromReader(bytes.NewReader(bodyBytes), parsedURL)
	if rErr == nil && article.Node != nil {
		title = article.Title()
		var buf bytes.Buffer
		if renderErr := article.RenderHTML(&buf); renderErr == nil {
			htmlToConvert = buf.String()
		}
	}

	if htmlToConvert == "" {
		htmlToConvert = string(bodyBytes)
	}

	// 2. Stage 3: HTML to Markdown conversion (token reduction + structure preservation)
	markdownText, convErr := htmltomarkdown.ConvertString(htmlToConvert)
	if convErr != nil || strings.TrimSpace(markdownText) == "" {
		markdownText = string(bodyBytes)
	}

	markdownText = strings.TrimSpace(markdownText)
	if len(markdownText) > 4000 {
		markdownText = markdownText[:4000] + "\n\n... [content truncated for token safety]"
	}

	return WebFetchOutput{
		Title:      title,
		URL:        parsedURL.String(),
		StatusCode: resp.StatusCode,
		Content:    markdownText,
	}, nil
}

func minVal(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func WebFetchTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "web_fetch",
		Description: "Fetches a web page URL, cleans clutter with Readability, and converts the content into LLM-optimized Markdown.",
	}, fetchWebPage)
}
