package controller

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/calvinmclean/autoroast"
	"github.com/calvinmclean/autoroast/twchart"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type Controller struct {
	twchartClient twchartClient
	port          serial.Port
	config        Config
}

type Config struct {
	SerialPort  string
	BaudRate    int
	TWChartAddr string

	SessionName string
	ProbesInput string

	IgnoreSerial bool
}

func GetSerialPorts() ([]string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, fmt.Errorf("error getting serial ports: %w", err)
	}

	var usbPorts []string
	for _, p := range ports {
		if p.IsUSB {
			usbPorts = append(usbPorts, p.Name)
		}
	}

	if len(usbPorts) == 0 {
		return nil, errors.New("no USB serial ports found")
	}

	return usbPorts, nil
}

func NewConfigFromEnv() (Config, error) {
	serialPort := os.Getenv("SERIAL_PORT")
	baudRateStr := os.Getenv("BAUD_RATE")
	twchartAddr := os.Getenv("TWCHART_ADDR")
	sessionName := os.Getenv("SESSION_NAME")
	probesInput := os.Getenv("PROBES_INPUT")
	ignoreSerial := os.Getenv("IGNORE_SERIAL") == "true"

	// Find default serial port if not set
	if serialPort == "" {
		ports, err := GetSerialPorts()
		if err != nil {
			return Config{}, fmt.Errorf("error getting serial ports: %w", err)
		}
		serialPort = ports[0]
	}

	// Parse baud rate, default to 115200
	baudRate := 115200
	if baudRateStr != "" {
		fmt.Sscanf(baudRateStr, "%d", &baudRate)
	}

	if sessionName == "" {
		sessionName = "default-session"
	}

	if probesInput == "" {
		probesInput = "1=Ambient,2=Beans"
	}

	if serialPort == "" && !ignoreSerial {
		return Config{}, errors.New("no serial port found")
	}

	if twchartAddr == "" {
		return Config{}, fmt.Errorf("missing required TWCHART_ADDR environment variable")
	}

	return Config{
		SerialPort:   serialPort,
		BaudRate:     baudRate,
		TWChartAddr:  twchartAddr,
		SessionName:  sessionName,
		ProbesInput:  probesInput,
		IgnoreSerial: ignoreSerial,
	}, nil
}

func NewFromEnv() (Controller, error) {
	cfg, err := NewConfigFromEnv()
	if err != nil {
		return Controller{}, err
	}
	return New(cfg)
}

func New(cfg Config) (Controller, error) {
	mode := &serial.Mode{
		BaudRate: cfg.BaudRate,
	}

	port, err := serial.Open(cfg.SerialPort, mode)
	if err != nil && !cfg.IgnoreSerial {
		return Controller{}, fmt.Errorf("unexpected error opening serial connection: %w", err)
	}

	var client twchartClient = noopTWChartClient{}
	if cfg.TWChartAddr != "mock" {
		client = twchart.NewClient(cfg.TWChartAddr)
	}

	return Controller{port: port, twchartClient: client, config: cfg}, nil
}

func (c Controller) Close() error {
	if c.port == nil {
		return nil
	}
	return c.port.Close()
}

func (c Controller) passthroughCommand(in []byte) (string, error) {
	if c.port == nil {
		return "", errors.New("no serial port")
	}

	_, err := c.port.Write(in)
	if err != nil {
		return "", fmt.Errorf("unexpected error writing serial: %w", err)
	}

	reader := bufio.NewReader(c.port)
	resp, err := reader.ReadString(autoroast.TerminationChar)
	if err != nil {
		return "", fmt.Errorf("unexpected error reading serial: %w", err)
	}
	return strings.TrimSpace(resp), nil
}

func (c Controller) Run(ctx context.Context, reader io.Reader, writer io.Writer) error {
	if c.config.SessionName == "" {
		return errors.New("missing -session")
	}

	probes := twchart.Probes{
		{Name: "Ambient", Position: 1},
		{Name: "Beans", Position: 2},
	}
	if c.config.ProbesInput != "" {
		var err error
		probes, err = twchart.ParseProbes(c.config.ProbesInput)
		if err != nil {
			return fmt.Errorf("invalid input for probes: %w", err)
		}
	}

	sessionID, err := c.twchartClient.CreateSession(ctx, c.config.SessionName, probes)
	if err != nil {
		return fmt.Errorf("error creating session: %w", err)
	}

	// TODO: save session ID to text file (.current_session) so it can be resumed. defer file deletion
	_ = sessionID

	// Use bufio.Scanner for line-by-line input
	scanner := bufio.NewScanner(reader)
	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			if scanner.Err() == nil {
				fmt.Println("\nReceived EOF (Ctrl-D). Exiting.")
				return nil
			}
			return scanner.Err()
		}

		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matched, err := c.handleExternalCommands(ctx, line)
		if err != nil {
			fmt.Fprintf(writer, "Error: %v\n", err)
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
			fmt.Fprintf(writer, "Error: %v\n", err)
			continue
		}

		resp, err := c.passthroughCommand([]byte(line))

		if err != nil {
			fmt.Fprintf(writer, "Error: %v\n", err)
		} else {
			fmt.Fprintln(writer, resp)
		}
	}
}

// handleExternalCommands is responsible for commands that do not get sent to the firmware controller.
// It returns 'true' if a command is matched.
func (c Controller) handleExternalCommands(ctx context.Context, line string) (bool, error) {
	switch line {
	case "PH", "PREHEAT":
		// TODO: should start if not already started
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
