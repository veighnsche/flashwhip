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
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/adk/v2/model"
	"google.golang.org/genai"

	"flashwhip/pkg/errors"
	fnet "flashwhip/pkg/net"
)

type Model struct {
	modelName string
	baseURL   string
	apiKey    string
	client    *http.Client
	adapter   ModelAdapter
	usage     *Usage
	ctxLength int
}

func NewModel(modelName, baseURL, apiKey string) (*Model, error) {
	if modelName == "" {
		return nil, errors.New(errors.ErrCodeProviderEmptyModel, "modelName cannot be empty")
	}
	if baseURL == "" {
		baseURL = "https://ollama.dimensionlab.net/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL = baseURL + "/v1"
	}

	adapter := SelectAdapter(modelName)
	ctxLen, _ := FetchModelContextLength(baseURL, apiKey, modelName)
	if ctxLen <= 0 {
		ctxLen = 32768
	}

	return &Model{
		modelName: modelName,
		baseURL:   baseURL,
		apiKey:    apiKey,
		client:    fnet.DefaultHTTPClient(),
		adapter:   adapter,
		usage:     NewUsage(),
		ctxLength: ctxLen,
	}, nil
}

// Usage returns the cumulative token usage tracker for this model session.
func (m *Model) Usage() *Usage {
	return m.usage
}

// ContextLength returns the discovered max context window length in tokens.
func (m *Model) ContextLength() int {
	if m.ctxLength <= 0 {
		return 32768
	}
	return m.ctxLength
}

// Name returns the model name identifier.
func (m *Model) Name() string {
	return m.modelName
}

