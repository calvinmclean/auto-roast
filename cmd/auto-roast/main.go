package main

import (
	"autoroast/twchart"
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.bug.st/serial"
)

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

type serialController struct {
	twchartClient twchart.Client
	port          serial.Port
}

func newSerialController(portName string) (serialController, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return serialController{}, fmt.Errorf("unexpected error opening serial connection: %w", err)
	}

	client := twchart.NewClient("http://server.local:8087")

	return serialController{port: port, twchartClient: client}, nil
}

func (m serialController) passthroughCommand(in []byte) (string, error) {
	_, err := m.port.Write(in)
	if err != nil {
		return "", fmt.Errorf("unexpected error writing serial: %w", err)
	}

	buf := make([]byte, 128)
	n, err := m.port.Read(buf)
	if err != nil {
		return "", fmt.Errorf("unexpected error reading serial: %w", err)
	}
	return string(buf[:n]), nil
}

func run() error {
	// TODO: find default port https://github.com/bugst/go-serial/blob/master/example_getportlist_test.go
	c, err := newSerialController("/dev/cu.usbmodem1101")
	if err != nil {
		panic(err)
	}
	defer c.port.Close()

	ctx := context.Background()

	sessionID, err := c.twchartClient.CreateSession(ctx, "Test Bean", twchart.Probes{
		{
			Name:     "Ambient",
			Position: 1,
		},
		{
			Name:     "Roaster",
			Position: 2,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating session: %w", err)
	}

	// TODO: Prompt user for details like bean name and amount and probe mapping
	// TODO: Create TWChart session and keep ID in-memory
	// TODO: save session ID to text file (.current_session) so it can be resumed. defer file deletion
	_ = sessionID

	// Use bufio.Scanner for line-by-line input
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matched, err := c.handleExternalCommands(ctx, line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
		if matched {
			continue
		}

		switch line[0] {
		case 'F', 'P':
			err = c.addEvent(ctx, line)
		case 'S':
			err = c.handleStartCommand(ctx)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		resp, err := c.passthroughCommand([]byte(line))

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		} else {
			fmt.Println(resp)
		}
		fmt.Print("> ")
	}

	return scanner.Err()
}

// handleExternalCommands is responsible for commands that do not get sent to the firmware controller.
// It returns 'true' if a command is matched.
func (sc serialController) handleExternalCommands(ctx context.Context, line string) (bool, error) {
	switch line {
	case "COOL":
		return true, sc.handleCoolingStage(ctx)
	case "FC":
		return true, sc.addEvent(ctx, "First Crack")
	case "DONE":
		return true, sc.handleFinishCommand(ctx)
	case "NOTE":
	default:
		if strings.HasPrefix(line, "NOTE") {
			return true, sc.addEvent(ctx, strings.TrimPrefix(line, "NOTE "))
		}
	}

	return false, nil
}

func (sc serialController) addEvent(ctx context.Context, in string) error {
	return sc.twchartClient.AddEvent(ctx, in, time.Now())
}

func (sc serialController) handleStartCommand(ctx context.Context) error {
	return sc.twchartClient.SetStartTime(ctx, time.Now())
}

func (sc serialController) handleFinishCommand(ctx context.Context) error {
	return sc.twchartClient.Finish(ctx)
}

func (sc serialController) handleCoolingStage(ctx context.Context) error {
	return sc.twchartClient.AddStage(ctx, "Cooling", time.Now())
}
