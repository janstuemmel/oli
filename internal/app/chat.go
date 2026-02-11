package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Delta represents the incremental content in a streaming response
type Delta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Choice represents a single response choice in the stream
type Choice struct {
	Index int   `json:"index"`
	Delta Delta `json:"delta"`
}

// Usage tracks API usage costs
type Usage struct {
	Cost float64
}

// Chunk represents a single SSE chunk from OpenRouter API
type Chunk struct {
	Choices   []Choice `json:"choices"`
	Usage     *Usage   `json:"usage"`
	Citations []string `json:"citations"`
}

// StreamHandler accumulates streaming responses and writes them to output
type StreamHandler struct {
	out       *StdinPipe
	answer    strings.Builder
	citations []string
}

func NewStreamHandler(out *StdinPipe) *StreamHandler {
	return &StreamHandler{out: out}
}

func (h *StreamHandler) HandleChunk(rawChunk string) error {

	if !strings.HasPrefix(rawChunk, "data:") || strings.HasSuffix(rawChunk, "[DONE]") {
		return nil
	}

	trimmed := strings.TrimSpace(strings.TrimPrefix(rawChunk, "data:"))
	var chunk Chunk
	if err := json.Unmarshal([]byte(trimmed), &chunk); err != nil {
		return fmt.Errorf("failed to unmarshal chunk: %w", err)
	}

	if len(chunk.Choices) == 0 {
		return nil
	}

	content := chunk.Choices[0].Delta.Content

	if content != "" {
		if err := h.out.Write(content); err != nil {
			return fmt.Errorf("failed to write content: %w", err)
		}
		h.answer.WriteString(content)
	}

	if len(chunk.Citations) > 0 {
		h.citations = chunk.Citations
	}

	return nil
}

func (h *StreamHandler) GetAnswer() string {
	return h.answer.String()
}

func (h *StreamHandler) GetCitations() []string {
	return h.citations
}

func getInitialMessages(system string) []Message {
	if system != "" {
		return []Message{{Role: "system", Content: system}}
	}
	return []Message{}
}

func ChatLoop(client *OpenRouterClient, config *Config) {
	messages := getInitialMessages(config.System)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stdout, "> ")
		line, err := reader.ReadString('\n')

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		prompt := strings.TrimSpace(line)

		if prompt == "" {
			continue
		}

		messages = append(messages, Message{Role: "user", Content: prompt})
		out := NewStdinPipe(config)
		handler := NewStreamHandler(out)

		err = client.HandleRequest(messages, func(chunk string) {
			if err := handler.HandleChunk(chunk); err != nil {
				fmt.Fprintf(os.Stderr, "chunk error: %v\n", err)
			}
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "request error: %v\n", err)
		}

		out.End()
		messages = append(messages, Message{Role: "assistant", Content: handler.GetAnswer()})
	}
}

func SingleShot(client *OpenRouterClient, config *Config, stdin []byte) {
	prompt := strings.TrimSpace(string(stdin))
	messages := getInitialMessages(config.System)
	messages = append(messages, Message{Role: "user", Content: prompt})
	out := NewStdinPipe(config)
	handler := NewStreamHandler(out)

	err := client.HandleRequest(messages, func(chunk string) {
		if err := handler.HandleChunk(chunk); err != nil {
			fmt.Fprintf(os.Stderr, "chunk error: %v\n", err)
		}
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "request error: %v\n", err)
	}

	citations := handler.GetCitations()
	if len(citations) > 0 {
		out.Write("\n\n")
		out.Write("Links:")
		out.Write("\n\n")
		out.Write(strings.Join(citations, "\n"))
	}

	out.End()
}
