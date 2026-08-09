package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strings"

	"google.golang.org/adk/v2/model"
	"google.golang.org/genai"
)

type Model struct {
	modelName string
	baseURL   string
	apiKey    string
	client    *http.Client
}

func NewModel(modelName, baseURL, apiKey string) (*Model, error) {
	if modelName == "" {
		return nil, fmt.Errorf("modelName cannot be empty")
	}
	if baseURL == "" {
		baseURL = "https://ollama.dimensionlab.net/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL = baseURL + "/v1"
	}

	return &Model{
		modelName: modelName,
		baseURL:   baseURL,
		apiKey:    apiKey,
		client:    &http.Client{},
	}, nil
}

func (m *Model) Name() string {
	return m.modelName
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatStreamDelta struct {
	Role      string `json:"role,omitempty"`
	Content   string `json:"content,omitempty"`
	Reasoning string `json:"reasoning,omitempty"`
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
}

type ChatChoiceMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Reasoning string `json:"reasoning"`
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
}

func (m *Model) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		var messages []ChatMessage

		if req.Config != nil && req.Config.SystemInstruction != nil {
			var sysTexts []string
			for _, p := range req.Config.SystemInstruction.Parts {
				if p.Text != "" {
					sysTexts = append(sysTexts, p.Text)
				}
			}
			if len(sysTexts) > 0 {
				messages = append(messages, ChatMessage{
					Role:    "system",
					Content: strings.Join(sysTexts, "\n"),
				})
			}
		}

		for _, c := range req.Contents {
			role := c.Role
			if role == "model" || role == "assistant" {
				role = "assistant"
			} else if role == "" || role == "user" {
				role = "user"
			}

			var textParts []string
			for _, p := range c.Parts {
				if p.Text != "" {
					textParts = append(textParts, p.Text)
				}
			}
			if len(textParts) > 0 {
				messages = append(messages, ChatMessage{
					Role:    role,
					Content: strings.Join(textParts, "\n"),
				})
			}
		}

		payload := ChatRequest{
			Model:    m.modelName,
			Messages: messages,
			Stream:   stream,
		}

		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			yield(nil, fmt.Errorf("failed to marshal chat request: %w", err))
			return
		}

		url := m.baseURL + "/chat/completions"
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			yield(nil, fmt.Errorf("failed to create http request: %w", err))
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		if m.apiKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)
		}

		resp, err := m.client.Do(httpReq)
		if err != nil {
			yield(nil, fmt.Errorf("http request failed: %w", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBytes, _ := io.ReadAll(resp.Body)
			yield(nil, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(respBytes)))
			return
		}

		if !stream {
			var chatResp ChatResponse
			if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
				yield(nil, fmt.Errorf("failed to decode chat response: %w", err))
				return
			}
			if len(chatResp.Choices) == 0 {
				yield(nil, fmt.Errorf("empty choices in chat response"))
				return
			}

			msg := chatResp.Choices[0].Message
			customMeta := map[string]any{}
			if msg.Reasoning != "" {
				customMeta["reasoning"] = msg.Reasoning
			}

			llmResp := &model.LLMResponse{
				Content: &genai.Content{
					Role: "model",
					Parts: []*genai.Part{
						{Text: msg.Content},
					},
				},
				CustomMetadata: customMeta,
				TurnComplete:   true,
			}
			yield(llmResp, nil)
			return
		}

		reader := bufio.NewReader(resp.Body)
		var fullText strings.Builder

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				yield(nil, fmt.Errorf("error reading stream line: %w", err))
				return
			}

			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")
				if dataStr == "[DONE]" {
					break
				}

				var streamChunk ChatStreamResponse
				if err := json.Unmarshal([]byte(dataStr), &streamChunk); err != nil {
					continue
				}

				if len(streamChunk.Choices) > 0 {
					delta := streamChunk.Choices[0].Delta

					if delta.Reasoning != "" {
						llmResp := &model.LLMResponse{
							CustomMetadata: map[string]any{
								"reasoning": delta.Reasoning,
							},
							Partial: true,
						}
						if !yield(llmResp, nil) {
							return
						}
					}

					if delta.Content != "" {
						fullText.WriteString(delta.Content)
						llmResp := &model.LLMResponse{
							Content: &genai.Content{
								Role: "model",
								Parts: []*genai.Part{
									{Text: delta.Content},
								},
							},
							Partial: true,
						}
						if !yield(llmResp, nil) {
							return
						}
					}
				}
			}
		}

		finalResp := &model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
				Parts: []*genai.Part{
					{Text: fullText.String()},
				},
			},
			TurnComplete: true,
		}
		yield(finalResp, nil)
	}
}
