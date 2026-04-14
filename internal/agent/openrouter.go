package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// OpenRouterModel implements the llms.Model interface using direct HTTP calls.
type OpenRouterModel struct {
	APIKey    string
	Model     string
	SessionID string
	Client    *http.Client
	Tools     []Tool
}

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
			Timeout: 120 * time.Second,
		},
		Tools: []Tool{},
	}
}

// AddWebSearchTool adds the OpenRouter web_search tool for Argentine domains
func (m *OpenRouterModel) AddWebSearchTool() {
	m.Tools = append(m.Tools, Tool{
		Type: "openrouter:web_search",
		Parameters: map[string]interface{}{
			"allowed_domains": []string{".com.ar", ".ar"},
			"max_results":     10,
		},
	})
}

// OpenRouterRequest represents the request body for the OpenRouter API.
type OpenRouterRequest struct {
	Model     string              `json:"model"`
	Messages  []OpenRouterMessage `json:"messages"`
	Tools     []Tool              `json:"tools,omitempty"`
	SessionID string              `json:"session_id,omitempty"`
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

// GenerateContent implements the llms.Model interface with tool support.
func (m *OpenRouterModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// Convert langchaingo messages to OpenRouter format
	var reqMessages []OpenRouterMessage
	for _, msg := range messages {
		var content string
		for _, part := range msg.Parts {
			if text, ok := part.(llms.TextContent); ok {
				content += text.Text
			}
		}
		reqMessages = append(reqMessages, OpenRouterMessage{
			Role:    string(msg.Role),
			Content: content,
		})
	}

	// Prepare the request payload
	payload := OpenRouterRequest{
		Model:     m.Model,
		Messages:  reqMessages,
		Tools:     m.Tools,
		SessionID: m.SessionID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.APIKey))
	// OpenRouter likes to know the app name for rankings
	req.Header.Set("HTTP-Referer", "https://pricenexus.ai")
	req.Header.Set("X-Title", "PriceNexus")

	// Execute the request
	resp, err := m.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenRouter API returned non-OK status: %d", resp.StatusCode)
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
	var contentStr string
	if content, ok := choice.Content.(string); ok {
		contentStr = content
	} else if contentArray, ok := choice.Content.([]interface{}); ok {
		// If content is an array, extract text parts
		for _, item := range contentArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if text, ok := itemMap["text"].(string); ok {
					contentStr += text
				}
			}
		}
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: contentStr,
			},
		},
	}, nil
}
