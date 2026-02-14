package llm

import (
    "context"

    "github.com/anthropics/anthropic-go/v2"
)

// Client handles LLM API calls
type Client struct {
    apiKey string
}

// NewClient creates a new LLM client
func NewClient(apiKey string) *Client {
    return &Client{apiKey: apiKey}
}

// Complete sends a completion request to the LLM
func (c *Client) Complete(ctx context.Context, prompt string, systemPrompt string) (string, error) {
    client := anthropic.NewClient(apiKey)
    
    resp, err := client.Messages.Create(ctx, anthropic.MessageCreateParams{
        Model:     anthropic.ModelClaude3Dot5Sonnet20250214,
        MaxTokens: 4096,
        System:    systemPrompt,
        Messages: []anthropic.Message{
            {Role: anthropic.MessageRoleUser, Content: prompt},
        },
    })

    if err != nil {
        return "", err
    }

    return resp.Content[0].Text, nil
}
