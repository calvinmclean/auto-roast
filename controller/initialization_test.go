package controller

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"go.bug.st/serial"
)

type stalledSetupPort struct {
	closed chan struct{}
	once   sync.Once
}

func (p *stalledSetupPort) Write(b []byte) (int, error) { return len(b), nil }
func (p *stalledSetupPort) Read([]byte) (int, error) {
	<-p.closed
	return 0, io.EOF
}
func (p *stalledSetupPort) Close() error {
	p.once.Do(func() { close(p.closed) })
	return nil
}

func TestInitializationTimeoutClosesStalledPort(t *testing.T) {
	port := &stalledSetupPort{closed: make(chan struct{})}
	cfg := Config{SerialPort: "fake", BaudRate: "115200", InitialFanSetting: 5, InitialPowerSetting: 5}
	done := make(chan error, 1)
	go func() {
		_, err := newWithTimeout(cfg, 20*time.Millisecond, func(string, *serial.Mode) (io.ReadWriteCloser, error) {
			return port, nil
		})
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("timeout error = %v, want deadline exceeded", err)
		}
	case <-time.After(time.Second):
		_ = port.Close()
		t.Fatal("initialization did not time out")
	}
	select {
	case <-port.closed:
	case <-time.After(time.Second):
		t.Fatal("timed-out port was not closed")
	}
}

type failedSetupPort struct{ *stalledSetupPort }

func (p *failedSetupPort) Read([]byte) (int, error) { return 0, io.EOF }

func TestInitializationClosesPortOnFailure(t *testing.T) {
	port := &failedSetupPort{&stalledSetupPort{closed: make(chan struct{})}}
	cfg := Config{SerialPort: "fake", BaudRate: "115200", InitialFanSetting: 5, InitialPowerSetting: 5}
	_, err := newWithTimeout(cfg, time.Second, func(string, *serial.Mode) (io.ReadWriteCloser, error) {
		return port, nil
	})
	if err == nil {
		t.Fatal("failed handshake succeeded")
	}
	select {
	case <-port.closed:
	default:
		t.Fatal("failed initialization leaked the port")
	}
}

func TestInitializationClosesPortOpenedAfterTimeout(t *testing.T) {
	port := &stalledSetupPort{closed: make(chan struct{})}
	release := make(chan struct{})
	_, err := newWithTimeout(Config{SerialPort: "fake", BaudRate: "115200"}, 20*time.Millisecond,
		func(string, *serial.Mode) (io.ReadWriteCloser, error) {
			<-release
			return port, nil
		})
	close(release)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want deadline exceeded", err)
	}
	select {
	case <-port.closed:
	case <-time.After(time.Second):
		t.Fatal("late-opened port was not closed")
	}
}

func TestInitializationNoneNeverOpensSerial(t *testing.T) {
	cfg := Config{SerialPort: SerialPortNone, BaudRate: "115200", InitialFanSetting: 5, InitialPowerSetting: 5}
	c, err := newWithTimeout(cfg, time.Second, func(string, *serial.Mode) (io.ReadWriteCloser, error) {
		return nil, errors.New("must not open serial for none")
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
}
