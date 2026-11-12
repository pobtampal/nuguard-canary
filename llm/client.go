// Package llm is a minimal chat-completion client, shaped like the common Go
// SDKs (a Config, NewClientWithConfig, CreateChatCompletion). It exists so
// the test program needs no third-party dependency. Nothing in this
// repository calls a real model.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Config configures a Client.
type Config struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// DefaultConfig returns a configuration for the given key.
func DefaultConfig(apiKey string) Config {
	return Config{APIKey: apiKey, BaseURL: "https://llm.invalid/v1", HTTPClient: http.DefaultClient}
}

// Client sends chat requests.
type Client struct {
	cfg Config
}

// NewClientWithConfig builds a Client.
func NewClientWithConfig(cfg Config) *Client {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	return &Client{cfg: cfg}
}

// Message is one chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is a chat-completion request.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// CreateChatCompletion sends the request and returns the reply text.
func (c *Client) CreateChatCompletion(ctx context.Context, req ChatRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	r.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	r.Header.Set("Content-Type", "application/json")
	resp, err := c.cfg.HTTPClient.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm: status %d", resp.StatusCode)
	}
	var out struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("llm: no choices")
	}
	return out.Choices[0].Message.Content, nil
}
