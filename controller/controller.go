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
	}, nil
}

// Start will set the start time and ensures that everything is ready to go
func (c *Controller) Start() error {
	if c.fan == 0 || c.power == 0 {
		return errors.New("set initial fan/power before starting")
	}
	c.startTime = time.Now()

	println(c.ts(), "Started...")

	return nil
}

// Duration returns the duration that this has been running
func (c *Controller) Duration() time.Duration {
	return time.Since(c.startTime)
}

// ClickButton uses the servo motor to click the FreshRoast button to enable setting changes
func (c *Controller) ClickButton() {
	if c.verbose {
		println(c.ts(), "ClickButton")
	}

	err := c.servo.SetAngle(c.calibrationCfg.ServoClickPosition)
	if err != nil {
		println(c.ts(), "error setting servo angle:", err.Error())
		return
	}

	time.Sleep(c.calibrationCfg.ServoPressDelay)

	err = c.servo.SetAngle(c.calibrationCfg.ServoBasePosition)
	if err != nil {
		println(c.ts(), "error resetting servo angle:", err.Error())
		return
	}

	c.lastClick = time.Now()
	time.Sleep(c.calibrationCfg.ServoResetDelay)
}

// GoToMode will click the FreshRoast button until the target ControlMode is active
func (c *Controller) GoToMode(target ControlMode) bool {
	if c.verbose {
		println(c.ts(), "GoToMode:", target)
	}
	if target == ControlModeUnknown {
		return false
	}

	if c.currentControlMode == target {
		return false
	}

	// If we have started running, then we need extra logic to see if we are in "select mode".
	// This is required because the "select mode" automatically stops after ~3s, so we need to
	// click an extra time to get back into "select mode"
	if !c.startTime.IsZero() {
		if time.Since(c.lastClick) > 3*time.Second {
			c.ClickButton()
		}
	}

	for c.currentControlMode != target {
		c.ClickButton()
		c.currentControlMode = c.currentControlMode.Next()
	}

	return true
}

// FixControlMode manually sets the ControlMode to account for errors
func (c *Controller) FixControlMode(cm ControlMode) {
	c.currentControlMode = cm
}

// MoveFan controls the FreshRoast to move the fan value by the specified number of increments.
// It does not change the "State" of the device. This is useful for fixing off-by-one movements
func (c *Controller) MoveFan(i int32) {
	if c.verbose {
		println(c.ts(), "MoveFan", i)
	}
	if c.GoToMode(ControlModeFan) {
		time.Sleep(200 * time.Millisecond)
	}
	c.Move(i)
}

// MovePower controls the FreshRoast to move the power value by the specified number of increments.
// It does not change the "State" of the device. This is useful for fixing off-by-one movements
func (c *Controller) MovePower(i int32) {
	if c.verbose {
		println(c.ts(), "MovePower", i)
	}
	if c.GoToMode(ControlModePower) {
		time.Sleep(200 * time.Millisecond)
	}
	c.Move(i)
}

// MoveTimer controls the FreshRoast to move the timer value by the specified number of increments.
// If this exceeds the bounds, it will still move by the number of increments.
func (c *Controller) MoveTimer(i int32) {
	if c.verbose {
		println(c.ts(), "MoveTimer", i)
	}
	c.GoToMode(ControlModeTimer)
	c.Move(i)
}

// FixPower manually sets the current power to the specified value to account for errors. It does not control the device
func (c *Controller) FixPower(p uint) {
	c.power = p
}

// FixFan manually sets the current fan to the specified value to account for errors. It does not control the device
func (c *Controller) FixFan(f uint) {
	c.fan = f
}

// SetFan sets the FreshRoast fan to the specified value
func (c *Controller) SetFan(f uint) {
	if c.verbose {
		println(c.ts(), "SetFan", f)
	}
	if f < 1 || f > 9 {
		return
	}

	println(c.ts(), levelStr("F", f))

	delta := int32(f) - int32(c.fan)

	// When moving to extremes, we can move extra to re-calibrate and account for inaccuracy
	if f == 9 {
		delta += 3
	}
	if f == 1 {
		delta -= 3
	}

	c.MoveFan(delta)

	c.fan = f
}

// SetPower sets the FreshRoast power to the specified value
func (c *Controller) SetPower(p uint) {
	if c.verbose {
		println(c.ts(), "SetPower", p)
	}
	if p < 1 || p > 9 {
		return
	}

	println(c.ts(), levelStr("P", p))

	// When moving to extremes, we can move extra to re-calibrate and account for inaccuracy
	delta := int32(p) - int32(c.power)
	if p == 9 {
		delta += 3
	}
	if p == 1 {
		delta -= 3
	}

	c.MovePower(delta)

	c.power = p
}

// IncreaseTime just increases the time on device by 5m
func (c *Controller) IncreaseTime() {
	if c.verbose {
		println(c.ts(), "IncreaseTime")
	}
	c.GoToMode(ControlModeTimer)
	c.Move(5)
}

// Move moves the stepper by the specified number of increments
func (c *Controller) Move(n int32) {
	rawMove := float32(n)*c.calibrationCfg.StepsPerIncrement + c.remainder

	move := int32(math.Round(float64(rawMove)))
	c.remainder = rawMove - float32(move)

	// move forward a bit extra to make sure we "click" into place and then back up to expected position
	var backsteps int32
	if c.calibrationCfg.BackstepRatio > 0 {
		numBacksteps := int32(c.calibrationCfg.StepsPerIncrement / c.calibrationCfg.BackstepRatio)
		if move < 0 {
			move -= backsteps
			backsteps = +numBacksteps
		} else {
			move += backsteps
			backsteps = -numBacksteps
		}
	}

	c.stepper.Move(move)

	if backsteps != 0 {
		time.Sleep(200 * time.Millisecond)
		// Move back slightly
		c.stepper.Move(backsteps)
	}

	time.Sleep(c.calibrationCfg.DelayAfterStepperMove)
}

// Debug pritns out details of the Controller's state
func (c *Controller) Debug() {
	d := c.ts() + " " + levelStr("F", c.fan) + "/" + levelStr("P", c.power)
	d += " mode=" + c.currentControlMode.String()
	println(d)
}

// Verbose sets the Controller to Verbose mode and increases logging
func (c *Controller) Verbose() {
	c.verbose = true
	println(c.ts(), "Set Verbose Mode")
}

// ts returns the duration timestamp for logging
func (c *Controller) ts() string {
	if c.startTime.IsZero() {
		return "[-]"
	}
	return "[" + c.Duration().String() + "]"
}

// levelStr formats a power/fan level setting like F9 or P9
func levelStr[T uint | int](character string, level T) string {
	return character + string(byte(level)+'0')
}

func (c *Controller) MicroStep(n int32) {
	c.stepper.Move(n)
}

// Settings returns the current fan and power positions
func (c *Controller) Settings() (uint, uint) {
	return c.fan, c.power
}
