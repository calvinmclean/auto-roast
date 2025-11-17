package controller

import (
	"errors"
	"machine"
	"time"

	"tinygo.org/x/drivers/easystepper"
)

type Stepper interface {
	Move(int32)
}

func NewEasyStepper(cfg easystepper.DeviceConfig) (*easystepper.Device, error) {
	stepper, err := easystepper.New(cfg)
	if err != nil {
		return nil, errors.New("error creating stepper: " + err.Error())
	}
	stepper.Configure()
	return stepper, nil
}

type WorkingStepper struct {
	Pins [4]machine.Pin
}

func NewWorkingStepper(cfg easystepper.DeviceConfig) (*WorkingStepper, error) {
	w := &WorkingStepper{
		Pins: [4]machine.Pin{cfg.Pin1, cfg.Pin2, cfg.Pin3, cfg.Pin4},
	}
	for _, p := range w.Pins {
		p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	return w, nil
}

var (
	// // 8-step half-step sequence
	// sequence = [8][4]bool{
	// 	{true, false, false, false},
	// 	{true, true, false, false},
	// 	{false, true, false, false},
	// 	{false, true, true, false},
	// 	{false, false, true, false},
	// 	{false, false, true, true},
	// 	{false, false, false, true},
	// 	{true, false, false, true},
	// }

	// 4-step sequence
	sequence = [4][4]bool{
		{true, false, false, false},
		{false, true, false, false},
		{false, false, true, false},
		{false, false, false, true},
	}
)

func (s *WorkingStepper) Move(steps int32) {
	f := s.stepForward
	if steps < 0 {
		f = s.stepBackward
		steps = -steps
	}

	for range steps {
		f()
	}
}

func (s *WorkingStepper) step(idx int) {
	s.Pins[0].Set(sequence[idx][0])
	s.Pins[1].Set(sequence[idx][1])
	s.Pins[2].Set(sequence[idx][2])
	s.Pins[3].Set(sequence[idx][3])
}

func (s *WorkingStepper) stepForward() {
	for i := 0; i < len(sequence); i++ {
		s.step(i)
		time.Sleep(2 * time.Millisecond)
	}
}

func (s *WorkingStepper) stepBackward() {
	for i := len(sequence) - 1; i >= 0; i-- {
		s.step(i)
		time.Sleep(2 * time.Millisecond)
	}
}
