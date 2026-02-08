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

	"golang.org/x/term"
)

type sseChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type outputWriter struct {
	write func(string) error
	end   func() error
}

func createOutputWriter() (*outputWriter, error) {
	filter := strings.TrimSpace(os.Getenv("OPENROUTER_STREAM_FILTER"))
	if filter == "" {
		return &outputWriter{
			write: func(s string) error {
				_, err := io.WriteString(os.Stdout, s)
				return err
			},
			end: func() error { return nil },
		}, nil
	}

	cmd := exec.Command("sh", "-c", filter)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &outputWriter{
		write: func(s string) error {
			_, err := io.WriteString(stdin, s)
			return err
		},
		end: func() error {
			_ = stdin.Close()
			return cmd.Wait()
		},
	}, nil
}

func streamPrompt(prompt, model, apiKey string, out *outputWriter) error {
	payload := map[string]interface{}{
		"model":  model,
		"stream": true,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openrouter error %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		line = strings.TrimRight(line, "\r\n")
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var chunk sseChunk
		if json.Unmarshal([]byte(data), &chunk) != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta != "" {
			if err := out.write(delta); err != nil {
				return err
			}
		}
	}

	return nil
}

func readAllStdin() (string, error) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func example() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Missing OPENROUTER_API_KEY env var.")
		os.Exit(1)
	}
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "openai/gpt-4o-mini"
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		prompt, err := readAllStdin()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if prompt == "" {
			fmt.Fprintln(os.Stderr, "No prompt provided.")
			os.Exit(1)
		}
		out, err := createOutputWriter()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := streamPrompt(prompt, model, apiKey, out); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		_ = out.write("\n")
		if err := out.end(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

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
		if prompt == "/exit" || prompt == "/quit" {
			break
		}
		out, err := createOutputWriter()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := streamPrompt(prompt, model, apiKey, out); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		_ = out.write("\n")
		if err := out.end(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
