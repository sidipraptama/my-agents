package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// openAICompatibleProvider talks to any server implementing OpenAI's
// /chat/completions contract — this is what a LiteLLM proxy speaks, and also
// covers self-hosted servers like vLLM or Ollama's OpenAI-compatible mode.
type openAICompatibleProvider struct {
	baseURL   string
	apiKey    string
	modelName string
	client    *http.Client
}

func NewOpenAICompatible(baseURL, apiKey, modelName string) Provider {
	return &openAICompatibleProvider{
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		apiKey:    apiKey,
		modelName: modelName,
		client:    &http.Client{Timeout: 60 * time.Second},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionsRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

type chatCompletionsResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (p *openAICompatibleProvider) Send(ctx context.Context, req Request) (string, error) {
	messages := make([]chatMessage, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		messages = append(messages, chatMessage{Role: "system", Content: req.SystemPrompt})
	}
	for _, m := range req.Messages {
		messages = append(messages, chatMessage{Role: m.Role, Content: m.Content})
	}

	body := chatCompletionsRequest{
		Model:     p.modelName,
		Messages:  messages,
		MaxTokens: req.MaxTokens,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("call litellm endpoint: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var parsed chatCompletionsResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if parsed.Error != nil {
		return "", fmt.Errorf("litellm error (%s): %s", parsed.Error.Type, parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("litellm endpoint returned status %d: %s", resp.StatusCode, string(raw))
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("litellm endpoint returned no choices")
	}

	return parsed.Choices[0].Message.Content, nil
}
