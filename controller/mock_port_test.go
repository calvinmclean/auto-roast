package controller

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/calvinmclean/autoroast"
)

func TestNewWithNoSerialUsesMockPort(t *testing.T) {
	c, err := New(Config{
		SerialPort:          SerialPortNone,
		BaudRate:            "115200",
		InitialFanSetting:   5,
		InitialPowerSetting: 5,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	port, ok := c.port.(*mockPort)
	if !ok {
		t.Fatalf("port = %T, want *mockPort", c.port)
	}
	if got, want := port.commands, []string{"I55"}; !equalStrings(got, want) {
		t.Errorf("commands = %q, want %q", got, want)
	}
}

func TestRunWithNoSerialLogsMockFirmwareCommands(t *testing.T) {
	c, err := New(Config{
		SerialPort:  SerialPortNone,
		BaudRate:    "115200",
		SessionName: "test",
		TWChartAddr: "mock",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var output bytes.Buffer
	err = c.Run(context.Background(), strings.NewReader("F5\n"), &output)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got, want := output.String(), fmt.Sprintf("[mock firmware] received F5%c\n", autoroast.TerminationChar); got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
