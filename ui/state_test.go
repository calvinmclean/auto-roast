package ui

import "testing"

func TestStateForCommand(t *testing.T) {
	tests := map[string]state{
		"PREHEAT":  statePreheat,
		"ROAST":    stateRoasting,
		"ROASTING": stateRoasting,
		"FC":       stateFirstCrack,
		"CRACK":    stateFirstCrack,
		"COOL":     stateCooling,
		"DONE":     stateDone,
		"F5":       stateNone,
	}

	for command, want := range tests {
		t.Run(command, func(t *testing.T) {
			if got := stateForCommand(command); got != want {
				t.Errorf("stateForCommand(%q) = %v, want %v", command, got, want)
			}
		})
	}
}

func TestChangesSetting(t *testing.T) {
	for command, want := range map[string]bool{
		"F5":      true,
		"P9":      true,
		"f5":      false,
		"p9":      false,
		"FC":      false,
		"PREHEAT": false,
		"F0":      false,
	} {
		t.Run(command, func(t *testing.T) {
			if got := changesSetting(command); got != want {
				t.Errorf("changesSetting(%q) = %t, want %t", command, got, want)
			}
		})
	}
}
