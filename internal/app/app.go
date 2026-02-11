package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func Run(config *Config) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {

		isTty := term.IsTerminal(int(os.Stdin.Fd()))

		ctx, stop := signal.NotifyContext(
			context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
		)

		client := NewOpenRouterClient(ctx, config)

		go func() {
			if isTty {
				prompt := strings.Join(args, " ")
				if prompt != "" {
					SingleShot(client, config, []byte(prompt))
				} else {
					ChatLoop(client, config)
				}
			} else {
				stdin, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				SingleShot(client, config, stdin)
			}

			// exit routine
			stop()
		}()

		<-ctx.Done()
	}
}
