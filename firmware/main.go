package main

import (
	"math"
	"time"

	"github.com/calvinmclean/autoroast/firmware/commands"
	"github.com/calvinmclean/autoroast/firmware/device"
)

func main() {
	stepperCfg := device.StepperConfig{
		Pins:      stepperPins,
		StepMode:  device.StepModeHalf,
		StepDelay: 3000 * time.Microsecond,
	}

	servoCfg := device.ServoConfig{
		PWM: servoPWM,
		Pin: servoPin,
	}
	stepsPerIncrement := nominalStepsPerIncrement(30, 9, 8, 4096)
	calibrationCfg := device.CalibrationConfig{
		ServoBasePosition:     30,
		ServoClickPosition:    70,
		ServoPressDelay:       200 * time.Millisecond,
		ServoResetDelay:       250 * time.Millisecond,
		StepsPerIncrement:     stepsPerIncrement,
		DelayAfterStepperMove: 500 * time.Millisecond,
		BackstepRatio:         2,
	}

	d, err := device.New(stepperCfg, servoCfg, calibrationCfg)
	if err != nil {
		panic(err)
	}

	commands.Run(&d)
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
