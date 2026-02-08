package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const url = "https://openrouter.ai/api/v1/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Payload struct {
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
	Messages []Message `json:"messages"`
}

type OpenRouterClient struct {
	apiKey string
	client *http.Client
	ctx    context.Context
}

func NewOpenRouterClient(ctx context.Context, apiKey string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey: apiKey,
		client: http.DefaultClient,
		ctx:    ctx,
	}
}

func (h *OpenRouterClient) HandleRequest(messages []Message, handle func(chunk string)) error {
	reqBody, err := json.Marshal(Payload{
		Model:    "google/gemini-2.5-flash",
		Stream:   true,
		Messages: messages,
	})

	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(h.ctx, "POST", url, bytes.NewReader(reqBody))

	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := h.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		if err := h.ctx.Err(); err != nil {
			return err
		}

		handle(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
