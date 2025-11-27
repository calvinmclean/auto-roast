package commands

import (
	"errors"
	"time"

	"autoroast"
)

type Command struct {
	Flag        byte
	InputSize   uint
	Run         func(Controller, []byte) error
	Description string
}

// Controller is used to control a device
type Controller interface {
	MoveFan(int32)
	SetFan(uint)
	MovePower(int32)
	SetPower(uint)
	GoToMode(autoroast.ControlMode) bool
	ClickButton()
	Start() error
	Debug()
	Verbose()
	IncreaseTime()
	Settings() (uint, uint)
	FixFan(uint)
	FixPower(uint)
	MicroStep(int32)
	Move(int32)

	// I/O
	ReadByte() (byte, error)
}

var (
	SetFanCommand = &Command{
		Flag:      'F',
		InputSize: 1,
		Run: func(c Controller, input []byte) error {
			switch in := input[0]; in {
			case '-':
				c.MoveFan(-1)
			case '+':
				c.MoveFan(+1)
			default:
				f := b2i(in)
				if f <= 0 || f > 9 {
					return errors.New("invalid input: " + string(input))
				}
				c.SetFan(f)
			}
			return nil
		},
		Description: "Set or adjust the fan speed. Input: '-', '+', or 1-9.",
	}
	SetPowerCommand = &Command{
		Flag:      'P',
		InputSize: 1,
		Run: func(c Controller, input []byte) error {
			switch in := input[0]; in {
			case '-':
				c.MovePower(-1)
			case '+':
				c.MovePower(+1)
			default:
				p := b2i(in)
				if p <= 0 || p > 9 {
					return errors.New("invalid input: " + string(input))
				}
				c.SetPower(p)
			}
			return nil
		},
		Description: "Set or adjust the power level. Input: '-', '+', or 1-9.",
	}
	SetModeCommand = &Command{
		Flag:      'M',
		InputSize: 1,
		Run: func(c Controller, input []byte) error {
			mode := autoroast.ControlModeUnknown
			switch in := input[0]; in {
			case 'F':
				mode = autoroast.ControlModeFan
			case 'P':
				mode = autoroast.ControlModePower
			case 'T':
				mode = autoroast.ControlModeTimer
			}
			c.GoToMode(mode)
			return nil
		},
		Description: "Switch control mode. Input: 'F' (Fan), 'P' (Power), 'T' (Timer).",
	}
	ClickCommand = &Command{
		Flag:      'C',
		InputSize: 0,
		Run: func(c Controller, input []byte) error {
			c.ClickButton()
			return nil
		},
		Description: "Click the button. This does not change the device's memory of where it is positioned.",
	}
	StartCommand = &Command{
		Flag:      'S',
		InputSize: 0,
		Run: func(c Controller, b []byte) error {
			return c.Start()
		},
		Description: "Start roasting. This sets the timer to track durations of each change.",
	}
	DebugCommand = &Command{
		Flag:      'D',
		InputSize: 0,
		Run: func(c Controller, b []byte) error {
			c.Debug()
			return nil
		},
		Description: "Print the current state.",
	}
	VerboseCommand = &Command{
		Flag:      'V',
		InputSize: 0,
		Run: func(c Controller, b []byte) error {
			c.Verbose()
			return nil
		},
		Description: "Enable verbose output.",
	}
	IncreaseTimeCommand = &Command{
		Flag:      'T',
		InputSize: 0,
		Run: func(c Controller, b []byte) error {
			c.IncreaseTime()
			return nil
		},
		Description: "Increase the timer value.",
	}
	FixFanCommand = &Command{
		Flag:      'f',
		InputSize: 1,
		Run: func(c Controller, b []byte) error {
			v := b2i(b[0])
			// get the currently-set target, reset current position, then move to target
			target, _ := c.Settings()
			c.FixFan(v)
			c.SetFan(target)
			return nil
		},
		Description: "Fix the fan at a specific value and restore target. Input: 1-9.",
	}
	FixPowerCommand = &Command{
		Flag:      'p',
		InputSize: 1,
		Run: func(c Controller, b []byte) error {
			v := b2i(b[0])
			// get the currently-set target, reset current position, then move to target
			_, target := c.Settings()
			c.FixPower(v)
			c.SetPower(target)
			return nil
		},
		Description: "Fix the power at a specific value and restore target. Input: 1-9.",
	}
	TestCommand = &Command{
		Flag:      'Z',
		InputSize: 1,
		Run: func(c Controller, b []byte) error {
			test := byte('1')
			if len(b) > 0 {
				test = b[0]
			}

			switch test {
			case '1': // Run a simple test that toggles values and does not start
				funcs := []func(){
					func() { c.SetFan(5) },
					func() { c.SetPower(5) },
					func() { c.SetFan(9) },
					func() { c.SetFan(8) },
					func() { c.SetPower(9) },
					func() { c.SetPower(8) },
				}

				// Run with short delay
				for _, f := range funcs {
					f()
					time.Sleep(500 * time.Millisecond)
				}

				c.SetFan(1)
				c.SetPower(1)

				// Run with no delay
				for _, f := range funcs {
					f()
				}
			case '2':
				for range 10 {
					c.SetFan(5)
					c.SetFan(4)
				}
			}

			return nil
		},
		Description: "Run test routines. Input: '1' (toggle test), '2' (fan test).",
	}
	StepCommand = &Command{
		Flag:      's',
		InputSize: 2,
		Run: func(c Controller, b []byte) error {
			s := int32(1)
			if b[0] == '-' {
				s = -1
			} else if b[0] != '+' {
				return errors.New("invalid input")
			}

			v := b2i(b[1])

			c.Move(int32(v) * s)

			return nil
		},
		Description: "Move stepper motor by steps. Input: '+' or '-', then step count (1-9).",
	}
	FullRevolutionCommand = &Command{
		Flag:      'R',
		InputSize: 0,
		Run: func(c Controller, b []byte) error {
			c.MicroStep(4096)
			return nil
		},
		Description: "Move stepper motor a full revolution.",
	}
	InitCommand = &Command{
		Flag:      'I',
		InputSize: 2,
		Run: func(c Controller, b []byte) error {
			fan := b2i(b[0])
			power := b2i(b[1])
			c.FixFan(fan)
			c.FixPower(power)
			return nil
		},
		Description: "Initialize fan and power to specific values. Input: fan(1-9), power(1-9).",
	}
	MicroStepCommand = &Command{
		Flag:      0x1B,
		InputSize: 2,
		Run: func(c Controller, b []byte) error {
			if b[0] != '[' {
				return errors.New("invalid input")
			}
			switch b[1] {
			case 'D':
				c.MicroStep(5)
			case 'C':
				c.MicroStep(-5)
			}
			return nil
		},
		Description: "Move stepper motor by microsteps. Use left and right arrow keys.",
	}
	HelpCommand = &Command{
		Flag:        'H',
		InputSize:   0,
		Description: "Show all available commands and their descriptions.",
		Run: func(c Controller, b []byte) error {
			println("Available Commands:")
			for _, cmd := range commands {
				flagStr := ""
				if cmd.Flag >= 32 && cmd.Flag <= 126 {
					flagStr = string(cmd.Flag)
				} else {
					flagStr = "0x" + string("0123456789ABCDEF"[(cmd.Flag>>4)&0xF]) + string("0123456789ABCDEF"[cmd.Flag&0xF])
				}
				println(flagStr + ": " + cmd.Description)
			}
			return nil
		},
	}
)

func b2i(b byte) uint {
	v := uint(b - '0')
	if v < 1 || v > 9 {
		return 0
	}
	return v
}

var commands = []*Command{
	SetFanCommand,
	SetPowerCommand,
	SetModeCommand,
	ClickCommand,
	StartCommand,
	DebugCommand,
	VerboseCommand,
	IncreaseTimeCommand,
	FixFanCommand,
	FixPowerCommand,
	TestCommand,
	StepCommand,
	FullRevolutionCommand,
	InitCommand,
	MicroStepCommand,
}

func Run(c Controller) {
	cmdMap := map[byte]*Command{
		HelpCommand.Flag: HelpCommand,
	}

	for _, cmd := range commands {
		cmdMap[cmd.Flag] = cmd
	}

	for {
		cmdIn, err := c.ReadByte()
		if err != nil {
			continue
		}

		cmd, ok := cmdMap[cmdIn]
		if !ok {
			continue
		}

		in := make([]byte, cmd.InputSize)
		for i := 0; i < int(cmd.InputSize); {
			b, err := c.ReadByte()
			if err != nil {
				continue
			}

			in[i] = b
			i++
		}

		err = cmd.Run(c, in)
		if err != nil {
			println("error:", err.Error())
		}
	}
}
