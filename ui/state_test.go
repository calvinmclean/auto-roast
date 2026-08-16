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

func TestSettingValue(t *testing.T) {
	for command, want := range map[string]struct {
		prefix byte
		value  float64
		ok     bool
	}{
		"F5":   {'F', 5, true},
		"P4.5": {'P', 4.5, true},
		"F10":  {0, 0, false},
		"P0":   {0, 0, false},
		"FC":   {0, 0, false},
	} {
		t.Run(command, func(t *testing.T) {
			prefix, value, ok := settingValue(command)
			if prefix != want.prefix || value != want.value || ok != want.ok {
				t.Errorf("settingValue(%q) = (%q, %v, %t), want (%q, %v, %t)", command, prefix, value, ok, want.prefix, want.value, want.ok)
			}
		})
	}
}
