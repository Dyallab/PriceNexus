package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// OpenRouterModel implements the llms.Model interface using direct HTTP calls.
type OpenRouterModel struct {
	APIKey     string
	Model      string
	SessionID  string
	Client     *http.Client
	Tools      []Tool
	MaxRetries int
	RetryDelay time.Duration
}

const (
	DefaultMaxRetries = 3
	DefaultRetryDelay = 2 * time.Second
)

// Tool represents a tool that can be used by the model
type Tool struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// NewOpenRouterModel creates a new instance of OpenRouterModel.
func NewOpenRouterModel(apiKey, model string) *OpenRouterModel {
	return &OpenRouterModel{
		APIKey: apiKey,
		Model:  model,
		Client: &http.Client{
			Timeout: 300 * time.Second,
		},
		Tools:      []Tool{},
		MaxRetries: DefaultMaxRetries,
		RetryDelay: DefaultRetryDelay,
	}
}

// AddWebSearchTool adds the OpenRouter web_search tool using runtime configuration.
func (m *OpenRouterModel) AddWebSearchTool(config SearchConfig) {
	if len(config.AllowedDomains) == 0 || len(config.ExcludedDomains) == 0 || config.SearchEngine == "" || config.MaxResults <= 0 {
		defaultConfig := DefaultSearchConfig()
		if len(config.AllowedDomains) == 0 {
			config.AllowedDomains = defaultConfig.AllowedDomains
		}
		if len(config.ExcludedDomains) == 0 {
			config.ExcludedDomains = defaultConfig.ExcludedDomains
		}
		if strings.TrimSpace(config.SearchEngine) == "" {
			config.SearchEngine = defaultConfig.SearchEngine
		}
		if config.MaxResults <= 0 {
			config.MaxResults = defaultConfig.MaxResults
		}
	}

	m.Tools = append(m.Tools, Tool{
		Type: "openrouter:web_search",
		Parameters: map[string]interface{}{
			"engine":           config.SearchEngine,
			"allowed_domains":  config.AllowedDomains,
			"excluded_domains": config.ExcludedDomains,
			"max_results":      config.MaxResults,
		},
	})
}

// OpenRouterRequest represents the request body for the OpenRouter API.
type OpenRouterRequest struct {
	Model          string                    `json:"model"`
	Messages       []OpenRouterMessage       `json:"messages"`
	Tools          []Tool                    `json:"tools,omitempty"`
	SessionID      string                    `json:"session_id,omitempty"`
	ResponseFormat *OpenRouterResponseFormat `json:"response_format,omitempty"`
}

type OpenRouterResponseFormat struct {
	Type       string                `json:"type"`
	JSONSchema *OpenRouterJSONSchema `json:"json_schema,omitempty"`
}

type OpenRouterJSONSchema struct {
	Name   string         `json:"name"`
	Strict bool           `json:"strict,omitempty"`
	Schema map[string]any `json:"schema"`
}

// OpenRouterMessage represents a single message in the chat history.
type OpenRouterMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// OpenRouterResponse represents the response body from the OpenRouter API.
type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Role      string                   `json:"role"`
			Content   interface{}              `json:"content"`
			ToolCalls []map[string]interface{} `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Call implements the llms.Model interface.
func (m *OpenRouterModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}
	resp, err := m.GenerateContent(ctx, messages, options...)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from model")
	}
	return resp.Choices[0].Content, nil
}

func (m *OpenRouterModel) CallWithJSONSchema(
	ctx context.Context,
	prompt string,
	schemaName string,
	schema map[string]any,
) (string, error) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	resp, err := m.generateContentWithRequest(
		ctx,
		messages,
		&OpenRouterResponseFormat{
			Type: "json_schema",
			JSONSchema: &OpenRouterJSONSchema{
				Name:   schemaName,
				Strict: true,
				Schema: schema,
			},
		},
	)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from model")
	}

	return resp.Choices[0].Content, nil
}

// GenerateContent implements the llms.Model interface with tool support.
func (m *OpenRouterModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return m.generateContentWithRequest(ctx, messages, nil)
}

func (m *OpenRouterModel) generateContentWithRequest(
	ctx context.Context,
	messages []llms.MessageContent,
	responseFormat *OpenRouterResponseFormat,
) (*llms.ContentResponse, error) {
	// Convert langchaingo messages to OpenRouter format
	reqMessages := make([]OpenRouterMessage, 0, len(messages))
	for _, msg := range messages {
		var contentBuilder strings.Builder
		for _, part := range msg.Parts {
			if text, ok := part.(llms.TextContent); ok {
				contentBuilder.WriteString(text.Text)
			}
		}
		reqMessages = append(reqMessages, OpenRouterMessage{
			Role:    mapOpenRouterRole(msg.Role),
			Content: contentBuilder.String(),
		})
	}

	// Prepare the request payload
	payload := OpenRouterRequest{
		Model:          m.Model,
		Messages:       reqMessages,
		Tools:          m.Tools,
		SessionID:      m.SessionID,
		ResponseFormat: responseFormat,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Execute the request with retry logic for transient errors
	var resp *http.Response

	for attempt := 0; attempt <= m.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(
			ctx,
			"POST",
			"https://openrouter.ai/api/v1/chat/completions",
			bytes.NewReader(body),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.APIKey))
		req.Header.Set("HTTP-Referer", "https://pricenexus.ai")
		req.Header.Set("X-Title", "PriceNexus")

		resp, err = m.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("HTTP request failed: %w", err)
		}

		// Success or non-retryable error
		if resp.StatusCode != http.StatusInternalServerError &&
			resp.StatusCode != http.StatusServiceUnavailable &&
			resp.StatusCode != http.StatusGatewayTimeout {
			break
		}

		// Close body before retry
		resp.Body.Close()

		// Don't sleep after the last attempt
		if attempt < m.MaxRetries {
			delay := m.RetryDelay * time.Duration(1<<attempt) // Exponential backoff
			time.Sleep(delay)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("OpenRouter API returned non-OK status: %d", resp.StatusCode)
		}

		trimmedBody := strings.TrimSpace(string(responseBody))
		if trimmedBody == "" {
			return nil, fmt.Errorf("OpenRouter API returned non-OK status: %d", resp.StatusCode)
		}

		return nil, fmt.Errorf(
			"OpenRouter API returned non-OK status: %d: %s",
			resp.StatusCode,
			trimmedBody,
		)
	}

	// Decode the response
	var openRouterResp OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if openRouterResp.Error.Message != "" {
		return nil, fmt.Errorf("OpenRouter API error: %s", openRouterResp.Error.Message)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("OpenRouter API returned no choices")
	}

	// Map the result back to langchaingo's ContentResponse
	choice := openRouterResp.Choices[0].Message

	// Extract content - handle both string and array formats
	contentStr := extractOpenRouterContent(choice.Content)

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: contentStr,
			},
		},
	}, nil
}

func extractOpenRouterContent(content any) string {
	switch value := content.(type) {
	case string:
		return value
	case []interface{}:
		var builder strings.Builder
		for _, item := range value {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			if text, ok := itemMap["text"].(string); ok {
				builder.WriteString(text)
			}

			if nestedText, ok := itemMap["content"].(string); ok {
				builder.WriteString(nestedText)
			}
		}
		return builder.String()
	default:
		return ""
	}
}

func mapOpenRouterRole(role llms.ChatMessageType) string {
	switch strings.ToLower(string(role)) {
	case "system":
		return "system"
	case "ai", "assistant":
		return "assistant"
	case "human", "user":
		return "user"
	case "tool":
		return "tool"
	case "function":
		return "function"
	default:
		return string(role)
	}
}
