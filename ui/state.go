package ui

import (
	"strconv"
	"strings"
)

type state int

const (
	stateNone state = iota
	statePreheat
	stateRoasting
	stateFirstCrack
	stateCooling
	stateDone
)

func (s state) String() string {
	switch s {
	case statePreheat:
		return "Preheat"
	case stateRoasting:
		return "Roasting"
	case stateFirstCrack:
		return "First Crack"
	case stateCooling:
		return "Cooling"
	case stateDone:
		return "Done"
	default:
		return "Unknown"
	}
}

func (s state) next() state {
	if s == stateDone {
		// Done has no next State
		return stateDone
	}
	return s + 1
}

func (s state) command() string {
	switch s {
	case statePreheat:
		return "S\nPREHEAT"
	case stateRoasting:
		return "ROASTING"
	case stateFirstCrack:
		return "FC"
	case stateCooling:
		return "COOL"
	case stateDone:
		return "DONE"
	default:
		return ""
	}
}

func stateForCommand(command string) state {
	switch strings.ToUpper(strings.TrimSpace(command)) {
	case "PREHEAT":
		return statePreheat
	case "ROAST", "ROASTING":
		return stateRoasting
	case "FC", "CRACK":
		return stateFirstCrack
	case "COOL":
		return stateCooling
	case "DONE":
		return stateDone
	default:
		return stateNone
	}
}

func settingValue(command string) (byte, float64, bool) {
	command = strings.TrimSpace(command)
	if len(command) < 2 || (command[0] != 'F' && command[0] != 'P') {
		return 0, 0, false
	}

	value, err := strconv.ParseFloat(command[1:], 64)
	if err != nil || value < 1 || value > 9 {
		return 0, 0, false
	}
	return command[0], value, true
}
