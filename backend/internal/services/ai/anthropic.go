package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AnthropicProvider calls the Anthropic Messages API directly (no SDK needed for MVP).
type AnthropicProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	return &AnthropicProvider{apiKey: apiKey, model: model, client: &http.Client{}}
}

func (a *AnthropicProvider) Name() string { return "anthropic" }

func (a *AnthropicProvider) ParseStrategy(ctx context.Context, naturalLanguage string) (*ParseResult, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"model":      a.model,
		"max_tokens": 4096,
		"system":     SystemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": naturalLanguage},
		},
	})
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("anthropic %d: %s", resp.StatusCode, string(data))
	}
	var parsed struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Content) == 0 {
		return nil, fmt.Errorf("anthropic: empty content")
	}
	return parseAIJSON(parsed.Content[0].Text)
}
