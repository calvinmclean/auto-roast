package main

import (
	"machine"
	"math"
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
		ServoBasePosition:     15,
		ServoClickPosition:    50,
		ServoPressDelay:       150 * time.Millisecond,
		ServoResetDelay:       250 * time.Millisecond,
		StepsPerIncrement:     nominalStepsPerIncrement(30, 9, 8, 4096),
		BacklashSteps:         70,
		DelayAfterStepperMove: 500 * time.Millisecond,
	}

	state, err := controller.New(stepperCfg, servoCfg, calibrationCfg)
	if err != nil {
		panic(err)
	}

	RunCommands(&state)
}

// nominalStepsPerIncrement returns the rounded nominal microsteps required
// driverTeeth = stepper gear teeth (9), drivenTeeth = encoder gear teeth (8)
func nominalStepsPerIncrement(encoderIncrements int, driverTeeth, drivenTeeth int, stepsPerRev int) float32 {
	degPerInc := 360.0 / float64(encoderIncrements)
	stepperDeg := degPerInc * float64(drivenTeeth) / float64(driverTeeth)
	stepsPerDeg := float64(stepsPerRev) / 360.0
	result := math.Round(stepperDeg * stepsPerDeg)
	return float32(result)
}
