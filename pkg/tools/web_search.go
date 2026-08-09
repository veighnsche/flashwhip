package tools

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

type WebSearchInput struct {
	Query string `json:"query" jsonschema:"Search terms or question"`
}

type SearchResultItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type WebSearchOutput struct {
	Query   string             `json:"query"`
	Results []SearchResultItem `json:"results"`
}

func performWebSearch(_ agent.Context, in WebSearchInput) (WebSearchOutput, error) {
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return WebSearchOutput{}, fmt.Errorf("search query cannot be empty")
	}

	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("POST", searchURL, strings.NewReader("q="+url.QueryEscape(query)))
	if err != nil {
		return WebSearchOutput{}, fmt.Errorf("failed to create search request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return WebSearchOutput{}, fmt.Errorf("search HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return WebSearchOutput{}, fmt.Errorf("failed to read search response body: %w", err)
	}

	results := parseDuckDuckGoHTML(bodyBytes)
	if len(results) > 6 {
		results = results[:6]
	}

	return WebSearchOutput{
		Query:   query,
		Results: results,
	}, nil
}

func parseDuckDuckGoHTML(htmlBytes []byte) []SearchResultItem {
	doc, err := html.Parse(bytes.NewReader(htmlBytes))
	if err != nil {
		return nil
	}

	var results []SearchResultItem
	var currentTitle, currentURL, currentSnippet string

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Find result title & link
			if n.Data == "a" {
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "result__a") {
						currentTitle = getNodeText(n)
						for _, a := range n.Attr {
							if a.Key == "href" {
								currentURL = decodeDDGURL(a.Val)
							}
						}
					}
				}
			}

			// Find result snippet
			if n.Data == "a" || n.Data == "div" {
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "result__snippet") {
						currentSnippet = getNodeText(n)
						if currentTitle != "" && currentURL != "" {
							results = append(results, SearchResultItem{
								Title:   strings.TrimSpace(currentTitle),
								URL:     strings.TrimSpace(currentURL),
								Snippet: strings.TrimSpace(currentSnippet),
							})
							currentTitle, currentURL, currentSnippet = "", "", ""
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return results
}

func getNodeText(n *html.Node) string {
	var sb strings.Builder
	var extract func(*html.Node)
	extract = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(n)
	return sb.String()
}

func decodeDDGURL(raw string) string {
	if strings.Contains(raw, "uddg=") {
		parts := strings.Split(raw, "uddg=")
		if len(parts) > 1 {
			decoded, err := url.QueryUnescape(strings.Split(parts[1], "&")[0])
			if err == nil {
				return decoded
			}
		}
	}
	return raw
}

func WebSearchTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "web_search",
		Description: "Searches the live web for a query and returns top matching results with titles, snippets, and URLs.",
	}, performWebSearch)
}
