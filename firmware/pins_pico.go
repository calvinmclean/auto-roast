//go:build pico

package main

import "machine"

var stepperPins = [4]machine.Pin{
	machine.GP16,
	machine.GP17,
	machine.GP18,
	machine.GP19,
}

var (
	servoPWM = machine.PWM3
	servoPin = machine.GP22
)
