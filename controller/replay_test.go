package controller

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestParseReplay(t *testing.T) {
	actions, err := ParseReplay(strings.NewReader("\n# First action\nS\nWAIT 30s\nF5\n"))
	if err != nil {
		t.Fatalf("ParseReplay() error = %v", err)
	}

	if got, want := len(actions), 3; got != want {
		t.Fatalf("len(actions) = %d, want %d", got, want)
	}
	if got, want := actions[0].command, "S"; got != want {
		t.Errorf("first command = %q, want %q", got, want)
	}
	if got, want := actions[1].wait, 30*time.Second; got != want {
		t.Errorf("wait = %s, want %s", got, want)
	}
	if got, want := actions[2].line, 5; got != want {
		t.Errorf("last line = %d, want %d", got, want)
	}
}

func TestParseReplayInvalidWait(t *testing.T) {
	for _, input := range []string{"WAIT", "WAIT 0s", "WAIT -1s", "WAIT nope", "WAIT 1s extra"} {
		t.Run(input, func(t *testing.T) {
			_, err := ParseReplay(strings.NewReader(input))
			if err == nil || !strings.Contains(err.Error(), "line 1") {
				t.Errorf("ParseReplay() error = %v, want line-numbered error", err)
			}
		})
	}
}

func TestRunReplayWritesCommandsInOrder(t *testing.T) {
	actions := []ReplayAction{
		{line: 1, command: "S"},
		{line: 2, wait: time.Millisecond},
		{line: 3, command: "F5"},
	}
	var output bytes.Buffer
	var states []ReplayState

	err := RunReplay(context.Background(), actions, &output, func(state ReplayState) {
		states = append(states, state)
	})
	if err != nil {
		t.Fatalf("RunReplay() error = %v", err)
	}
	if got, want := output.String(), "S\nF5\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
	if got := states[len(states)-1]; got.Running || got.Cancelled || len(got.Queued) != 0 {
		t.Errorf("final state = %#v, want completed state", got)
	}
}

func TestRunReplayCancellationStopsWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	actions := []ReplayAction{{line: 1, wait: time.Hour}, {line: 2, command: "F5"}}
	states := make(chan ReplayState, 4)
	done := make(chan error, 1)

	go func() {
		done <- RunReplay(ctx, actions, &bytes.Buffer{}, func(state ReplayState) {
			states <- state
		})
	}()

	<-states // initial state
	<-states // active wait state
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("RunReplay() error = %v", err)
	}
	final := <-states
	if !final.Cancelled || final.Running || final.Current != "WAIT 1h0m0s" {
		t.Errorf("final state = %#v, want cancelled wait state", final)
	}
}
