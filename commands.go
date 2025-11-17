package main

import (
	"errors"
	"machine"
	"time"

	"autoroast/controller"
)

type Command struct {
	Flag      byte
	InputSize uint
	Run       func(*controller.Controller, []byte) error
}

var (
	SetFanCommand = &Command{
		Flag:      'F',
		InputSize: 1,
		Run: func(s *controller.Controller, input []byte) error {
			switch in := input[0]; in {
			case '-':
				s.MoveFan(-1)
			case '+':
				s.MoveFan(+1)
			default:
				f := b2i(in)
				if f <= 0 || f > 9 {
					return errors.New("invalid input: " + string(input))
				}

				s.SetFan(f)
			}
			return nil
		},
	}
	SetPowerCommand = &Command{
		Flag:      'P',
		InputSize: 1,
		Run: func(s *controller.Controller, input []byte) error {
			switch in := input[0]; in {
			case '-':
				s.MovePower(-1)
			case '+':
				s.MovePower(+1)
			default:
				p := b2i(in)
				if p <= 0 || p > 9 {
					return errors.New("invalid input: " + string(input))
				}

				s.SetPower(p)
			}
			return nil
		},
	}
	SetModeCommand = &Command{
		Flag:      'M',
		InputSize: 1,
		Run: func(s *controller.Controller, input []byte) error {
			mode := controller.ControlModeUnknown
			switch in := input[0]; in {
			case 'F':
				mode = controller.ControlModeFan
			case 'P':
				mode = controller.ControlModePower
			case 'T':
				mode = controller.ControlModeTimer
			}
			s.GoToMode(mode)
			return nil
		},
	}
	ClickCommand = &Command{
		Flag:      'C',
		InputSize: 0,
		Run: func(s *controller.Controller, input []byte) error {
			s.ClickButton()
			return nil
		},
	}
	StartCommand = &Command{
		Flag:      'S',
		InputSize: 0,
		Run: func(s *controller.Controller, b []byte) error {
			return s.Start()
		},
	}
	DebugCommand = &Command{
		Flag:      'D',
		InputSize: 0,
		Run: func(s *controller.Controller, b []byte) error {
			s.Debug()
			return nil
		},
	}
	VerboseCommand = &Command{
		Flag:      'V',
		InputSize: 0,
		Run: func(s *controller.Controller, b []byte) error {
			s.Verbose()
			return nil
		},
	}
	IncreaseTimeCommand = &Command{
		Flag:      'T',
		InputSize: 0,
		Run: func(s *controller.Controller, b []byte) error {
			s.IncreaseTime()
			return nil
		},
	}
	RecoverCommand = &Command{
		Flag:      'R',
		InputSize: 2,
		Run: func(c *controller.Controller, b []byte) error {
			if len(b) != 2 {
				return errors.New("invalid input")
			}

			v := b2i(b[1])

			switch b[0] {
			case 'F':
				c.FixFan(v)
			case 'P':
				c.FixPower(v)
			}

			return nil
		},
	}
	TestCommand = &Command{
		Flag:      'Z',
		InputSize: 1,
		Run: func(c *controller.Controller, b []byte) error {
			test := byte('1')
			if len(b) > 0 {
				test = b[0]
			}

			switch test {
			case '1': // Run a simple test that toggles values and does not start
				funcs := []func(){
					func() { c.SetFan(5) },
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
			}

			return nil
		},
	}
)

var commands = []*Command{
	SetFanCommand,
	SetPowerCommand,
	SetModeCommand,
	ClickCommand,
	StartCommand,
	DebugCommand,
	VerboseCommand,
	IncreaseTimeCommand,
	RecoverCommand,
	TestCommand,
}

func RunCommands(s *controller.Controller) {
	cmdMap := map[byte]*Command{}

	for _, cmd := range commands {
		cmdMap[cmd.Flag] = cmd
	}

	for {
		cmdIn, err := machine.Serial.ReadByte()
		if err != nil {
			continue
		}

		cmd, ok := cmdMap[cmdIn]
		if !ok {
			continue
		}

		in := make([]byte, cmd.InputSize)
		for i := range cmd.InputSize {
			b, err := machine.Serial.ReadByte()
			if err != nil {
				println(err)
				continue
			}
			in[i] = b
		}

		err = cmd.Run(s, in)
		if err != nil {
			println("error:", err.Error())
		}
	}
}

func b2i(b byte) uint {
	return uint(b - '0')
}

func readLine() []byte {
	var result []byte
	for {
		c, err := machine.Serial.ReadByte()
		if err != nil {
			continue
		}
		if c == '\n' {
			break
		}
		result = append(result, c)
	}
	return result
}
