package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func getInitialMessages(system string) []Message {
	if system != "" {
		return []Message{{Role: "system", Content: system}}
	}
	return []Message{}
}

func ChatLoop(client *OpenRouterClient, config *Config) {
	messages := getInitialMessages("")
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
		out := NewStdinPipe(config)

		client.HandleRequest(messages, func(chunk string) {
			s, _, _ := HandleOpenRouterChunk(chunk)
			out.Write(s)
			answer += s
		})

		out.End()
		messages = append(messages, Message{Role: "assistant", Content: answer})
	}
}

func SingleShot(client *OpenRouterClient, config *Config, stdin []byte) {
	prompt := strings.TrimSpace(string(stdin))
	messages := getInitialMessages(config.System)
	messages = append(messages, Message{Role: "user", Content: prompt})
	out := NewStdinPipe(config)

	// TODO: handle citations in chunk handler
	var cit []string
	client.HandleRequest(messages, func(chunk string) {
		s, d, _ := HandleOpenRouterChunk(chunk)
		if d != nil && len(d.Citations) > 0 {
			cit = d.Citations
		}
		out.Write(s)
	})

	if len(cit) > 0 {
		out.Write("\n\n")
		out.Write("Links:")
		out.Write("\n\n")
		out.Write(strings.Join(cit, "\n"))
	}

	out.End()
}
