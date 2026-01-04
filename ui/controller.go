package ui

import (
	"fmt"
	"io"
	"time"
)

type controllerWrapper struct {
	writer         io.Writer
	lastEventTimer *timer
}

func (c *controllerWrapper) SetFan(value float64) {
	c.lastEventTimer.Set(time.Now())
	fmt.Fprintf(c.writer, "F%.0f\n", value)
}

func (c *controllerWrapper) FixFan(value int) {
	fmt.Fprintf(c.writer, "f%d\n", value)
}

func (c *controllerWrapper) SetPower(value float64) {
	c.lastEventTimer.Set(time.Now())
	fmt.Fprintf(c.writer, "P%.0f\n", value)
}

func (c *controllerWrapper) FixPower(value int) {
	fmt.Fprintf(c.writer, "p%d\n", value)
}

func (c *controllerWrapper) RunStateCommand(s state) {
	stateCommand := s.command()
	if stateCommand != "" {
		fmt.Fprintf(c.writer, "%s\n", stateCommand)
	}
}
