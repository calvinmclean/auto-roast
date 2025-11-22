package controller

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/servo"
)

// StepperConfig ...
type StepperConfig struct {
	Pins      [4]machine.Pin
	StepMode  StepMode
	StepDelay time.Duration
}

// ServoConfig has device-level values for setting up the Servo
type ServoConfig struct {
	Pin machine.Pin
	PWM servo.PWM
}

// CalibrationConfig has values for the moving parts that depend on positioning and motor specifics
type CalibrationConfig struct {
	ServoBasePosition     int
	ServoClickPosition    int
	ServoPressDelay       time.Duration
	ServoResetDelay       time.Duration
	StepsPerIncrement     float32
	BacklashSteps         int32
	DelayAfterStepperMove time.Duration
}
