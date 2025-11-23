package controller

import (
	"errors"
	"math"
	"time"

	"tinygo.org/x/drivers/servo"
)

// Controller controls the FreshRoast SR800. It manages the Stepper and Servo motors and the machine's state
type Controller struct {
	stepper        *Stepper
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

	remainder float32
}

// New intializes the state with the provided configs
func New(stepperCfg StepperConfig, servoCfg ServoConfig, calibrationCfg CalibrationConfig) (Controller, error) {
	stepper, err := NewStepper(stepperCfg)
	if err != nil {
		return Controller{}, errors.New("error creating stepper: " + err.Error())
	}

	var myServo servo.Servo
	if servoCfg != (ServoConfig{}) {
		myServo, err = servo.New(servoCfg.PWM, servoCfg.Pin)
		if err != nil {
			return Controller{}, errors.New("error creating servo: " + err.Error())
		}
		err := myServo.SetAngle(calibrationCfg.ServoBasePosition)
		if err != nil {
			return Controller{}, errors.New("error setting servo angle: " + err.Error())
		}
	}

	return Controller{
		stepper:            stepper,
		servo:              myServo,
		calibrationCfg:     calibrationCfg,
		currentControlMode: ControlModeFan,
		fan:                0,
		power:              0,
		startTime:          time.Time{},
		lastClick:          time.Time{},
		verbose:            false,
		lastDirection:      +1,
	}, nil
}

// Start will set the start time and ensures that everything is ready to go
func (s *Controller) Start() error {
	if s.fan == 0 || s.power == 0 {
		return errors.New("set initial fan/power before starting")
	}
	s.startTime = time.Now()

	println(s.ts(), "Started...")

	return nil
}

// Duration returns the duration that this has been running
func (s *Controller) Duration() time.Duration {
	return time.Since(s.startTime)
}

// ClickButton uses the servo motor to click the FreshRoast button to enable setting changes
func (s *Controller) ClickButton() {
	if s.verbose {
		println(s.ts(), "ClickButton")
	}

	err := s.servo.SetAngle(s.calibrationCfg.ServoClickPosition)
	if err != nil {
		println(s.ts(), "error setting servo angle:", err.Error())
		return
	}

	time.Sleep(s.calibrationCfg.ServoPressDelay)

	err = s.servo.SetAngle(s.calibrationCfg.ServoBasePosition)
	if err != nil {
		println(s.ts(), "error resetting servo angle:", err.Error())
		return
	}

	s.lastClick = time.Now()
	time.Sleep(s.calibrationCfg.ServoResetDelay)
}

// GoToMode will click the FreshRoast button until the target ControlMode is active
func (s *Controller) GoToMode(target ControlMode) {
	if s.verbose {
		println(s.ts(), "GoToMode:", target)
	}
	if target == ControlModeUnknown {
		return
	}

	// If we have started running, then we need extra logic to see if we are in "select mode".
	// This is required because the "select mode" automatically stops after ~3s, so we need to
	// click an extra time to get back into "select mode"
	if !s.startTime.IsZero() {
		if time.Since(s.lastClick) > 3*time.Second {
			s.ClickButton()
		}
	}

	for s.currentControlMode != target {
		s.ClickButton()
		s.currentControlMode = s.currentControlMode.Next()
	}
}

// FixControlMode manually sets the ControlMode to account for errors
func (s *Controller) FixControlMode(cm ControlMode) {
	s.currentControlMode = cm
}

// MoveFan controls the FreshRoast to move the fan value by the specified number of increments.
// It does not change the "State" of the device. This is useful for fixing off-by-one movements
func (s *Controller) MoveFan(i int32) {
	if s.verbose {
		println(s.ts(), "MoveFan", i)
	}
	s.GoToMode(ControlModeFan)
	s.Move(i)
}

