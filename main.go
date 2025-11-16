package main

import (
	"errors"
	"machine"
	"time"

	"tinygo.org/x/drivers/easystepper"
	"tinygo.org/x/drivers/servo"
)

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

// State controls the state of the device. It manages the Stepper and Servo motors. It also tracks the current
// state of the FreshRoast.
type State struct {
	stepper        Stepper
	servo          servo.Servo
	calibrationCfg CalibrationConfig

	currentControlMode ControlMode
	fan                uint
	power              uint

	startTime time.Time

	// lastClick is used to track the last click of the FreshRoast button. This is important because it lets us
	// know if the first click will enable changing currentControlMode or will increment the ControlMode
	lastClick time.Time

	// lastDirection tells if the stepper was previously moving forwards or backwards. It is useful
	// for backlash compensation. -1, 0, +1
	lastDirection int

	verbose bool
}

// ServoConfig has device-level values for setting up the Servo
type ServoConfig struct {
	Pin machine.Pin
	PWM servo.PWM
}

// CalibrationConfig has values for the moving parts that depend on positioning and motor specifics
type CalibrationConfig struct {
	ServoBasePosition  int
	ServoClickPosition int
	ServoPressDelay    time.Duration
	ServoResetDelay    time.Duration
	StepsPerIncrement  uint
	BacklashSteps      uint
}

// NewState intializes the state with the provided configs
func NewState(stepperCfg easystepper.DeviceConfig, servoCfg ServoConfig, calibrationCfg CalibrationConfig) (State, error) {
	stepper, err := NewWorkingStepper(stepperCfg)
	if err != nil {
		return State{}, errors.New("error creating stepper: " + err.Error())
	}

	var myServo servo.Servo
	if servoCfg != (ServoConfig{}) {
		myServo, err = servo.New(servoCfg.PWM, servoCfg.Pin)
		if err != nil {
			return State{}, errors.New("error creating servo: " + err.Error())
		}
		err := myServo.SetAngle(calibrationCfg.ServoBasePosition)
		if err != nil {
			return State{}, errors.New("error setting servo angle: " + err.Error())
		}
	}

	return State{
		stepper:            stepper,
		servo:              myServo,
		calibrationCfg:     calibrationCfg,
		currentControlMode: ControlModeFan,
		fan:                0,
		power:              0,
		startTime:          time.Time{},
		lastClick:          time.Time{},
		verbose:            false,
	}, nil
}

// Start will set the start time and ensures that everything is ready to go
func (s *State) Start() error {
	if s.fan == 0 || s.power == 0 {
		return errors.New("set initial fan/power before starting")
	}
	s.startTime = time.Now()

	println("Started...")

	return nil
}

// Duration returns the duration that this has been running
func (s *State) Duration() time.Duration {
	return time.Since(s.startTime)
}

// ClickButton uses the servo motor to click the FreshRoast button to enable setting changes
func (s *State) ClickButton() ControlMode {
	if s.verbose {
		println("ClickButton")
	}

	err := s.servo.SetAngle(s.calibrationCfg.ServoClickPosition)
	if err != nil {
		println("error setting servo angle:", err.Error())
		return s.currentControlMode
	}

	time.Sleep(s.calibrationCfg.ServoPressDelay)

	err = s.servo.SetAngle(s.calibrationCfg.ServoBasePosition)
	if err != nil {
		println("error resetting servo angle:", err.Error())
		return s.currentControlMode
	}

	time.Sleep(s.calibrationCfg.ServoResetDelay)

	s.currentControlMode = s.currentControlMode.Next()
	return s.currentControlMode
}

// GoToMode will click the FreshRoast button until the target ControlMode is active
func (s *State) GoToMode(target ControlMode) {
	if s.verbose {
		println("GoToMode:", target)
	}

	if target == ControlModeUnknown {
		return
	}
	for s.currentControlMode != target {
		_ = s.ClickButton()
	}
}

