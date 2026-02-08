package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"mdgo/internal"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/term"
)

const apiURL = "https://openrouter.ai/api/v1/chat/completions"

type Payload struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func main() {
	fmt.Println(term.IsTerminal(int(os.Stdin.Fd())))

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Missing OPENROUTER_API_KEY")
		os.Exit(1)
	}

	client := internal.NewOpenRouterClient(ctx, apiKey)

	go func() {
		messages := []Message{}
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

			answer := ""
			messages = append(messages, Message{Role: "user", Content: prompt})
			payload := map[string]any{
				"model":    "openai/gpt-5",
				"stream":   true,
				"messages": messages,
			}

			err = client.HandleRequest(payload, func(chunk string) {
				s, _, _ := internal.HandleOpenRouterChunk(chunk)
				fmt.Print(s)
				answer += s
			})

			messages = append(messages, Message{Role: "assistant", Content: answer})
			fmt.Println()
			fmt.Println()
		}
	}()

	<-ctx.Done()
}
