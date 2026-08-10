package ollama

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/adk/v2/model"
	"google.golang.org/genai"

	"flashwhip/pkg/agent/middleware"
)

// XMLStreamFilter filters out <tool_call>...</tool_call> blocks during real-time SSE streaming.
type XMLStreamFilter struct {
	buf strings.Builder
}

func NewXMLStreamFilter() *XMLStreamFilter {
	return &XMLStreamFilter{}
}

// Feed receives incoming text delta chunks and returns text parts that are safe to yield immediately.
func (f *XMLStreamFilter) Feed(deltaText string) string {
	if deltaText == "" {
		return ""
	}
	f.buf.WriteString(deltaText)
	return f.process()
}

func (f *XMLStreamFilter) process() string {
	s := f.buf.String()
	if s == "" {
		return ""
	}

	var output strings.Builder

	for len(s) > 0 {
		toolCallIdx := strings.Index(s, "<tool_call")
		if toolCallIdx == -1 {
			prefixLen := matchToolCallPrefix(s)
			if prefixLen > 0 {
				safeText := s[:len(s)-prefixLen]
				output.WriteString(safeText)
				f.buf.Reset()
				f.buf.WriteString(s[len(s)-prefixLen:])
				return output.String()
			}

			output.WriteString(s)
			f.buf.Reset()
			return output.String()
		}

		if toolCallIdx > 0 {
			output.WriteString(s[:toolCallIdx])
			s = s[toolCallIdx:]
		}

		endIdx := strings.Index(s, "</tool_call>")
		if endIdx != -1 {
			s = s[endIdx+len("</tool_call>"):]
		} else {
			f.buf.Reset()
			f.buf.WriteString(s)
			return output.String()
		}
	}

	f.buf.Reset()
	return output.String()
}

func (f *XMLStreamFilter) Flush() string {
	s := f.buf.String()
	f.buf.Reset()
	if idx := strings.Index(s, "<tool_call"); idx != -1 {
		return s[:idx]
	}
	return s
}

func matchToolCallPrefix(s string) int {
	target := "<tool_call"
	maxCheck := len(target) - 1
	if len(s) < maxCheck {
		maxCheck = len(s)
	}
	for l := maxCheck; l >= 1; l-- {
		if strings.HasSuffix(s, target[:l]) {
			return l
		}
	}
	return 0
}

// StreamState holds accumulated token state during streaming.
type StreamState struct {
	Reasoning strings.Builder
	Content   strings.Builder
	ToolCalls map[int]*ToolCallBuilder
	xmlFilter *XMLStreamFilter
}

type ToolCallBuilder struct {
	ID   string
	Name string
	Args strings.Builder
}

func NewStreamState() *StreamState {
	return &StreamState{
		ToolCalls: make(map[int]*ToolCallBuilder),
		xmlFilter: NewXMLStreamFilter(),
	}
}

// ModelAdapter abstracts payload construction and delta parsing for different model families.
type ModelAdapter interface {
	BuildMessages(req *model.LLMRequest) ([]ChatMessage, error)
	ProcessStreamDelta(delta ChatStreamDelta, state *StreamState) []*genai.Part
}

// SelectAdapter chooses the optimal adapter based on the model identifier.
func SelectAdapter(modelName string) ModelAdapter {
	lower := strings.ToLower(modelName)
	if strings.Contains(lower, "qwen") || strings.Contains(lower, "deepseek") || strings.Contains(lower, "kat-coder") {
		return &QwenDeepSeekAdapter{}
	}
	return &OpenAIStandardAdapter{}
}

// OpenAIStandardAdapter implements the standard OpenAI /v1 specification.
type OpenAIStandardAdapter struct{}

