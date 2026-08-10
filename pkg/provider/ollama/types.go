package ollama

type OpenAIFunction struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

type OpenAITool struct {
	Type     string         `json:"type"`
	Function OpenAIFunction `json:"function"`
}

type OpenAIToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenAIToolCall struct {
	ID       string                 `json:"id,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Function OpenAIToolCallFunction `json:"function"`
}

type ChatMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	Name       string           `json:"name,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type ChatRequest struct {
	Model     string         `json:"model"`
	Messages  []ChatMessage  `json:"messages"`
	Tools     []OpenAITool   `json:"tools,omitempty"`
	Stream    bool           `json:"stream"`
	StreamOps *StreamOptions `json:"stream_options,omitempty"`
	Options   map[string]any `json:"options,omitempty"`
}

// UsageChunk holds token counts from Ollama usage response chunks.
type UsageChunk struct {
	PromptTokens  int `json:"prompt_tokens"`
	CompletionTok int `json:"completion_tokens"`
	TotalTokens   int `json:"total_tokens"`
}

type ChatStreamToolCallDelta struct {
	Index    int                    `json:"index"`
	ID       string                 `json:"id,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Function OpenAIToolCallFunction `json:"function"`
}

type ChatStreamDelta struct {
	Role      string                    `json:"role,omitempty"`
	Content   string                    `json:"content,omitempty"`
	Reasoning string                    `json:"reasoning,omitempty"`
	ToolCalls []ChatStreamToolCallDelta `json:"tool_calls,omitempty"`
}

type ChatStreamChoice struct {
	Index        int             `json:"index"`
	Delta        ChatStreamDelta `json:"delta"`
	FinishReason string          `json:"finish_reason,omitempty"`
}

type ChatStreamResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Choices []ChatStreamChoice `json:"choices"`
	Usage   *UsageChunk        `json:"usage,omitempty"`
}

type ChatChoiceMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	Reasoning string           `json:"reasoning"`
	ToolCalls []OpenAIToolCall `json:"tool_calls"`
}

type ChatChoice struct {
	Index        int               `json:"index"`
	Message      ChatChoiceMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   *UsageChunk  `json:"usage,omitempty"`
}

type ModelInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Object  string `json:"object,omitempty"`
	Created int64  `json:"created,omitempty"`
	OwnedBy string `json:"owned_by,omitempty"`
}

type ModelListResponse struct {
	Object string      `json:"object,omitempty"`
	Data   []ModelInfo `json:"data,omitempty"`
	Models []ModelInfo `json:"models,omitempty"`
}

type ShowModelRequest struct {
	Name  string `json:"name,omitempty"`
	Model string `json:"model,omitempty"`
}

type ShowModelResponse struct {
	Modelfile  string         `json:"modelfile,omitempty"`
	Parameters string         `json:"parameters,omitempty"`
	Template   string         `json:"template,omitempty"`
	Details    any            `json:"details,omitempty"`
	ModelInfo  map[string]any `json:"model_info,omitempty"`
}
