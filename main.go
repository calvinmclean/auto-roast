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
	stepper        *easystepper.Device
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
}

// ServoConfig has device-level values for setting up the Servo
type ServoConfig struct {
	Pin machine.Pin
	PWM servo.PWM
}

// CalibrationConfig has values for the moving parts that depend on positioning and motor specifics
type CalibrationConfig struct {
	ServoBasePosition  uint
	ServoClickPosition uint
	StepsPerIncrement  uint
	BacklashSteps      uint
}

// NewState intializes the state with the provided configs
func NewState(stepperCfg easystepper.DeviceConfig, servoCfg ServoConfig, calibrationCfg CalibrationConfig) (State, error) {
	stepper, err := easystepper.New(stepperCfg)
	if err != nil {
		return State{}, errors.New("error creating stepper: " + err.Error())
	}
	stepper.Configure()

	var myServo servo.Servo
	if servoCfg != (ServoConfig{}) {
		array, err := servo.NewArray(servoCfg.PWM)
		if err != nil {
			return State{}, errors.New("error creating servo array: " + err.Error())
		}

		myServo, err = array.Add(servoCfg.Pin)
		if err != nil {
			return State{}, errors.New("error adding servo to array: " + err.Error())
		}
	}

	return State{
		stepper:            stepper,
		servo:              myServo,
		calibrationCfg:     calibrationCfg,
		currentControlMode: ControlModeFan,
		fan:                1,
		power:              1,
		startTime:          time.Now(),
		lastClick:          time.Now(),
	}, nil
}

// ClickButton uses the servo motor to click the FreshRoast button to enable setting changes
func (s *State) ClickButton() ControlMode {
	// TODO: implement
	s.currentControlMode = s.currentControlMode.Next()
	return s.currentControlMode
}

// GoToMode will click the FreshRoast button until the target ControlMode is active
func (s *State) GoToMode(target ControlMode) {
	// TODO: Implement
}

// FixControlMode manually sets the ControlMode to account for errors
func (s *State) FixControlMode(cm ControlMode) {
	s.currentControlMode = cm
}

// MoveFan controls the FreshRoast to move the fan value by the specified number of increments and returns the state.
// If this exceeds the bounds, it will still move by the number of increments, but still returns the actual current value.
func (s *State) MoveFan(i int) uint {
	println("MoveFan", i)
	// TODO: GoToMode(Fan)
	// TODO: move correct number of increments
	return s.fan
}

// MovePower controls the FreshRoast to move the power value by the specified number of increments and returns the state.
// If this exceeds the bounds, it will still move by the number of increments, but still returns the actual current value.
func (s *State) MovePower(i int) uint {
	// TODO: GoToMode(Power)
	// TODO: move correct number of increments
	return s.power
}

// MoveTimer controls the FreshRoast to move the timer value by the specified number of increments.
// If this exceeds the bounds, it will still move by the number of increments.
func (s *State) MoveTimer(i int) {
	// TODO: GoToMode(Timer)
	// TODO: move correct number of increments
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
	println("SetFan", f)
	// TODO: calculate diff and call MoveFan
}

// SetPower sets the FreshRoast power to the specified value
func (s *State) SetPower(p uint) {
	// TODO: calculate diff and call MovePower
}

func main() {
	stepperCfg := easystepper.DeviceConfig{
		Pin1: machine.D8, Pin2: machine.D9, Pin3: machine.D10, Pin4: machine.D11,
		StepCount: 200,
		RPM:       50,
		Mode:      easystepper.ModeFour,
	}
	servoCfg := ServoConfig{
		// PWM: machine.Timer2,
		// Pin: machine.D6,
	}
	calibrationCfg := CalibrationConfig{
		ServoBasePosition:  10,
		ServoClickPosition: 30,
		StepsPerIncrement:  61,
		BacklashSteps:      2,
	}

	state, err := NewState(stepperCfg, servoCfg, calibrationCfg)
	if err != nil {
		panic(err)
	}

	RunCommands(&state)
}
