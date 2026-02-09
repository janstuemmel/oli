package internal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func ChatLoop(client *OpenRouterClient) {
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
		out := NewStdinPipe()

		client.HandleRequest(messages, func(chunk string) {
			s, _, _ := HandleOpenRouterChunk(chunk)
			out.Write(s)
			answer += s
		})

		out.End()
		messages = append(messages, Message{Role: "assistant", Content: answer})
	}
}

func SingleShot(client *OpenRouterClient, stdin []byte) {
	prompt := strings.TrimSpace(string(stdin))
	out := NewStdinPipe()

	client.HandleRequest([]Message{{Role: "user", Content: prompt}}, func(chunk string) {
		s, _, _ := HandleOpenRouterChunk(chunk)
		out.Write(s)
	})

	out.End()
}
