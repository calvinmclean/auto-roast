package ui

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func runCommands(reader io.Reader, writer io.Writer, handle func(string)) error {
	if closer, ok := writer.(io.Closer); ok {
		defer closer.Close()
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		command := strings.TrimSpace(scanner.Text())
		if command == "" {
			continue
		}

		handle(command)
		if _, err := fmt.Fprintln(writer, command); err != nil {
			return fmt.Errorf("forward command: %w", err)
		}
	}
	return scanner.Err()
}