func (a *OpenAIStandardAdapter) BuildMessages(req *model.LLMRequest) ([]ChatMessage, error) {
	var messages []ChatMessage

	if req.Config != nil && req.Config.SystemInstruction != nil {
		var sysTexts []string
		for _, p := range req.Config.SystemInstruction.Parts {
			if p.Text != "" {
				sysTexts = append(sysTexts, sanitizeUTF8(p.Text))
			}
		}
		if len(sysTexts) > 0 {
			messages = append(messages, ChatMessage{
				Role:    "system",
				Content: strings.Join(sysTexts, "\n"),
			})
		}
	}

	pendingToolCalls := make(map[string][]string)
	var globalCallCounter int

	contents := middleware.PruneContents(req.Contents, 3)

	for _, c := range contents {
		role := c.Role
		if role == "model" || role == "assistant" {
			role = "assistant"
		} else if role == "" || role == "user" {
			role = "user"
		}

		var textParts []string
		var toolCalls []OpenAIToolCall
		var toolResponses []ChatMessage

		for _, p := range c.Parts {
			if p.Text != "" {
				textParts = append(textParts, sanitizeUTF8(p.Text))
			}
			if p.FunctionCall != nil {
				globalCallCounter++
				callID := fmt.Sprintf("call_%d_%s", globalCallCounter, p.FunctionCall.Name)
				pendingToolCalls[p.FunctionCall.Name] = append(pendingToolCalls[p.FunctionCall.Name], callID)

				var argsBytes []byte
				if p.FunctionCall.Args != nil {
					argsBytes, _ = json.Marshal(p.FunctionCall.Args)
				}
				if len(argsBytes) == 0 || string(argsBytes) == "null" {
					argsBytes = []byte("{}")
				}
				toolCalls = append(toolCalls, OpenAIToolCall{
					ID:   callID,
					Type: "function",
					Function: OpenAIToolCallFunction{
						Name:      p.FunctionCall.Name,
						Arguments: sanitizeUTF8(string(argsBytes)),
					},
				})
			}
			if p.FunctionResponse != nil {
				var callID string
				queue := pendingToolCalls[p.FunctionResponse.Name]
				if len(queue) > 0 {
					callID = queue[0]
					pendingToolCalls[p.FunctionResponse.Name] = queue[1:]
				} else {
					globalCallCounter++
					callID = fmt.Sprintf("call_%d_%s", globalCallCounter, p.FunctionResponse.Name)
				}

				var respBytes []byte
				if p.FunctionResponse.Response != nil {
					respBytes, _ = json.Marshal(p.FunctionResponse.Response)
				}
				if len(respBytes) == 0 || string(respBytes) == "null" {
					respBytes = []byte("{}")
				}
				toolResponses = append(toolResponses, ChatMessage{
					Role:       "tool",
					ToolCallID: callID,
					Name:       p.FunctionResponse.Name,
					Content:    sanitizeUTF8(string(respBytes)),
				})
			}
		}

		if len(textParts) > 0 || len(toolCalls) > 0 {
			messages = append(messages, ChatMessage{
				Role:      role,
				Content:   strings.Join(textParts, "\n"),
				ToolCalls: toolCalls,
			})
		}

		if len(toolResponses) > 0 {
			messages = append(messages, toolResponses...)
		}
	}

	return messages, nil
}

func (a *OpenAIStandardAdapter) ProcessStreamDelta(delta ChatStreamDelta, state *StreamState) []*genai.Part {
	var parts []*genai.Part

	if delta.Reasoning != "" {
		state.Reasoning.WriteString(delta.Reasoning)
		parts = append(parts, &genai.Part{
			Text:    delta.Reasoning,
			Thought: true,
		})
	}

	if delta.Content != "" {
		state.Content.WriteString(delta.Content)
		safeText := state.xmlFilter.Feed(delta.Content)
		if safeText != "" {
			parts = append(parts, &genai.Part{
				Text:    safeText,
				Thought: false,
			})
		}
	}

	for _, tc := range delta.ToolCalls {
		idx := tc.Index
		b, ok := state.ToolCalls[idx]
		if !ok {
			b = &ToolCallBuilder{}
			state.ToolCalls[idx] = b
		}
		if tc.ID != "" {
			b.ID = tc.ID
		}
		if tc.Function.Name != "" {
			b.Name = tc.Function.Name
		}
		if tc.Function.Arguments != "" {
			b.Args.WriteString(tc.Function.Arguments)
		}
	}

	return parts
}

var (
	toolCallRegex = regexp.MustCompile(`(?s)<tool_call>\s*<function=([^>]+)>\s*(.*?)\s*</function>\s*</tool_call>`)
	paramRegex    = regexp.MustCompile(`(?s)<parameter=([^>]+)>\s*(.*?)\s*</parameter>`)
)

func parseXMLParamValue(val string) any {
	trimmed := strings.TrimSpace(val)
	if strings.EqualFold(trimmed, "true") {
		return true
	}
	if strings.EqualFold(trimmed, "false") {
		return false
	}
	if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return f
	}
	var jsonVal any
	if err := json.Unmarshal([]byte(trimmed), &jsonVal); err == nil {
		return jsonVal
	}
	return trimmed
}

func extractXMLToolCalls(content string) (string, []*genai.Part) {
	matches := toolCallRegex.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return content, nil
	}

	var toolCallParts []*genai.Part
	var sb strings.Builder
	lastIdx := 0

	for _, m := range matches {
		sb.WriteString(content[lastIdx:m[0]])
		lastIdx = m[1]

		funcName := strings.TrimSpace(content[m[2]:m[3]])
		paramBlock := content[m[4]:m[5]]

		argsMap := make(map[string]any)
		paramMatches := paramRegex.FindAllStringSubmatch(paramBlock, -1)
		for _, pm := range paramMatches {
			pName := strings.TrimSpace(pm[1])
			pVal := strings.TrimSpace(pm[2])
			argsMap[pName] = parseXMLParamValue(pVal)
		}

		toolCallParts = append(toolCallParts, &genai.Part{
			FunctionCall: &genai.FunctionCall{
				Name: funcName,
				Args: argsMap,
			},
		})
	}
	sb.WriteString(content[lastIdx:])
	cleanedContent := strings.TrimSpace(sb.String())
	return cleanedContent, toolCallParts
}

// QwenDeepSeekAdapter handles reasoning and tool formatting for Qwen/DeepSeek family models.
type QwenDeepSeekAdapter struct {
	OpenAIStandardAdapter
}
