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
		{name: "classic.roast", actions: 11, firstCommand: "S", firstWait: 45 * time.Second},
		{name: "short.roast", actions: 5, firstCommand: "S", firstWait: 15 * time.Second},
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

func TestLoadReplayWithNote(t *testing.T) {
	actions, err := LoadReplay(filepath.Join("testdata", "note.roast"))
	if err != nil {
		t.Fatalf("LoadReplay() error = %v", err)
	}
	var found bool
	for _, action := range actions {
		if action.command == "NOTE first crack approaching" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("actions = %v, want NOTE command", actions)
	}
}

func TestLoadReplayWithAlert(t *testing.T) {
	actions, err := LoadReplay(filepath.Join("testdata", "alert.roast"))
	if err != nil {
		t.Fatalf("LoadReplay() error = %v", err)
	}
	var found int
	for _, action := range actions {
		if action.alert != "" {
			found++
		}
	}
	if found != 2 {
		t.Errorf("alert actions = %d, want 2", found)
	}
}

func TestLoadReplayInvalidWait(t *testing.T) {
	_, err := LoadReplay(filepath.Join("testdata", "invalid-wait.roast"))
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

func TestReplayRemoveQueuedBeforeStart(t *testing.T) {
	var output bytes.Buffer
	replay := NewReplay([]ReplayAction{
		{line: 1, command: "S"},
		{line: 2, command: "F5"},
	}, func(ReplayState) {}, nil)

	if !replay.RemoveQueued(0) {
		t.Fatal("RemoveQueued() = false, want true")
	}
	if err := replay.Run(context.Background(), &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got, want := output.String(), "F5\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestReplayRemoveQueuedDuringWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var output bytes.Buffer
	states := make(chan ReplayState, 4)
	replay := NewReplay([]ReplayAction{
		{line: 1, wait: time.Hour},
		{line: 2, command: "F5"},
	}, func(state ReplayState) { states <- state }, nil)
	done := make(chan error, 1)

	go func() { done <- replay.Run(ctx, &output) }()
	<-states // initial queue
	active := <-states
	if active.Current != "WAIT 1h0m0s" {
		t.Fatalf("current = %q, want active wait", active.Current)
	}
	if replay.RemoveQueued(0) {
		t.Error("RemoveQueued() removed the active wait, want false")
	}
	if !replay.RemoveQueued(1) {
		t.Fatal("RemoveQueued() = false, want true")
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := output.String(); got != "" {
		t.Errorf("output = %q, want removed command not dispatched", got)
	}
}

func TestReplayAddAndMoveQueued(t *testing.T) {
	replay := NewReplay([]ReplayAction{{line: 1, command: "F5"}}, func(ReplayState) {}, nil)
	if err := replay.AddQueued("WAIT 30s"); err != nil {
		t.Fatalf("AddQueued() error = %v", err)
	}
	state := replay.State()
	if got, want := state.Queued[1].Text, "WAIT 30s"; got != want {
		t.Fatalf("second action = %q, want %q", got, want)
	}
	if !replay.MoveQueued(state.Queued[1].ID, -1) {
		t.Fatal("MoveQueued() = false, want true")
	}
	if got, want := replay.State().Queued[0].Text, "WAIT 30s"; got != want {
		t.Errorf("first action after move = %q, want %q", got, want)
	}
	if err := replay.AddQueued("WAIT never"); err == nil {
		t.Error("AddQueued() error = nil, want invalid wait error")
	}
}

func TestReplayMoveQueuedTo(t *testing.T) {
	replay := NewReplay([]ReplayAction{
		{line: 1, command: "F5"},
		{line: 2, command: "P5"},
		{line: 3, command: "F6"},
	}, func(ReplayState) {}, nil)
	if !replay.MoveQueuedTo(0, 2) {
		t.Fatal("MoveQueuedTo() = false, want true")
	}
	state := replay.State()
	if got, want := state.Queued[0].Text, "P5"; got != want {
		t.Errorf("first action = %q, want %q", got, want)
	}
	if got, want := state.Queued[1].Text, "F6"; got != want {
		t.Errorf("second action = %q, want %q", got, want)
	}
	if got, want := state.Queued[2].Text, "F5"; got != want {
		t.Errorf("third action = %q, want %q", got, want)
	}
	if replay.MoveQueuedTo(0, 3) {
		t.Error("MoveQueuedTo() = true out of bounds, want false")
	}
}

func TestReplaySkipWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var output bytes.Buffer
	states := make(chan ReplayState, 4)
	replay := NewReplay([]ReplayAction{
		{line: 1, wait: time.Hour},
		{line: 2, command: "F5"},
	}, func(state ReplayState) { states <- state }, nil)
	done := make(chan error, 1)

	go func() { done <- replay.Run(ctx, &output) }()
	<-states // initial queue
	<-states // active wait
	if !replay.Skip() {
		t.Fatal("Skip() = false, want true")
	}
	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := output.String(); got != "F5\n" {
		t.Errorf("output = %q, want %q", got, "F5\n")
	}
}

func TestReplaySkipCommandReturnsFalse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var output bytes.Buffer
	states := make(chan ReplayState, 4)
	replay := NewReplay([]ReplayAction{
		{line: 1, command: "S"},
	}, func(state ReplayState) { states <- state }, nil)
	done := make(chan error, 1)

	go func() { done <- replay.Run(ctx, &output) }()
	<-states // initial queue
	<-states // active command S
	if replay.Skip() {
		t.Error("Skip() = true, want false for command actions")
	}
	cancel()
	<-done
}

func TestReplaySkipWhenNotRunning(t *testing.T) {
	replay := NewReplay([]ReplayAction{{line: 1, wait: time.Hour}}, func(ReplayState) {}, nil)
	if replay.Skip() {
		t.Error("Skip() = true, want false when not running")
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

func TestParseReplayAlert(t *testing.T) {
	actions, err := ParseReplay(strings.NewReader("ALERT load beans"))
	if err != nil {
		t.Fatalf("ParseReplay() error = %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("len(actions) = %d, want 1", len(actions))
	}
	if got, want := actions[0].alert, "load beans"; got != want {
		t.Errorf("alert = %q, want %q", got, want)
	}
	if got, want := actions[0].String(), "ALERT load beans"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestParseReplayAlertRequiresMessage(t *testing.T) {
	_, err := ParseReplay(strings.NewReader("ALERT"))
	if err == nil || !strings.Contains(err.Error(), "line 1") {
		t.Errorf("ParseReplay() error = %v, want line-numbered error", err)
	}
}

func TestReplayAlertBlocksUntilDismissed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var output bytes.Buffer
	states := make(chan ReplayState, 4)
	dismissed := make(chan string, 1)
	dismiss := make(chan struct{})

	replay := NewReplay([]ReplayAction{
		{line: 1, alert: "load beans"},
		{line: 2, command: "F5"},
	}, func(state ReplayState) { states <- state }, func(message string) {
		dismissed <- message
		<-dismiss
	})

	done := make(chan error, 1)
	go func() { done <- replay.Run(ctx, &output) }()

	<-states // initial queue
	active := <-states
	if active.Current != "ALERT load beans" {
		t.Fatalf("current = %q, want alert state", active.Current)
	}

	msg := <-dismissed
	if msg != "load beans" {
		t.Errorf("alert message = %q, want %q", msg, "load beans")
	}

	select {
	case <-done:
		t.Fatal("Run() returned before alert was dismissed")
	default:
	}

	if output.String() != "" {
		t.Errorf("output = %q before dismiss, want empty", output.String())
	}

	close(dismiss)
	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if got, want := output.String(), "F5\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}