// MovePower controls the FreshRoast to move the power value by the specified number of increments.
// It does not change the "State" of the device. This is useful for fixing off-by-one movements
func (s *Controller) MovePower(i int32) {
	if s.verbose {
		println(s.ts(), "MovePower", i)
	}
	s.GoToMode(ControlModePower)
	s.Move(i)
}

// MoveTimer controls the FreshRoast to move the timer value by the specified number of increments.
// If this exceeds the bounds, it will still move by the number of increments.
func (s *Controller) MoveTimer(i int32) {
	if s.verbose {
		println(s.ts(), "MoveTimer", i)
	}
	s.GoToMode(ControlModeTimer)
	s.Move(i)
}

// FixPower manually sets the current power to the specified value to account for errors. It does not control the device
func (s *Controller) FixPower(p uint) {
	s.power = p
}

// FixFan manually sets the current fan to the specified value to account for errors. It does not control the device
func (s *Controller) FixFan(f uint) {
	s.fan = f
}

// SetFan sets the FreshRoast fan to the specified value
func (s *Controller) SetFan(f uint) {
	if s.verbose {
		println(s.ts(), "SetFan", f)
	}
	if f < 1 || f > 9 {
		return
	}

	println(s.ts(), levelStr("F", f))

	delta := int32(f) - int32(s.fan)

	// When moving to extremes, we can move extra to re-calibrate and account for inaccuracy
	if f == 9 {
		delta += 3
	}
	if f == 1 {
		delta -= 3
	}

	s.MoveFan(delta)

	s.fan = f
}

// SetPower sets the FreshRoast power to the specified value
func (s *Controller) SetPower(p uint) {
	if s.verbose {
		println(s.ts(), "SetPower", p)
	}
	if p < 1 || p > 9 {
		return
	}

	println(s.ts(), levelStr("P", p))

	// When moving to extremes, we can move extra to re-calibrate and account for inaccuracy
	delta := int32(p) - int32(s.power)
	if p == 9 {
		delta += 3
	}
	if p == 1 {
		delta -= 3
	}

	s.MovePower(delta)

	s.power = p
}

// IncreaseTime just increases the time on device by 5m
func (s *Controller) IncreaseTime() {
	if s.verbose {
		println(s.ts(), "IncreaseTime")
	}
	s.GoToMode(ControlModeTimer)
	s.Move(5)
}

// Move simply moves the stepper by the specified number of increments
func (s *Controller) Move(n int32) {
	rawMove := float32(n)*s.calibrationCfg.StepsPerIncrement + s.remainder

	// add or subtract backlash steps based on direction change
	if s.lastDirection < 0 && rawMove > 0 {
		rawMove += s.calibrationCfg.BacklashSteps
		s.lastDirection = +1
	} else if s.lastDirection > 0 && rawMove < 0 {
		rawMove -= s.calibrationCfg.BacklashSteps
		s.lastDirection = -1
	}

	move := int32(math.Round(float64(rawMove)))
	s.remainder = rawMove - float32(move)

	s.stepper.Move(move)

	time.Sleep(s.calibrationCfg.DelayAfterStepperMove)
}

// Debug pritns out details of the Controller's state
func (s *Controller) Debug() {
	d := s.ts() + " " + levelStr("F", s.fan) + "/" + levelStr("P", s.power)
	d += " mode=" + s.currentControlMode.String()
	println(d)
}

// Verbose sets the Controller to Verbose mode and increases logging
func (s *Controller) Verbose() {
	s.verbose = true
	println(s.ts(), "Set Verbose Mode")
}

// ts returns the duration timestamp for logging
func (s *Controller) ts() string {
	if s.startTime.IsZero() {
		return "[-]"
	}
	return "[" + s.Duration().String() + "]"
}

// levelStr formats a power/fan level setting like F9 or P9
func levelStr[T uint | int](character string, level T) string {
	return character + string(byte(level)+'0')
}

func (s *Controller) FullRev() {
	s.stepper.Move(4096)
}

// Settings returns the current fan and power positions
func (s *Controller) Settings() (uint, uint) {
	return s.fan, s.power
}