// FetchAvailableModels queries the target endpoint (e.g. GET baseURL/models) to list available models.
func FetchAvailableModels(baseURL, apiKey string) ([]string, error) {
	if baseURL == "" {
		baseURL = "https://ollama.dimensionlab.net/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	modelsURL := baseURL + "/models"
	client := fnet.DefaultHTTPClient()

	req, err := http.NewRequest("GET", modelsURL, nil)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeModelFetchFailed, "failed to create request", err)
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(errors.ErrCodeModelFetchFailed, err, "failed to connect to model endpoint %q", modelsURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf(errors.ErrCodeModelFetchFailed, "endpoint returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var listResp ModelListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, errors.Wrap(errors.ErrCodeModelFetchFailed, "failed to decode models response", err)
	}

	var modelNames []string
	seen := make(map[string]bool)

	for _, m := range listResp.Data {
		name := m.ID
		if name == "" {
			name = m.Name
		}
		if name != "" && !seen[name] {
			seen[name] = true
			modelNames = append(modelNames, name)
		}
	}

	for _, m := range listResp.Models {
		name := m.ID
		if name == "" {
			name = m.Name
		}
		if name != "" && !seen[name] {
			seen[name] = true
			modelNames = append(modelNames, name)
		}
	}

	if len(modelNames) == 0 {
		return nil, errors.New(errors.ErrCodeProviderNoModelsFound, "no models returned from endpoint")
	}

	return modelNames, nil
}

// FetchModelContextLength queries the Ollama endpoint (e.g. POST hostURL/api/show) to discover the model's max token context length.
// Returns 32768 as a default fallback if context length is not specified or endpoint is unsupported.
func FetchModelContextLength(baseURL, apiKey, modelName string) (int, error) {
	if baseURL == "" {
		baseURL = "https://ollama.dimensionlab.net/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	// Strip trailing /v1 to reach Ollama native API root
	hostURL := strings.TrimSuffix(baseURL, "/v1")

	showURL := hostURL + "/api/show"
	client := fnet.DefaultHTTPClient()

	reqPayload := ShowModelRequest{
		Name:  modelName,
		Model: modelName,
	}
	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return 32768, err
	}

	req, err := http.NewRequest("POST", showURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return 32768, err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 32768, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 32768, errors.Errorf(errors.ErrCodeProviderHTTPStatus, "show endpoint returned status %d", resp.StatusCode)
	}

	var showResp ShowModelResponse
	if err := json.NewDecoder(resp.Body).Decode(&showResp); err != nil {
		return 32768, err
	}

	// 1. Parse parameters string for num_ctx <number> (explicit modelfile config)
	if showResp.Parameters != "" {
		re := regexp.MustCompile(`(?i)num_ctx\s+(\d+)`)
		matches := re.FindStringSubmatch(showResp.Parameters)
		if len(matches) > 1 {
			if n, err := strconv.Atoi(matches[1]); err == nil && n > 0 {
				return n, nil
			}
		}
	}

	// 2. Inspect model_info map for *.context_length entries (cap default at 32768)
	if showResp.ModelInfo != nil {
		for k, v := range showResp.ModelInfo {
			if strings.HasSuffix(k, ".context_length") {
				var valInt int
				switch val := v.(type) {
				case float64:
					valInt = int(val)
				case int:
					valInt = val
				}
				if valInt > 0 {
					if valInt > 32768 {
						return 32768, nil
					}
					return valInt, nil
				}
			}
		}
	}

	return 32768, nil
}

func sanitizeUTF8(s string) string {
	s = strings.ToValidUTF8(s, "")
	return strings.Map(func(r rune) rune {
		if r == 0 || (r < 32 && r != '\n' && r != '\r' && r != '\t') {
			return -1
		}
		return r
	}, s)
}

func (m *Model) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		messages, err := m.adapter.BuildMessages(req)
		if err != nil {
			yield(nil, errors.Wrap(errors.ErrCodeProviderMessageBuildFailed, "failed to build messages", err))
			return
		}

		// Convert ADK tools into OpenAI tool specifications
		var openAITools []OpenAITool
		if req.Config != nil && len(req.Config.Tools) > 0 {
			for _, t := range req.Config.Tools {
				if t == nil {
					continue
				}
				for _, decl := range t.FunctionDeclarations {
					if decl == nil {
						continue
					}
					params := decl.ParametersJsonSchema
					if params == nil {
						params = decl.Parameters
					}
					if params == nil {
						params = map[string]any{
							"type":       "object",
							"properties": map[string]any{},
						}
					}
					openAITools = append(openAITools, OpenAITool{
						Type: "function",
						Function: OpenAIFunction{
							Name:        decl.Name,
							Description: decl.Description,
							Parameters:  params,
						},
					})
				}
			}
		}

		payload := ChatRequest{
			Model:    m.modelName,
			Messages: messages,
			Tools:    openAITools,
			Stream:   stream,
			Options:  map[string]any{"num_ctx": m.ctxLength},
		}
		if stream {
			payload.StreamOps = &StreamOptions{IncludeUsage: true}
		}

		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			yield(nil, errors.Wrap(errors.ErrCodeProviderMarshalFailed, "failed to marshal chat request", err))
			return
		}

		if os.Getenv("DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "\n[DEBUG Ollama Request Payload]: %s\n\n", string(bodyBytes))
		}

		url := m.baseURL + "/chat/completions"
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			yield(nil, errors.Wrap(errors.ErrCodeNetHTTPClientFailed, "failed to create http request", err))
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		if m.apiKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)
		}

		resp, err := m.client.Do(httpReq)
		if err != nil {
			yield(nil, errors.Wrap(errors.ErrCodeProviderRequestFailed, "http request failed", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBytes, _ := io.ReadAll(resp.Body)
			yield(nil, errors.Errorf(errors.ErrCodeProviderHTTPStatus, "Ollama API returned status %d: %s", resp.StatusCode, string(respBytes)))
			return
		}

		if !stream {
			var chatResp ChatResponse
			if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
				yield(nil, errors.Wrap(errors.ErrCodeProviderResponseDecodeFailed, "failed to decode chat response", err))
				return
			}
			if chatResp.Usage != nil {
				m.usage.Record(chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTok, chatResp.Usage.TotalTokens)
			}
			if len(chatResp.Choices) == 0 {
				yield(nil, errors.New(errors.ErrCodeProviderEmptyResponse, "empty choices in chat response"))
				return
			}

			msg := chatResp.Choices[0].Message
			var parts []*genai.Part

			if msg.Reasoning != "" {
				parts = append(parts, &genai.Part{
					Text:    msg.Reasoning,
					Thought: true,
				})
			}

			if msg.Content != "" {
				parts = append(parts, &genai.Part{
					Text:    msg.Content,
					Thought: false,
				})
			}

			for _, tc := range msg.ToolCalls {
				var argsMap map[string]any
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &argsMap)
				if argsMap == nil {
					argsMap = map[string]any{}
				}
				parts = append(parts, &genai.Part{
					FunctionCall: &genai.FunctionCall{
						Name: tc.Function.Name,
						Args: argsMap,
					},
				})
			}

			llmResp := &model.LLMResponse{
				Content: &genai.Content{
					Role:  "model",
					Parts: parts,
				},
				TurnComplete: true,
			}
			yield(llmResp, nil)
			return
		}

		reader := bufio.NewReader(resp.Body)
		state := NewStreamState()

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				yield(nil, errors.Wrap(errors.ErrCodeStreamReadFailed, "error reading stream line", err))
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

				// Track token usage from Ollama's special usage chunk
				if streamChunk.Usage != nil {
					m.usage.Record(streamChunk.Usage.PromptTokens, streamChunk.Usage.CompletionTok, streamChunk.Usage.TotalTokens)
				}

				if len(streamChunk.Choices) > 0 {
					delta := streamChunk.Choices[0].Delta
					deltaParts := m.adapter.ProcessStreamDelta(delta, state)
					for _, p := range deltaParts {
						llmResp := &model.LLMResponse{
							Content: &genai.Content{
								Role:  "model",
								Parts: []*genai.Part{p},
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

		var turnParts []*genai.Part

		if state.Reasoning.Len() > 0 {
			turnParts = append(turnParts, &genai.Part{
				Text:    state.Reasoning.String(),
				Thought: true,
			})
		}

		cleanContent, xmlToolParts := extractXMLToolCalls(state.Content.String())

		if cleanContent != "" {
			turnParts = append(turnParts, &genai.Part{
				Text:    cleanContent,
				Thought: false,
			})
		}

		var indices []int
		for idx := range state.ToolCalls {
			indices = append(indices, idx)
		}
		sort.Ints(indices)

		for _, idx := range indices {
			b := state.ToolCalls[idx]
			var argsMap map[string]any
			argsStr := b.Args.String()
			if argsStr != "" {
				_ = json.Unmarshal([]byte(argsStr), &argsMap)
			}
			if argsMap == nil {
				argsMap = map[string]any{}
			}
			toolPart := &genai.Part{
				FunctionCall: &genai.FunctionCall{
					Name: b.Name,
					Args: argsMap,
				},
			}
			turnParts = append(turnParts, toolPart)

			// Yield the complete function call to the runner stream
			llmResp := &model.LLMResponse{
				Content: &genai.Content{
					Role:  "model",
					Parts: []*genai.Part{toolPart},
				},
				Partial: true,
			}
			if !yield(llmResp, nil) {
				return
			}
		}

		for _, toolPart := range xmlToolParts {
			turnParts = append(turnParts, toolPart)

			// Yield fallback XML tool call to the runner stream
			llmResp := &model.LLMResponse{
				Content: &genai.Content{
					Role:  "model",
					Parts: []*genai.Part{toolPart},
				},
				Partial: true,
			}
			if !yield(llmResp, nil) {
				return
			}
		}

		finalResp := &model.LLMResponse{
			Content: &genai.Content{
				Role:  "model",
				Parts: turnParts,
			},
			TurnComplete: true,
		}
		yield(finalResp, nil)
	}
}


