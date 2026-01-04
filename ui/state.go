package ui

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
		// Set initial values to F1P1, then Start and Preaheat
		return "I11\nS\nPREHEAT"
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
