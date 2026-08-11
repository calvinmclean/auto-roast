package controller

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
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
	Queued    []string
	Running   bool
	Cancelled bool
	WaitUntil time.Time
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
	emit := func(current int, running, cancelled bool, waitUntil time.Time) {
		state := ReplayState{Running: running, Cancelled: cancelled, WaitUntil: waitUntil}
		if current >= 0 && current < len(actions) {
			state.Current = actions[current].String()
			for _, action := range actions[current+1:] {
				state.Queued = append(state.Queued, action.String())
			}
		} else if current < 0 {
			for _, action := range actions {
				state.Queued = append(state.Queued, action.String())
			}
		}
		notify(state)
	}

	emit(-1, true, false, time.Time{})
	for i, action := range actions {
		if err := ctx.Err(); err != nil {
			emit(i, false, true, time.Time{})
			return nil
		}

		if action.wait > 0 {
			waitUntil := time.Now().Add(action.wait)
			emit(i, true, false, waitUntil)
			timer := time.NewTimer(time.Until(waitUntil))
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				emit(i, false, true, time.Time{})
				return nil
			case <-timer.C:
			}
			continue
		}

		emit(i, true, false, time.Time{})

		if _, err := fmt.Fprintln(writer, action.command); err != nil {
			return fmt.Errorf("send replay command from line %d: %w", action.line, err)
		}
	}

	emit(len(actions), false, false, time.Time{})
	return nil
}
