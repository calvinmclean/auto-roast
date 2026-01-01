package device

import (
	"errors"
	"machine"
	"math"
	"time"

	"autoroast"

	"tinygo.org/x/drivers/servo"
)

// Device controls the FreshRoast SR800. It manages the Stepper and Servo motors and the machine's state
type Device struct {
	stepper        *Stepper
	servo          servo.Servo
	calibrationCfg CalibrationConfig

	currentControlMode autoroast.ControlMode
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
func New(stepperCfg StepperConfig, servoCfg ServoConfig, calibrationCfg CalibrationConfig) (Device, error) {
	stepper, err := NewStepper(stepperCfg)
	if err != nil {
		return Device{}, errors.New("error creating stepper: " + err.Error())
	}

	var myServo servo.Servo
	if servoCfg != (ServoConfig{}) {
		myServo, err = servo.New(servoCfg.PWM, servoCfg.Pin)
		if err != nil {
			return Device{}, errors.New("error creating servo: " + err.Error())
		}
		err := myServo.SetAngle(calibrationCfg.ServoBasePosition)
		if err != nil {
			return Device{}, errors.New("error setting servo angle: " + err.Error())
		}
	}

	return Device{
		stepper:            stepper,
		servo:              myServo,
		calibrationCfg:     calibrationCfg,
		currentControlMode: autoroast.ControlModeFan,
		fan:                0,
		power:              0,
		startTime:          time.Time{},
		lastClick:          time.Time{},
		verbose:            false,
	}, nil
}

// Start will set the start time and ensures that everything is ready to go
func (d *Device) Start() error {
	if d.fan == 0 || d.power == 0 {
		return errors.New("set initial fan/power before starting")
	}
	d.startTime = time.Now()

	println(d.ts(), "Started...")

	return nil
}

// Duration returns the duration that this has been running
func (d *Device) Duration() time.Duration {
	return time.Since(d.startTime)
}

// ClickButton uses the servo motor to click the FreshRoast button to enable setting changes
func (d *Device) ClickButton() {
	if d.verbose {
		println(d.ts(), "ClickButton")
	}

	err := d.servo.SetAngle(d.calibrationCfg.ServoClickPosition)
	if err != nil {
		println(d.ts(), "error setting servo angle:", err.Error())
		return
	}

	time.Sleep(d.calibrationCfg.ServoPressDelay)

	err = d.servo.SetAngle(d.calibrationCfg.ServoBasePosition)
	if err != nil {
		println(d.ts(), "error resetting servo angle:", err.Error())
		return
	}

	d.lastClick = time.Now()
	time.Sleep(d.calibrationCfg.ServoResetDelay)
}

// GoToMode will click the FreshRoast button until the target ControlMode is active
func (d *Device) GoToMode(target autoroast.ControlMode) bool {
	if d.verbose {
		println(d.ts(), "GoToMode:", target)
	}

	if target == autoroast.ControlModeUnknown {
		return false
	}

	// If we have started running, then we need extra logic to see if we are in "select mode".
	// This is required because the "select mode" automatically stops after ~3s, so we need to
	// click an extra time to get back into "select mode"
	var clicked bool
	if !d.startTime.IsZero() {
		if time.Since(d.lastClick) > 3*time.Second {
			d.ClickButton()
			clicked = true
		}
	}

	if d.currentControlMode == target {
		return clicked
	}

	for d.currentControlMode != target {
		d.ClickButton()
		d.currentControlMode = d.currentControlMode.Next()
		clicked = true
	}

	return clicked
}

// FixControlMode manually sets the ControlMode to account for errors
func (d *Device) FixControlMode(cm autoroast.ControlMode) {
	d.currentControlMode = cm
}

// MoveFan controls the FreshRoast to move the fan value by the specified number of increments.
// It does not change the "State" of the device. This is useful for fixing off-by-one movements
func (d *Device) MoveFan(i int32) {
	if d.verbose {
		println(d.ts(), "MoveFan", i)
	}
	if d.GoToMode(autoroast.ControlModeFan) {
		time.Sleep(200 * time.Millisecond)
	}
	d.Move(i)
}

// MovePower controls the FreshRoast to move the power value by the specified number of increments.
// It does not change the "State" of the device. This is useful for fixing off-by-one movements
func (d *Device) MovePower(i int32) {
	if d.verbose {
		println(d.ts(), "MovePower", i)
	}
	if d.GoToMode(autoroast.ControlModePower) {
		time.Sleep(200 * time.Millisecond)
	}
	d.Move(i)
}

// MoveTimer controls the FreshRoast to move the timer value by the specified number of increments.
// If this exceeds the bounds, it will still move by the number of increments.
func (d *Device) MoveTimer(i int32) {
	if d.verbose {
		println(d.ts(), "MoveTimer", i)
	}
	d.GoToMode(autoroast.ControlModeTimer)
	d.Move(i)
}

// FixPower manually sets the current power to the specified value to account for errors. It does not control the device
func (d *Device) FixPower(p uint) {
	d.power = p
}

// FixFan manually sets the current fan to the specified value to account for errors. It does not control the device
func (d *Device) FixFan(f uint) {
	d.fan = f
}

// SetFan sets the FreshRoast fan to the specified value
func (d *Device) SetFan(f uint) {
	if d.verbose {
		println(d.ts(), "SetFan", f)
	}
	if f < 1 || f > 9 {
		return
	}

	println(d.ts(), levelStr("F", f))

	delta := int32(f) - int32(d.fan)

	// When moving to extremes, we can move extra to re-calibrate and account for inaccuracy
	if f == 9 {
		delta += 3
	}
	if f == 1 {
		delta -= 3
	}

	d.MoveFan(delta)

	d.fan = f
}

// SetPower sets the FreshRoast power to the specified value
func (d *Device) SetPower(p uint) {
	if d.verbose {
		println(d.ts(), "SetPower", p)
	}
	if p < 1 || p > 9 {
		return
	}

	println(d.ts(), levelStr("P", p))

	// When moving to extremes, we can move extra to re-calibrate and account for inaccuracy
	delta := int32(p) - int32(d.power)
	if p == 9 {
		delta += 3
	}
	if p == 1 {
		delta -= 3
	}

	d.MovePower(delta)

	d.power = p
}

// IncreaseTime just increases the time on device by 5m
func (d *Device) IncreaseTime() {
	if d.verbose {
		println(d.ts(), "IncreaseTime")
	}
	d.GoToMode(autoroast.ControlModeTimer)
	d.Move(5)
}

// Move moves the stepper by the specified number of increments
func (d *Device) Move(n int32) {
	rawMove := float32(n)*d.calibrationCfg.StepsPerIncrement + d.remainder

	move := int32(math.Round(float64(rawMove)))
	d.remainder = rawMove - float32(move)

	// move forward a bit extra to make sure we "click" into place and then back up to expected position
	var backsteps int32
	if d.calibrationCfg.BackstepRatio > 0 {
		numBacksteps := int32(d.calibrationCfg.StepsPerIncrement / d.calibrationCfg.BackstepRatio)
		if move < 0 {
			move -= backsteps
			backsteps = +numBacksteps
		} else {
			move += backsteps
			backsteps = -numBacksteps
		}
	}

	d.stepper.Move(move)

	if backsteps != 0 {
		time.Sleep(200 * time.Millisecond)
		// Move back slightly
		d.stepper.Move(backsteps)
	}

	time.Sleep(d.calibrationCfg.DelayAfterStepperMove)
}

// Debug pritns out details of the Device's state
func (c *Device) Debug() {
	d := c.ts() + " " + levelStr("F", c.fan) + "/" + levelStr("P", c.power)
	d += " mode=" + c.currentControlMode.String()
	println(d)
}

// Verbose sets the Device to Verbose mode and increases logging
func (d *Device) Verbose() {
	d.verbose = true
	println(d.ts(), "Set Verbose Mode")
}

// ts returns the duration timestamp for logging
func (d *Device) ts() string {
	if d.startTime.IsZero() {
		return "[-]"
	}
	return "[" + d.Duration().String() + "]"
}

// levelStr formats a power/fan level setting like F9 or P9
func levelStr[T uint | int](character string, level T) string {
	return character + string(byte(level)+'0')
}

func (d *Device) MicroStep(n int32) {
	d.stepper.Move(n)
}

// Settings returns the current fan and power positions
func (d *Device) Settings() (uint, uint) {
	return d.fan, d.power
}

func (d *Device) ReadByte() (byte, error) {
	return machine.Serial.ReadByte()
}

func (d *Device) WriteByte(b byte) error {
	return machine.Serial.WriteByte(b)
}
