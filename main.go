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

func chatLoop(client *internal.OpenRouterClient) {
	messages := []internal.Message{}
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
		messages = append(messages, internal.Message{Role: "user", Content: prompt})

		client.HandleRequest(messages, func(chunk string) {
			s, _, _ := internal.HandleOpenRouterChunk(chunk)
			fmt.Print(s)
			answer += s
		})

		messages = append(messages, internal.Message{Role: "assistant", Content: answer})
		fmt.Println()
		fmt.Println()
	}
}

func singleShot(client *internal.OpenRouterClient, stdin []byte) {
	prompt := strings.TrimSpace(string(stdin))

	client.HandleRequest([]internal.Message{{Role: "user", Content: prompt}}, func(chunk string) {
		s, _, _ := internal.HandleOpenRouterChunk(chunk)
		fmt.Print(s)
	})

	fmt.Println()
}

func main() {
	isTty := term.IsTerminal(int(os.Stdin.Fd()))
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
		if isTty {
			chatLoop(client)
		} else {
			stdin, err := io.ReadAll(os.Stdin)

			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			singleShot(client, stdin)
		}

		// exit routine
		stop()
	}()

	<-ctx.Done()
}
