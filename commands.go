package main

import (
	"errors"
	"machine"
)

type Command struct {
	Flag byte
	// TODO: add configs so all input parsing can be done externally?
	// InputSize uint
	// InputType string
	Run func(*State, []byte) error
}

var SetFanCommand = Command{
	Flag: 'F',
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

var commands = []Command{
	SetFanCommand,
}

func RunCommands(s *State) {
	cmdMap := map[byte]Command{}

	for _, cmd := range commands {
		cmdMap[cmd.Flag] = cmd
	}

	for {
		input := readLine()
		if len(input) != 2 {
			continue
		}

		cmd, ok := cmdMap[input[0]]
		if !ok {
			continue
		}

		err := cmd.Run(s, input[1:])
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
