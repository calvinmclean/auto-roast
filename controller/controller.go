package controller

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

type Controller struct {
	twchartClient twchart.Client
	port          serial.Port
}

type Config struct {
	SerialPort  string
	BaudRate    int
	TWChartAddr string
}

func NewFromEnv() (Controller, error) {

	serialPort := os.Getenv("SERIAL_PORT")
	baudRateStr := os.Getenv("BAUD_RATE")
	twchartAddr := os.Getenv("TWCHART_ADDR")

	// Find default serial port if not set
	if serialPort == "" {
		ports, err := serial.GetPortsList()
		if err != nil {
			return Controller{}, fmt.Errorf("error getting serial ports: %w", err)
		}
		if len(ports) == 0 {
			return Controller{}, fmt.Errorf("no serial ports found")
		}
		serialPort = ports[0]
	}

	// Parse baud rate, default to 115200
	baudRate := 115200
	if baudRateStr != "" {
		fmt.Sscanf(baudRateStr, "%d", &baudRate)
	}

	// Error if missing TWCHART_URL
	if twchartAddr == "" {
		return Controller{}, fmt.Errorf("missing required TWCHART_ADDR environment variable")
	}

	cfg := Config{
		SerialPort:  serialPort,
		BaudRate:    baudRate,
		TWChartAddr: twchartAddr,
	}
	return New(cfg)
}

func New(cfg Config) (Controller, error) {
	mode := &serial.Mode{
		BaudRate: cfg.BaudRate,
	}

	port, err := serial.Open(cfg.SerialPort, mode)
	if err != nil {
		return Controller{}, fmt.Errorf("unexpected error opening serial connection: %w", err)
	}

	client := twchart.NewClient(cfg.TWChartAddr)

	return Controller{port: port, twchartClient: client}, nil
}

func (c Controller) Close() error {
	return c.port.Close()
}

func (c Controller) passthroughCommand(in []byte) (string, error) {
	_, err := c.port.Write(in)
	if err != nil {
		return "", fmt.Errorf("unexpected error writing serial: %w", err)
	}

	buf := make([]byte, 128)
	n, err := c.port.Read(buf)
	if err != nil {
		return "", fmt.Errorf("unexpected error reading serial: %w", err)
	}
	return string(buf[:n]), nil
}

func (c Controller) Run(ctx context.Context) error {
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
			err = c.twchartClient.AddEvent(ctx, line, time.Now())
		case 'S':
			err = c.twchartClient.SetStartTime(ctx, time.Now())
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
func (c Controller) handleExternalCommands(ctx context.Context, line string) (bool, error) {
	switch line {
	case "PH", "PREHEAT":
		return true, c.twchartClient.AddStage(ctx, "Preheat", time.Now())
	case "ROAST", "ROASTING":
		return true, c.twchartClient.AddStage(ctx, "Roasting", time.Now())
	case "FC", "CRACK":
		return true, c.twchartClient.AddEvent(ctx, "First Crack", time.Now())
	case "COOL":
		return true, c.twchartClient.AddStage(ctx, "Cooling", time.Now())
	case "DONE":
		return true, c.twchartClient.Done(ctx)
	default:
		if strings.HasPrefix(line, "NOTE") {
			return true, c.twchartClient.AddEvent(ctx, strings.TrimPrefix(line, "NOTE "), time.Now())
		}
	}

	return false, nil
}
