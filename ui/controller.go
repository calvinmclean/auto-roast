package ui

import (
	"fmt"
	"io"
	"time"
)

type controller struct {
	writer         io.Writer
	lastEventTimer *timer
}

func (c *controller) SetFan(value float64) {
	c.lastEventTimer.Set(time.Now())
	fmt.Fprintf(c.writer, "F%.0f\n", value)
}

func (c *controller) FixFan(value int) {
	fmt.Fprintf(c.writer, "f%d\n", value)
}

func (c *controller) SetPower(value float64) {
	c.lastEventTimer.Set(time.Now())
	fmt.Fprintf(c.writer, "P%.0f\n", value)
}

func (c *controller) FixPower(value int) {
	fmt.Fprintf(c.writer, "p%d\n", value)
}

func (c *controller) RunStateCommand(s state) {
	stateCommand := s.command()
	if stateCommand != "" {
		fmt.Fprintf(c.writer, "%s\n", stateCommand)
	}
}
