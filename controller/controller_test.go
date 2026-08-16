package controller

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/calvinmclean/autoroast/twchart"
)

type recordingTWChartClient struct {
	events []string
}

func (r *recordingTWChartClient) CreateSession(ctx context.Context, beanName string, probes twchart.Probes) (string, error) {
	return "", nil
}

func (r *recordingTWChartClient) SetStartTime(ctx context.Context, startTime time.Time) error {
	return nil
}

func (r *recordingTWChartClient) AddEvent(ctx context.Context, note string, now time.Time) error {
	r.events = append(r.events, note)
	return nil
}

func (r *recordingTWChartClient) AddStage(ctx context.Context, name string, now time.Time) error {
	return nil
}

func (r *recordingTWChartClient) Done(ctx context.Context) error {
	return nil
}

func TestControllerSkipsTWChartEventsAfterDone(t *testing.T) {
	mock := &recordingTWChartClient{}
	c := &Controller{
		config: Config{
			SessionName: "test",
		},
		twchartClient: mock,
		port:          &mockPort{},
	}

	var output bytes.Buffer
	input := strings.NewReader("F5\nDONE\nF6\nP7\n")
	if err := c.Run(context.Background(), input, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := []string{"F5"}
	if len(mock.events) != len(want) {
		t.Fatalf("AddEvent calls = %v, want %v", mock.events, want)
	}
	for i, event := range want {
		if mock.events[i] != event {
			t.Errorf("event[%d] = %q, want %q", i, mock.events[i], event)
		}
	}
}

func TestControllerStillPassthroughsCommandsAfterDone(t *testing.T) {
	mock := &recordingTWChartClient{}
	port := &mockPort{}
	c := &Controller{
		config: Config{
			SessionName: "test",
		},
		twchartClient: mock,
		port:          port,
	}

	var output bytes.Buffer
	input := strings.NewReader("DONE\nF6\n")
	if err := c.Run(context.Background(), input, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := []string{"F6"}
	if !equalStrings(port.commands, want) {
		t.Errorf("port commands = %q, want %q", port.commands, want)
	}
}
