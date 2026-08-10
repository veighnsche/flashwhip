package tools

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	golanghtml "golang.org/x/net/html"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"flashwhip/pkg/errors"
	fnet "flashwhip/pkg/net"
)

type WebFetchInput struct {
	URL string `json:"url" jsonschema:"The web page URL to fetch content from"`
}

type WebFetchOutput struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Content    string `json:"content"`
}

func decompressGzip(data []byte) ([]byte, error) {
	if bytes.HasPrefix(data, []byte("\x1f\x8b")) {
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return data, err
		}
		defer r.Close()
		return io.ReadAll(r)
	}
	return data, nil
}

func fetchWebPage(_ agent.Context, in WebFetchInput) (WebFetchOutput, error) {
	rawURL := strings.TrimSpace(in.URL)
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := nurl.Parse(rawURL)
	if err != nil {
		return WebFetchOutput{}, errors.Wrapf(errors.ErrCodeToolInvalidArgs, err, "invalid URL %q", rawURL)
	}

	client := fnet.DefaultHTTPClient()

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return WebFetchOutput{}, errors.Wrap(errors.ErrCodeToolNetworkError, "failed to create request", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return WebFetchOutput{}, errors.Wrapf(errors.ErrCodeToolNetworkError, err, "failed to fetch URL %q", parsedURL.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyBytes, _ = decompressGzip(bodyBytes)
		snippet := cleanText(string(bodyBytes[:minVal(len(bodyBytes), 300)]))
		return WebFetchOutput{
			URL:        parsedURL.String(),
			StatusCode: resp.StatusCode,
			Content:    fmt.Sprintf("HTTP %d Error: Page blocked or unavailable. Tip: This URL is blocking automated requests. Search for alternative sources or rely on search snippets instead of retrying this URL.\nDetails: %s", resp.StatusCode, snippet),
		}, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return WebFetchOutput{}, errors.Wrap(errors.ErrCodeToolNetworkError, "failed to read response body", err)
	}
	bodyBytes, _ = decompressGzip(bodyBytes)

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
		htmlToConvert = cleanHTMLDOM(bodyBytes)
	}

	// 2. Stage 3: HTML to Markdown conversion (token reduction + structure preservation)
	markdownText, convErr := htmltomarkdown.ConvertString(htmlToConvert)
	if convErr != nil || strings.TrimSpace(markdownText) == "" {
		markdownText = string(bodyBytes)
	}

	markdownText = strings.TrimSpace(markdownText)

	isBlocked := resp.StatusCode == 403 || resp.StatusCode == 429 ||
		strings.Contains(markdownText, "Cloudflare") ||
		strings.Contains(markdownText, "Attention Required! | Cloudflare") ||
		strings.Contains(markdownText, "Access Denied") ||
		strings.Contains(markdownText, "Just a moment...")

	// Strategy 2: Fallback to Jina Reader Proxy (r.jina.ai) if direct fetch is blocked
	if isBlocked {
		jinaURL := fmt.Sprintf("https://r.jina.ai/%s", parsedURL.String())
		jinaReq, jErr := http.NewRequest("GET", jinaURL, nil)
		if jErr == nil {
			jinaReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
			jinaResp, doErr := client.Do(jinaReq)
			if doErr == nil {
				defer jinaResp.Body.Close()
				if jinaResp.StatusCode == http.StatusOK {
					jinaBytes, _ := io.ReadAll(jinaResp.Body)
					jinaBytes, _ = decompressGzip(jinaBytes)
					jinaText := cleanText(string(jinaBytes))
					if len(jinaText) > 100 && !strings.Contains(jinaText, "403 Forbidden") {
						if len(jinaText) > 4000 {
							jinaText = jinaText[:4000] + "\n\n... [content truncated for token safety]"
						}
						return WebFetchOutput{
							Title:      title,
							URL:        parsedURL.String(),
							StatusCode: http.StatusOK,
							Content:    jinaText,
						}, nil
					}
				}
			}
		}

		markdownText = fmt.Sprintf("Anti-bot protection page detected (Cloudflare/BotGuard). Tip: Search for alternative sources or rely on web_search snippets.\n\n%s", markdownText[:minVal(len(markdownText), 400)])
	} else if len(markdownText) > 4000 {
		markdownText = markdownText[:4000] + "\n\n... [content truncated for token safety]"
	}

	return WebFetchOutput{
		Title:      title,
		URL:        parsedURL.String(),
		StatusCode: resp.StatusCode,
		Content:    markdownText,
	}, nil
}

func cleanHTMLDOM(htmlBytes []byte) string {
	doc, err := golanghtml.Parse(bytes.NewReader(htmlBytes))
	if err != nil {
		return string(htmlBytes)
	}

	var removeUnwanted func(*golanghtml.Node)
	removeUnwanted = func(n *golanghtml.Node) {
		for c := n.FirstChild; c != nil; {
			next := c.NextSibling
			if c.Type == golanghtml.ElementNode {
				tag := strings.ToLower(c.Data)
				if tag == "script" || tag == "style" || tag == "nav" || tag == "header" || tag == "footer" || tag == "noscript" || tag == "svg" || tag == "iframe" {
					n.RemoveChild(c)
					c = next
					continue
				}
			}
			removeUnwanted(c)
			c = next
		}
	}
	removeUnwanted(doc)

	var buf bytes.Buffer
	if err := golanghtml.Render(&buf, doc); err == nil {
		return buf.String()
	}
	return string(htmlBytes)
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
		Description: "Fetches a web page URL, cleans clutter with Readability, and converts the content into LLM-optimized Markdown. If a site blocks fetching (403/Cloudflare), use web_search instead.",
	}, fetchWebPage)
}

