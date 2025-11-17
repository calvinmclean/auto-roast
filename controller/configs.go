package controller

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/servo"
)

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
	StepsPerIncrement     uint
	BacklashSteps         uint
	DelayAfterStepperMove time.Duration
}
