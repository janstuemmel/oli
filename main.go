package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const defaultBaseURL = "https://openrouter.ai/api/v1/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamChunk struct {
	Choices []struct {
		Delta        struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		Message      *Message `json:"message,omitempty"`
		FinishReason *string  `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func main() {
	apiKey := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	if apiKey == "" {
		exitErr("missing OPENROUTER_API_KEY")
	}

	model := strings.TrimSpace(os.Getenv("OPENROUTER_MODEL"))

	baseURL := strings.TrimSpace(os.Getenv("OPENROUTER_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	httpReferer := strings.TrimSpace(os.Getenv("OPENROUTER_HTTP_REFERER"))
	appTitle := strings.TrimSpace(os.Getenv("OPENROUTER_APP_TITLE"))

	client := &http.Client{Timeout: 0}
	reader := bufio.NewReader(os.Stdin)

	out, cleanup, err := setupStreamWriter()
	if err != nil {
		exitErr("setup stream output: " + err.Error())
	}
	if cleanup != nil {
		defer cleanup()
	}

	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant."},
	}

	fmt.Fprintln(os.Stderr, "OpenRouter chat. Type your message and press Enter. Commands: /exit, /reset")

	for {
		fmt.Fprint(os.Stderr, "> ")
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			exitErr("read input: " + err.Error())
		}
		line = strings.TrimSpace(line)

		if line == "" && errors.Is(err, io.EOF) {
			break
		}
		if line == "" {
			if errors.Is(err, io.EOF) {
				break
			}
			continue
		}
		if line == "/exit" || line == "/quit" {
			break
		}
		if line == "/reset" {
			messages = messages[:1]
			fmt.Println("History cleared.")
			continue
		}

		messages = append(messages, Message{Role: "user", Content: line})

		reply, err := streamChat(out, client, baseURL, apiKey, httpReferer, appTitle, model, messages)
		if err != nil {
			fmt.Fprintln(os.Stderr, "request error:", err)
			continue
		}
		fmt.Fprintln(out)
		if reply != "" {
			messages = append(messages, Message{Role: "assistant", Content: reply})
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}
}

func streamChat(out io.Writer, client *http.Client, url, apiKey, httpReferer, appTitle, model string, messages []Message) (string, error) {
	payload := map[string]any{
		"stream":   true,
		"messages": messages,
	}
	if model != "" {
		payload["model"] = model
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	if httpReferer != "" {
		req.Header.Set("HTTP-Referer", httpReferer)
	}
	if appTitle != "" {
		req.Header.Set("X-Title", appTitle)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var assistantReply strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.Error != nil {
			return "", errors.New(chunk.Error.Message)
		}
		if len(chunk.Choices) == 0 {
			continue
		}

		content := chunk.Choices[0].Delta.Content
		if content == "" && chunk.Choices[0].Message != nil {
			content = chunk.Choices[0].Message.Content
		}
		if content == "" {
			continue
		}

		if _, err := io.WriteString(out, content); err != nil {
			return "", err
		}
		assistantReply.WriteString(content)
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return assistantReply.String(), nil
}

func setupStreamWriter() (io.Writer, func() error, error) {
	pipeCmd := strings.TrimSpace(os.Getenv("OPENROUTER_PIPE"))
	if pipeCmd == "" {
		return os.Stdout, nil, nil
	}

	cmd := exec.Command("sh", "-c", pipeCmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		_ = stdin.Close()
		return cmd.Wait()
	}
	return stdin, cleanup, nil
}

func exitErr(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	fmt.Fprintln(os.Stderr, "Set OPENROUTER_API_KEY, and optionally OPENROUTER_MODEL/OPENROUTER_HTTP_REFERER/OPENROUTER_APP_TITLE.")
	os.Exit(1)
}
