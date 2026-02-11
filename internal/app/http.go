package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const url = "https://openrouter.ai/api/v1/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Plugin = struct {
	Id string `json:"id"`
}

type Payload struct {
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
	Messages []Message `json:"messages"`
	Plugins  []Plugin  `json:"plugins"`
}

type OpenRouterClient struct {
	config *Config
	client *http.Client
	ctx    context.Context
}

func NewOpenRouterClient(ctx context.Context, config *Config) *OpenRouterClient {
	return &OpenRouterClient{
		config: config,
		client: http.DefaultClient,
		ctx:    ctx,
	}
}

func (h *OpenRouterClient) HandleRequest(messages []Message, handle func(chunk string)) error {
	plugins := []Plugin{}
	if h.config.Online {
		plugins = append(plugins, Plugin{Id: "web"})
	}

	reqBody, err := json.Marshal(Payload{
		Model:    h.config.Model,
		Stream:   true,
		Messages: messages,
		Plugins:  plugins,
	})

	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(h.ctx, "POST", url, bytes.NewReader(reqBody))

	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.config.Apikey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := h.client.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "Client returned with: %s\n", resp.Status)
		os.Exit(1)
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
