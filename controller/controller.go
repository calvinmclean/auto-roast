package controller

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/calvinmclean/autoroast"
	"github.com/calvinmclean/autoroast/twchart"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

const SerialPortNone = "none"

var ErrNoUSBSerial = errors.New("no USB serial ports found")

type Controller struct {
	twchartClient twchartClient
	port          serial.Port
	config        Config
}

type Config struct {
	SerialPort          string
	BaudRate            string
	TWChartAddr         string
	SessionName         string
	ProbesInput         string
	InitialFanSetting   int
	InitialPowerSetting int
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
		return nil, ErrNoUSBSerial
	}

	return usbPorts, nil
}

func NewConfigFromEnv() Config {
	serialPort := os.Getenv("SERIAL_PORT")
	baudRate := os.Getenv("BAUD_RATE")
	twchartAddr := os.Getenv("TWCHART_ADDR")
	sessionName := os.Getenv("SESSION_NAME")
	probesInput := os.Getenv("PROBES_INPUT")

	if baudRate == "" {
		baudRate = "115200"
	}

	if probesInput == "" {
		probesInput = "1=Ambient,2=Beans"
	}

	initialFanSetting := 0
	if fanStr := os.Getenv("INITIAL_FAN_SETTING"); fanStr != "" {
		if fan, err := strconv.Atoi(fanStr); err == nil && fan >= 1 && fan <= 9 {
			initialFanSetting = fan
		}
	}

	initialPowerSetting := 0
	if powerStr := os.Getenv("INITIAL_POWER_SETTING"); powerStr != "" {
		if power, err := strconv.Atoi(powerStr); err == nil && power >= 1 && power <= 9 {
			initialPowerSetting = power
		}
	}

	return Config{
		SerialPort:          serialPort,
		BaudRate:            baudRate,
		TWChartAddr:         twchartAddr,
		SessionName:         sessionName,
		ProbesInput:         probesInput,
		InitialFanSetting:   initialFanSetting,
		InitialPowerSetting: initialPowerSetting,
	}
}

func NewFromEnv() (Controller, error) {
	return New(NewConfigFromEnv())
}

func New(cfg Config) (Controller, error) {
	// Find default serial port if not set
	if cfg.SerialPort == "" {
		ports, err := GetSerialPorts()
		if err != nil {
			return Controller{}, fmt.Errorf("error getting serial ports: %w", err)
		}
		cfg.SerialPort = ports[0]
	}

	baudRate, err := strconv.Atoi(cfg.BaudRate)
	if err != nil {
		return Controller{}, fmt.Errorf("invalid BaudRate: %w", err)
	}
	mode := &serial.Mode{
		BaudRate: baudRate,
	}

	var port serial.Port
	if cfg.SerialPort != SerialPortNone {
		var err error
		port, err = serial.Open(cfg.SerialPort, mode)
		if err != nil {
			return Controller{}, fmt.Errorf("unexpected error opening serial connection: %w", err)
		}
	}

	controller := Controller{port: port, twchartClient: noopTWChartClient{}, config: cfg}

	// Set initial fan and power values if they are non-zero
	if cfg.InitialFanSetting != 0 && cfg.InitialPowerSetting != 0 {
		cmd := fmt.Sprintf("I%d%d", cfg.InitialFanSetting, cfg.InitialPowerSetting)
		_, err := controller.passthroughCommand([]byte(cmd))
		if err != nil {
			err = fmt.Errorf("error setting initial fan and power: %w", err)
			if cfg.SerialPort == SerialPortNone {
				fmt.Println(err)
			} else {
				return Controller{}, err
			}
		}
	}

	if cfg.TWChartAddr != "mock" && cfg.TWChartAddr != "" {
		controller.twchartClient = twchart.NewClient(cfg.TWChartAddr)
	}

	return controller, nil
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
		return errors.New("missing SessionName")
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
