package autoroast

const TerminationChar = 0x04 // ascii EOT (End of Transmission)

// ControlMode is the mode that the FreshRoast's display is showing
type ControlMode int

const (
	ControlModeUnknown ControlMode = iota
	ControlModeFan
	ControlModePower
	ControlModeTimer
)

func (cm ControlMode) String() string {
	switch cm {
	case ControlModeFan:
		return "Fan"
	case ControlModePower:
		return "Power"
	case ControlModeTimer:
		return "Timer"
	default:
		fallthrough
	case ControlModeUnknown:
		return "Unknown"
	}
}

// Next goes to the next mode on the FreshRoast display
func (cm ControlMode) Next() ControlMode {
	if cm == ControlModeTimer {
		return ControlModeFan
	}
	return cm + 1
}
