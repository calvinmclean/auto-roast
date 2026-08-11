package controller

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadReplay(t *testing.T) {
	tests := []struct {
		name         string
		actions      int
		firstCommand string
		firstWait    time.Duration
	}{
		{name: "classic-roast.txt", actions: 11, firstCommand: "S", firstWait: 45 * time.Second},
		{name: "short-roast.txt", actions: 5, firstCommand: "S", firstWait: 15 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions, err := LoadReplay(filepath.Join("testdata", tt.name))
			if err != nil {
				t.Fatalf("LoadReplay() error = %v", err)
			}
			if got := len(actions); got != tt.actions {
				t.Errorf("len(actions) = %d, want %d", got, tt.actions)
			}
			if got := actions[0].command; got != tt.firstCommand {
				t.Errorf("first command = %q, want %q", got, tt.firstCommand)
			}
			var firstWait time.Duration
			for _, action := range actions {
				if action.wait > 0 {
					firstWait = action.wait
					break
				}
			}
			if got := firstWait; got != tt.firstWait {
				t.Errorf("first wait = %s, want %s", got, tt.firstWait)
			}
		})
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

func TestLoadReplayInvalidWait(t *testing.T) {
	_, err := LoadReplay(filepath.Join("testdata", "invalid-wait-roast.txt"))
	if err == nil || !strings.Contains(err.Error(), "line 3") {
		t.Errorf("LoadReplay() error = %v, want line-numbered error", err)
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

	<-states           // initial state
	active := <-states // active wait state
	if active.WaitUntil.Before(time.Now().Add(59 * time.Minute)) {
		t.Errorf("WaitUntil = %v, want roughly one hour from now", active.WaitUntil)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("RunReplay() error = %v", err)
	}
	final := <-states
	if !final.Cancelled || final.Running || final.Current != "WAIT 1h0m0s" {
		t.Errorf("final state = %#v, want cancelled wait state", final)
	}
	if !final.WaitUntil.IsZero() {
		t.Errorf("WaitUntil = %v after cancellation, want zero", final.WaitUntil)
	}
}
