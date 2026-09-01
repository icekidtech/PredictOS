package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = openai.GPT4oMini
	}
	return &OpenAIProvider{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

func (o *OpenAIProvider) Name() string { return "openai" }

func (o *OpenAIProvider) ParseStrategy(ctx context.Context, naturalLanguage string) (*ParseResult, error) {
	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: o.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: SystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: naturalLanguage},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai: no choices returned")
	}
	content := resp.Choices[0].Message.Content
	return parseAIJSON(content)
}

func parseAIJSON(content string) (*ParseResult, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON from AI: %w — raw: %s", err, content)
	}
	confRaw, hasConf := raw["_confidence"]
	warnRaw, hasWarn := raw["_warnings"]
	delete(raw, "_confidence")
	delete(raw, "_warnings")

	strategyBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	result := &ParseResult{Strategy: strategyBytes, Confidence: 0.8}
	if hasConf {
		var c float64
		if err := json.Unmarshal(confRaw, &c); err == nil {
			result.Confidence = c
		}
	}
	if hasWarn {
		var w []string
		if err := json.Unmarshal(warnRaw, &w); err == nil {
			result.Warnings = w
		}
	}
	return result, nil
}
