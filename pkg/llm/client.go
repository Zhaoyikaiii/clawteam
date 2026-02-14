// Package llm provides LLM client integration (Anthropic Claude)
package llm

import (
    "context"
    "fmt"

    "github.com/anthropics/anthropic-go/v2"
)

// Client wraps the Anthropic API client
type Client struct {
    client *anthropic.Client
    model  string
}

// NewClient creates a new LLM client
func NewClient(apiKey, model string) *Client {
    client := anthropic.NewClient(apiKey)
    return &Client{
        client: client,
        model:  model,
    }
}

// Complete sends a completion request to the LLM
func (c *Client) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
    message := anthropic.NewMessage(anthropic.MessageRoleUser, userPrompt)

    resp, err := c.client.Messages.Create(ctx, anthropic.MessageCreateParams{
        Model:     c.model,
        System:    anthropic.String(systemPrompt),
        MaxTokens: anthropic.Int(4096),
        Messages:  []anthropic.Message{message},
    })

    if err != nil {
        return "", err
    }

    if len(resp.Content) == 0 {
        return "", fmt.Errorf("empty response from LLM")
    }

    return resp.Content[0].Text, nil
}
