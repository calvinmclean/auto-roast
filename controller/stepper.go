package controller

import (
	"errors"
	"machine"
	"time"
)

const defaultStepDelay = 2000 * time.Microsecond

type StepMode int

const (
	StepModeFull StepMode = iota
	StepModeHalf
)

type Stepper struct {
	pins        [4]machine.Pin
	stepMode    StepMode
	currentStep int
	stepDelay   time.Duration
}

func NewStepper(cfg StepperConfig) (*Stepper, error) {
	if cfg.StepMode != StepModeFull && cfg.StepMode != StepModeHalf {
		return nil, errors.New("invalid StepMode")
	}

	if cfg.StepDelay == 0 {
		cfg.StepDelay = defaultStepDelay
	}

	w := &Stepper{
		pins:        [4]machine.Pin{cfg.Pins[0], cfg.Pins[1], cfg.Pins[2], cfg.Pins[3]},
		stepMode:    cfg.StepMode,
		stepDelay:   cfg.StepDelay,
		currentStep: 0,
	}
	for _, p := range w.pins {
		p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	return w, nil
}

var (
	// 8-step half-step halfStepSequence
	halfStepSequence = [8][4]bool{
		{true, false, false, false},
		{true, true, false, false},
		{false, true, false, false},
		{false, true, true, false},
		{false, false, true, false},
		{false, false, true, true},
		{false, false, false, true},
		{true, false, false, true},
	}

	// 4-step sequence
	fullStepSequence = [4][4]bool{
		{true, false, false, false},
		{false, true, false, false},
		{false, false, true, false},
		{false, false, false, true},
	}
)

func (s *Stepper) applyStep() {
	var sequence [4]bool
	switch s.stepMode {
	default:
		fallthrough
	case StepModeFull:
		sequence = fullStepSequence[s.currentStep]
	case StepModeHalf:
		sequence = halfStepSequence[s.currentStep]
	}

	for i := range 4 {
		s.pins[i].Set(sequence[i])
	}
}

func (s *Stepper) StepForward() {
	sequenceLen := 4
	if s.stepMode == StepModeHalf {
		sequenceLen = 8
	}

	s.currentStep = (s.currentStep + 1) % sequenceLen
	s.applyStep()
	time.Sleep(s.stepDelay)
}

func (s *Stepper) StepBackward() {
	sequenceLen := 4
	if s.stepMode == StepModeHalf {
		sequenceLen = 8
	}

	s.currentStep = (s.currentStep - 1 + sequenceLen) % sequenceLen
	s.applyStep()
	time.Sleep(s.stepDelay)
}

func (s *Stepper) Move(steps int32) {
	if steps > 0 {
		for range steps {
			s.StepForward()
		}
	} else {
		for range -steps {
			s.StepBackward()
		}
	}
}
