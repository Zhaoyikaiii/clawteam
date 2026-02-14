// Package llm provides internal LLM client integration
package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client wraps Anthropic API client
type Client struct {
	client anthropic.Client
	model  anthropic.Model
}

// NewClient creates a new LLM client
func NewClient(apiKey, model string) *Client {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &Client{
		client: client,
		model:  anthropic.Model(model),
	}
}

// Complete sends a completion request to the LLM
func (c *Client) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Build messages
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
	}

	// Create request - directly set required fields
	req := anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{{
			Text: systemPrompt,
		}},
		Messages: messages,
	}

	// Send request
	resp, err := c.client.Messages.New(ctx, req)
	if err != nil {
		return "", fmt.Errorf("LLM request failed: %w", err)
	}

	// Extract text from response
	if len(resp.Content) == 0 {
		return "", fmt.Errorf("empty response from LLM")
	}

	// Check content block type
	block := resp.Content[0]
	if block.Type != "text" {
		return "", fmt.Errorf("unexpected content type: %s", block.Type)
	}

	return block.Text, nil
}
