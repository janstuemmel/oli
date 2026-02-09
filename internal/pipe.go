package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopWriteCloser{Writer: w}
}

type StdinPipe struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

func NewStdinPipe(config *Config) *StdinPipe {
	filter := strings.TrimSpace(config.Pipe)
	stdin := NopWriteCloser(os.Stdout)
	var cmd *exec.Cmd

	if filter != "" {
		cmd := exec.Command("sh", "-c", filter)
		cmd.Stdin = nil
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		stdin, err := cmd.StdinPipe()

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := cmd.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		return &StdinPipe{cmd: cmd, stdin: stdin}
	}

	return &StdinPipe{cmd: cmd, stdin: stdin}
}

func (s *StdinPipe) Write(str string) error {
	_, err := io.WriteString(s.stdin, str)
	return err
}

func (s *StdinPipe) End() error {
	s.Write("\n")
	s.stdin.Close()
	if s.cmd != nil {
		return s.cmd.Wait()
	}
	return nil
}
