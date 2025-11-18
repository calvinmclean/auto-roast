package main

import (
	"machine"
	"time"

	"autoroast/controller"

	"tinygo.org/x/drivers/easystepper"
)

func main() {
	stepperCfg := easystepper.DeviceConfig{
		Pin1: machine.GP0, Pin2: machine.GP1, Pin3: machine.GP2, Pin4: machine.GP3,
		// StepCount: 200,
		// RPM:       50,
		// Mode:      easystepper.ModeFour,
	}

	servoCfg := controller.ServoConfig{
		PWM: machine.PWM2,
		Pin: machine.GP4,
	}
	calibrationCfg := controller.CalibrationConfig{
		ServoBasePosition:  10,
		ServoClickPosition: 42,
		ServoPressDelay:    150 * time.Millisecond,
		ServoResetDelay:    250 * time.Millisecond,
		// StepsPerIncrement:  61,
		StepsPerIncrement:     30.5,
		BacklashSteps:         0,
		DelayAfterStepperMove: 750 * time.Millisecond,
	}

	state, err := controller.New(stepperCfg, servoCfg, calibrationCfg)
	if err != nil {
		panic(err)
	}

	RunCommands(&state)
}
