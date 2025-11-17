package main

import (
	"errors"
	"machine"
)

type Command struct {
	Flag      byte
	InputSize uint
	Run       func(*State, []byte) error
}

var (
	SetFanCommand = &Command{
		Flag:      'F',
		InputSize: 1,
		Run: func(s *State, input []byte) error {
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

				s.SetFan(uint(f))
			}
			return nil
		},
	}
	SetPowerCommand = &Command{
		Flag:      'P',
		InputSize: 1,
		Run: func(s *State, input []byte) error {
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

				s.SetPower(uint(p))
			}
			return nil
		},
	}
	SetModeCommand = &Command{
		Flag:      'M',
		InputSize: 1,
		Run: func(s *State, input []byte) error {
			mode := ControlModeUnknown
			switch in := input[0]; in {
			case 'F':
				mode = ControlModeFan
			case 'P':
				mode = ControlModePower
			case 'T':
				mode = ControlModeTimer
			}
			s.GoToMode(mode)
			return nil
		},
	}
	ClickCommand = &Command{
		Flag:      'C',
		InputSize: 0,
		Run: func(s *State, input []byte) error {
			s.ClickButton()
			return nil
		},
	}
	StartCommand = &Command{
		Flag:      'S',
		InputSize: 0,
		Run: func(s *State, b []byte) error {
			return s.Start()
		},
	}
	DebugCommand = &Command{
		Flag:      'D',
		InputSize: 0,
		Run: func(s *State, b []byte) error {
			d := s.ts() + " " + levelStr("F", s.fan) + "/" + levelStr("P", s.power)
			d += " mode=" + s.currentControlMode.String()
			println(d)
			return nil
		},
	}
	VerboseCommand = &Command{
		Flag:      'V',
		InputSize: 0,
		Run: func(s *State, b []byte) error {
			s.verbose = true
			println(s.ts(), "Set Verbose Mode")
			return nil
		},
	}
	IncreaseTimeCommand = &Command{
		Flag:      'T',
		InputSize: 0,
		Run: func(s *State, b []byte) error {
			s.IncreaseTime()
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
}

func RunCommands(s *State) {
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

func b2i(b byte) int {
	return int(b - '0')
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