// FixControlMode manually sets the ControlMode to account for errors
func (s *State) FixControlMode(cm ControlMode) {
	s.currentControlMode = cm
}

// MoveFan controls the FreshRoast to move the fan value by the specified number of increments.
// It does not change the "State" of the device. This is useful for fixing off-by-one movements
func (s *State) MoveFan(i int32) {
	if s.verbose {
		println("MoveFan", i)
	}
	s.GoToMode(ControlModeFan)
	s.move(i)
}

// MovePower controls the FreshRoast to move the power value by the specified number of increments.
// It does not change the "State" of the device. This is useful for fixing off-by-one movements
func (s *State) MovePower(i int32) {
	if s.verbose {
		println("MovePower", i)
	}
	s.GoToMode(ControlModePower)
	s.move(i)
}

// MoveTimer controls the FreshRoast to move the timer value by the specified number of increments.
// If this exceeds the bounds, it will still move by the number of increments.
func (s *State) MoveTimer(i int32) {
	if s.verbose {
		println("MoveTimer", i)
	}
	s.GoToMode(ControlModeTimer)
	s.move(i)
}

// FixPower manually sets the current power to the specified value to account for errors. It does not control the device
func (s *State) FixPower(p uint) {
	s.power = p
}

// FixFan manually sets the current fan to the specified value to account for errors. It does not control the device
func (s *State) FixFan(f uint) {
	s.fan = f
}

// SetFan sets the FreshRoast fan to the specified value
func (s *State) SetFan(f uint) {
	if s.verbose {
		println("SetFan", f)
	}
	if f < 1 || f > 9 {
		return
	}

	s.fan = uint(f)

	// Only actually move if started. This allows setting base values
	if !s.startTime.IsZero() {
		delta := f - s.fan
		s.MoveFan(int32(delta))
	} else if s.verbose {
		println("Not moving since this has not started")
	}
}

// SetPower sets the FreshRoast power to the specified value
func (s *State) SetPower(p uint) {
	if s.verbose {
		println("SetPower", p)
	}
	if p < 1 || p > 9 {
		return
	}

	s.power = p

	// Only actually move if started. This allows setting base values
	if !s.startTime.IsZero() {
		delta := p - s.power
		s.MovePower(int32(delta))
	} else if s.verbose {
		println("Not moving since this has not started")
	}
}

// move simply moves the stepper by the specified number of increments
func (s *State) move(n int32) {
	move := n * int32(s.calibrationCfg.StepsPerIncrement)

	// add or subtract backlash steps based on direction change
	if s.lastDirection < 0 && move > 0 {
		move += int32(s.calibrationCfg.BacklashSteps)
		s.lastDirection = +1
	} else if s.lastDirection > 0 && move < 0 {
		move -= int32(s.calibrationCfg.BacklashSteps)
		s.lastDirection = -1
	}

	s.stepper.Move(move)
}

func main() {
	stepperCfg := easystepper.DeviceConfig{
		Pin1: machine.GP0, Pin2: machine.GP1, Pin3: machine.GP2, Pin4: machine.GP3,
		// Pin1: machine.D8, Pin2: machine.D9, Pin3: machine.D11, Pin4: machine.D12,
		// StepCount: 200,
		// RPM:       50,
		// Mode:      easystepper.ModeFour,
	}

	servoCfg := ServoConfig{
		// PWM: machine.Timer1,
		// Pin: machine.D10,
		PWM: machine.PWM2,
		Pin: machine.GP4,
	}
	calibrationCfg := CalibrationConfig{
		ServoBasePosition:  10,
		ServoClickPosition: 65,
		ServoPressDelay:    250 * time.Millisecond,
		ServoResetDelay:    250 * time.Millisecond,
		// StepsPerIncrement:  61,
		StepsPerIncrement: 28,
		BacklashSteps:     2,
	}

	state, err := NewState(stepperCfg, servoCfg, calibrationCfg)
	if err != nil {
		panic(err)
	}

	RunCommands(&state)
}
