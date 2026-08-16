package ui

import (
	"fmt"
	"io"
)

type controllerWrapper struct {
	writer io.Writer
}

func (c *controllerWrapper) write(format string, args ...any) {
	if c.writer == nil {
		return
	}
	fmt.Fprintf(c.writer, format, args...)
}

func (c *controllerWrapper) Note(note string) {
	c.write("NOTE %s\n", note)
}

func (c *controllerWrapper) Click() {
	c.write("C\n")
}

func (c *controllerWrapper) Debug() {
	c.write("D\n")
}

func (c *controllerWrapper) IncreaseTime() {
	c.write("T\n")
}

func (c *controllerWrapper) SetFan(value float64) {
	c.write("F%.0f\n", value)
}

func (c *controllerWrapper) FixFan(value int) {
	c.write("f%d\n", value)
}

func (c *controllerWrapper) SetPower(value float64) {
	c.write("P%.0f\n", value)
}

func (c *controllerWrapper) FixPower(value int) {
	c.write("p%d\n", value)
}

func (c *controllerWrapper) RunStateCommand(s state) {
	stateCommand := s.command()
	if stateCommand != "" {
		c.write("%s\n", stateCommand)
	}
}
