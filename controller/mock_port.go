package controller

import (
	"bytes"
	"fmt"

	"github.com/calvinmclean/autoroast"
)

type mockPort struct {
	commands  []string
	responses bytes.Buffer
}

func (p *mockPort) Write(command []byte) (int, error) {
	p.commands = append(p.commands, string(command))
	fmt.Fprintf(&p.responses, "[mock firmware] received %s%c", command, autoroast.TerminationChar)
	return len(command), nil
}

func (p *mockPort) Read(out []byte) (int, error) {
	return p.responses.Read(out)
}

func (p *mockPort) Close() error {
	return nil
}
