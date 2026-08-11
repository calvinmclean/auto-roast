package ui

import (
	"testing"
	"time"
)

func TestFormatWaitRemaining(t *testing.T) {
	for remaining, want := range map[time.Duration]string{
		0:                      "00:00",
		100 * time.Millisecond: "00:01",
		59 * time.Second:       "00:59",
		61 * time.Second:       "01:01",
		-1 * time.Second:       "00:00",
	} {
		if got := formatWaitRemaining(remaining); got != want {
			t.Errorf("formatWaitRemaining(%s) = %q, want %q", remaining, got, want)
		}
	}
}
