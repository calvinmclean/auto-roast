package ui

import (
	"testing"
	"time"
)

func TestAdjustmentTimerIgnoresMilestones(t *testing.T) {
	timer := newTimer(true)
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	for _, command := range []string{"PREHEAT", "ROASTING", "FC"} {
		timer.setForAdjustment(command, now)
	}
	if !timer.startTime.IsZero() {
		t.Fatal("milestone started the adjustment timer before any adjustment")
	}

	setting, value, ok := timer.setForAdjustment("P5", now)
	if !ok || setting != 'P' || value != 5 || !timer.startTime.Equal(now) {
		t.Fatal("power command did not start the adjustment timer")
	}
	for _, command := range []string{"FC", "COOL", "DONE", "NOTE first crack", "P10", "f5"} {
		timer.setForAdjustment(command, now.Add(time.Minute))
		if !timer.startTime.Equal(now) {
			t.Fatalf("%q reset the adjustment timer", command)
		}
	}

	later := now.Add(2 * time.Minute)
	setting, value, ok = timer.setForAdjustment("F6", later)
	if !ok || setting != 'F' || value != 6 || !timer.startTime.Equal(later) {
		t.Fatal("fan command did not reset the adjustment timer")
	}
}
