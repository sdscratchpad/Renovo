// Package llm provides a client for the GitHub Models GPT-4o inference endpoint.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	defaultEndpoint = "https://models.inference.ai.azure.com/chat/completions"
	defaultModel    = "gpt-4o"
)

// Client calls the GitHub Models chat completions API.
type Client struct {
	endpoint string
	model    string
	token    string
	http     *http.Client
}

// New creates a Client authenticated with the given GitHub token.
func New(token string) *Client {
	return &Client{
		endpoint: defaultEndpoint,
		model:    defaultModel,
		token:    token,
		http:     &http.Client{Timeout: 60 * time.Second},
	}
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}

// Complete sends a system + user prompt to GPT-4o and returns the assistant reply.
func (c *Client) Complete(ctx context.Context, system, user string) (string, error) {
	reqBody := chatRequest{
		Model: c.model,
		Messages: []message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.2,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("llm: marshal request: %w", err)
	}

	if os.Getenv("DEBUG_LLM") != "" {
		log.Printf("llm: REQUEST\n=== system ===\n%s\n=== user ===\n%s\n", system, user)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("llm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm: http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("llm: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm: status %d: %s", resp.StatusCode, string(body))
	}

	if os.Getenv("DEBUG_LLM") != "" {
		log.Printf("llm: RESPONSE status=%d\n%s\n", resp.StatusCode, string(body))
	}

	var cr chatResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return "", fmt.Errorf("llm: parse response: %w", err)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("llm: no choices in response")
	}
	return cr.Choices[0].Message.Content, nil
}
