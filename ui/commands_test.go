package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunCommandsAppliesAndForwardsCommands(t *testing.T) {
	var output bytes.Buffer
	var handled []string

	err := runCommands(strings.NewReader("\n S \nF5\n\nPREHEAT\n"), &output, func(command string) {
		handled = append(handled, command)
	})
	if err != nil {
		t.Fatalf("runCommands() error = %v", err)
	}
	if got, want := output.String(), "S\nF5\nPREHEAT\n"; got != want {
		t.Errorf("forwarded commands = %q, want %q", got, want)
	}
	if got, want := strings.Join(handled, ","), "S,F5,PREHEAT"; got != want {
		t.Errorf("handled commands = %q, want %q", got, want)
	}
}
