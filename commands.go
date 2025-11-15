package main

import (
	"errors"
	"machine"
)

type Command struct {
	Flag byte
	Run  func(*State, []byte) error
}

var SetFanCommand = Command{
	Flag: 'F',
	Run: func(s *State, input []byte) error {
		f := b2i(input[0])
		if f <= 0 || f > 9 {
			return errors.New("invalid input: " + string(input))
		}

		println("running SetFan:", f)
		s.SetFan(uint(f))

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
		println("Input:", string(input))

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
