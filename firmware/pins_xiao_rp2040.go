//go:build xiao_rp2040

package main

import "machine"

var stepperPins = [4]machine.Pin{
	machine.D7,
	machine.D8,
	machine.D9,
	machine.D10,
}

var (
	servoPWM = machine.PWM0
	servoPin = machine.D6
)
