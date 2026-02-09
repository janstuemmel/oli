package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/janstuemmel/oli/internal"
	"golang.org/x/term"
)

func getApiKey() string {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Missing OPENROUTER_API_KEY")
		os.Exit(1)
	}
	return apiKey
}

func main() {
	isTty := term.IsTerminal(int(os.Stdin.Fd()))
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	apiKey := getApiKey()
	model := "google/gemini-2.5-flash"
	client := internal.NewOpenRouterClient(ctx, apiKey, model)

	go func() {
		if isTty {
			internal.ChatLoop(client)
		} else {
			stdin, err := io.ReadAll(os.Stdin)

			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			internal.SingleShot(client, stdin)
		}

		// exit routine
		stop()
	}()

	<-ctx.Done()
}
