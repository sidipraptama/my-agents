// Package llm abstracts model providers so an Agent can bring its own model (BYOM).
package llm

import (
	"context"
	"fmt"
)

type Message struct {
	Role    string // "user" | "assistant"
	Content string
}

type Request struct {
	SystemPrompt string
	Messages     []Message
	MaxTokens    int
}

type Provider interface {
	Send(ctx context.Context, req Request) (string, error)
}

// New resolves a Provider from a stored model binding's provider name, API key, and config.
func New(provider, modelName, apiKey string, config map[string]any) (Provider, error) {
	switch provider {
	case "anthropic":
		return NewAnthropic(apiKey, modelName), nil
	case "litellm":
		baseURL, _ := config["base_url"].(string)
		if baseURL == "" {
			return nil, fmt.Errorf("base_url is required in the model binding config for provider %q", provider)
		}
		return NewOpenAICompatible(baseURL, apiKey, modelName), nil
	default:
		return nil, ErrUnsupportedProvider{Provider: provider}
	}
}

type ErrUnsupportedProvider struct {
	Provider string
}

func (e ErrUnsupportedProvider) Error() string {
	return "unsupported model provider: " + e.Provider
}
