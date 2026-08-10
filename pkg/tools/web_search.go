package tools

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"

	golanghtml "golang.org/x/net/html"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"flashwhip/pkg/errors"
	fnet "flashwhip/pkg/net"
)

type WebSearchInput struct {
	Query string `json:"query" jsonschema:"The web search query"`
}

type SearchResultItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type WebSearchOutput struct {
	Query         string             `json:"query"`
	InstantAnswer string             `json:"instant_answer,omitempty"`
	Results       []SearchResultItem `json:"results"`
}

func performWebSearch(_ agent.Context, in WebSearchInput) (WebSearchOutput, error) {
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return WebSearchOutput{}, errors.New(errors.ErrCodeToolInvalidArgs, "search query cannot be empty")
	}

	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	client := fnet.DefaultHTTPClient()

	req, err := http.NewRequest("POST", searchURL, strings.NewReader("q="+url.QueryEscape(query)))
	if err != nil {
		return WebSearchOutput{}, errors.Wrap(errors.ErrCodeToolNetworkError, "failed to create search request", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return WebSearchOutput{}, errors.Wrap(errors.ErrCodeToolNetworkError, "search HTTP request failed", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return WebSearchOutput{}, errors.Wrap(errors.ErrCodeToolNetworkError, "failed to read search response body", err)
	}

	instantAnswer, results := parseDuckDuckGoHTML(bodyBytes)
	if len(results) > 6 {
		results = results[:6]
	}

	return WebSearchOutput{
		Query:         query,
		InstantAnswer: instantAnswer,
		Results:       results,
	}, nil
}

func parseDuckDuckGoHTML(htmlBytes []byte) (string, []SearchResultItem) {
	doc, err := golanghtml.Parse(bytes.NewReader(htmlBytes))
	if err != nil {
		return "", nil
	}

	var instantAnswer string
	var results []SearchResultItem
	var currentTitle, currentURL, currentSnippet string

	var traverse func(*golanghtml.Node)
	traverse = func(n *golanghtml.Node) {
		if n.Type == golanghtml.ElementNode {
			// Extract instant answer / zero-click box if present
			if instantAnswer == "" {
				for _, attr := range n.Attr {
					if attr.Key == "class" && (strings.Contains(attr.Val, "zci") || strings.Contains(attr.Val, "zero-click") || strings.Contains(attr.Val, "module--weather") || strings.Contains(attr.Val, "msg-optional")) {
						txt := cleanText(getNodeText(n))
						if len(txt) > 10 {
							instantAnswer = txt
						}
					}
				}
			}

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
								Title:   cleanText(currentTitle),
								URL:     strings.TrimSpace(currentURL),
								Snippet: cleanText(currentSnippet),
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
	return instantAnswer, results
}

func cleanText(s string) string {
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\u00a0", " ")
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

func getNodeText(n *golanghtml.Node) string {
	var sb strings.Builder
	var extract func(*golanghtml.Node)
	extract = func(node *golanghtml.Node) {
		if node.Type == golanghtml.TextNode {
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
		Description: "Searches the live web for a query and returns top matching results, snippets, and instant answer cards (e.g. weather, calculations, definitions, facts). Use specific search terms to get direct answers in snippets.",
	}, performWebSearch)
}

