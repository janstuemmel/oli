package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/janstuemmel/oli/internal"
	"github.com/urfave/cli/v3"
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

func action(ctx context.Context, cmd *cli.Command) error {
	isTty := term.IsTerminal(int(os.Stdin.Fd()))
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	apiKey := getApiKey()
	model := cmd.String("model")
	client := internal.NewOpenRouterClient(ctx, apiKey, model)

	go func() {
		if isTty {
			prompt := strings.Join(cmd.StringArgs("prompt"), " ")
			if prompt != "" {
				internal.SingleShot(client, []byte(prompt))
			} else {
				internal.ChatLoop(client)
			}
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

	return nil
}

func main() {
	cmd := &cli.Command{
		Name:                  "oli",
		Action:                action,
		EnableShellCompletion: true,
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:      "prompt",
				UsageText: "your prompt",
				Min:       0,
				Max:       -1,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Usage:   "select model",
				Value:   "google/gemini-2.5-flash",
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

}
