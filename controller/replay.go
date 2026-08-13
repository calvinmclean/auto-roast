package controller

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type ReplayAction struct {
	line    int
	command string
	wait    time.Duration
}

func (a ReplayAction) String() string {
	if a.wait > 0 {
		return fmt.Sprintf("WAIT %s", a.wait)
	}
	return a.command
}

type ReplayState struct {
	Current   string
	Queued    []ReplayQueuedAction
	Started   bool
	Running   bool
	Cancelled bool
	WaitUntil time.Time
}

type ReplayQueuedAction struct {
	ID   int
	Text string
}

type replayItem struct {
	id     int
	action ReplayAction
}

type Replay struct {
	mu        sync.Mutex
	queued    []replayItem
	current   *replayItem
	started   bool
	running   bool
	cancelled bool
	waitUntil time.Time
	notify    func(ReplayState)
}

func NewReplay(actions []ReplayAction, notify func(ReplayState)) *Replay {
	r := &Replay{notify: notify}
	for id, action := range actions {
		r.queued = append(r.queued, replayItem{id: id, action: action})
	}
	return r
}

func (r *Replay) RemoveQueued(id int) bool {
	r.mu.Lock()
	for i, item := range r.queued {
		if item.id == id {
			r.queued = append(r.queued[:i], r.queued[i+1:]...)
			state := r.stateLocked()
			r.mu.Unlock()
			r.notify(state)
			return true
		}
	}
	r.mu.Unlock()
	return false
}

func (r *Replay) State() ReplayState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stateLocked()
}

func (r *Replay) stateLocked() ReplayState {
	state := ReplayState{Started: r.started, Running: r.running, Cancelled: r.cancelled, WaitUntil: r.waitUntil}
	if r.current != nil {
		state.Current = r.current.action.String()
	}
	for _, item := range r.queued {
		state.Queued = append(state.Queued, ReplayQueuedAction{ID: item.id, Text: item.action.String()})
	}
	return state
}

func (r *Replay) emit() {
	r.notify(r.State())
}

func LoadReplay(path string) ([]ReplayAction, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open replay file: %w", err)
	}
	defer f.Close()

	return ParseReplay(f)
}

func ParseReplay(r io.Reader) ([]ReplayAction, error) {
	scanner := bufio.NewScanner(r)
	var actions []ReplayAction
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if strings.EqualFold(fields[0], "WAIT") {
			if len(fields) != 2 {
				return nil, fmt.Errorf("line %d: WAIT requires exactly one duration", lineNumber)
			}
			duration, err := time.ParseDuration(fields[1])
			if err != nil || duration <= 0 {
				return nil, fmt.Errorf("line %d: invalid WAIT duration %q", lineNumber, fields[1])
			}
			actions = append(actions, ReplayAction{line: lineNumber, wait: duration})
			continue
		}

		actions = append(actions, ReplayAction{line: lineNumber, command: line})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read replay file: %w", err)
	}
	return actions, nil
}

func RunReplay(ctx context.Context, actions []ReplayAction, writer io.Writer, notify func(ReplayState)) error {
	return NewReplay(actions, notify).Run(ctx, writer)
}

func (r *Replay) Run(ctx context.Context, writer io.Writer) error {
	r.mu.Lock()
	r.started = true
	r.running = true
	r.cancelled = false
	r.mu.Unlock()
	r.emit()

	for {
		if ctx.Err() != nil {
			r.mu.Lock()
			r.running = false
			r.cancelled = true
			r.waitUntil = time.Time{}
			r.mu.Unlock()
			r.emit()
			return nil
		}

		r.mu.Lock()
		if len(r.queued) == 0 {
			r.current = nil
			r.running = false
			r.waitUntil = time.Time{}
			r.mu.Unlock()
			r.emit()
			return nil
		}
		item := r.queued[0]
		r.queued = r.queued[1:]
		r.current = &item
		if item.action.wait > 0 {
			r.waitUntil = time.Now().Add(item.action.wait)
		} else {
			r.waitUntil = time.Time{}
		}
		waitUntil := r.waitUntil
		r.mu.Unlock()
		r.emit()

		if item.action.wait > 0 {
			timer := time.NewTimer(time.Until(waitUntil))
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				continue
			case <-timer.C:
			}
			continue
		}

		if _, err := fmt.Fprintln(writer, item.action.command); err != nil {
			return fmt.Errorf("send replay command from line %d: %w", item.action.line, err)
		}
	}
}
