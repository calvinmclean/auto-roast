package main_test

import (
	"strings"
	"testing"
	"time"

	"go.bug.st/serial"
)

const port = "/dev/cu.usbmodem2101"

func sendSerial(t *testing.T, in string, expectedLen int) string {
	t.Helper()
	mode := &serial.Mode{
		BaudRate: 115200,
	}

	port, err := serial.Open(port, mode)
	if err != nil {
		t.Errorf("unexpected error opening serial connection: %v", err)
		return ""
	}
	defer port.Close()

	_, err = port.Write([]byte(in))
	if err != nil {
		t.Errorf("unexpected error writing serial: %v", err)
		return ""
	}
	time.Sleep(100 * time.Millisecond)

	buf := make([]byte, expectedLen)
	total := 0
	port.SetReadTimeout(1 * time.Second)
	deadline := time.Now().Add(1 * time.Second)
	for total < expectedLen && time.Now().Before(deadline) {
		n, err := port.Read(buf)
		if err != nil {
			t.Errorf("unexpected error reading serial: %v", err)
			return ""
		}
		total += n
	}
	return string(buf[:total])
}

func TestSerial(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{
			"SetFanAndPower",
			"F1P1 D",
			`[-] F1
[-] P1
[-] F1/P1 mode=Fan
`,
		},
		{
			"RecoverFanAndPower",
			"D f5 p6 D",
			`[-] F1/P1 mode=Fan
[-] F5/P6 mode=Fan
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := strings.ReplaceAll(tt.expected, "\n", "\r\n")
			out := sendSerial(t, tt.in, len(expected))
			clean := strings.Trim(out, "\x00")
			if clean != expected {
				t.Errorf("expected=%q, got=%q", expected, clean)
			}
		})
	}
}
