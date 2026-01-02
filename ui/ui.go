package ui

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type state int

const (
	stateNone state = iota
	statePreheat
	stateRoasting
	stateFirstCrack
	stateCooling
	stateDone
)

func (s state) String() string {
	switch s {
	case statePreheat:
		return "Preheat"
	case stateRoasting:
		return "Roasting"
	case stateFirstCrack:
		return "First Crack"
	case stateCooling:
		return "Cooling"
	case stateDone:
		return "Done"
	default:
		return "Unknown"
	}
}

func (s state) next() state {
	if s == stateDone {
		// Done has no next State
		return stateDone
	}
	return s + 1
}

func (s state) command() string {
	switch s {
	case statePreheat:
		// Start and Preaheat
		return "S\nPREHEAT"
	case stateRoasting:
		return "ROASTING"
	case stateFirstCrack:
		return "FC"
	case stateCooling:
		return "COOL"
	case stateDone:
		return "DONE"
	default:
		return ""
	}
}

func createSlider(labelText string, onSet func(float64), onFix func(int)) *fyne.Container {
	defaultValue := 9.0
	valueLabel := widget.NewLabel(fmt.Sprintf("%.0f", defaultValue))

	slider := widget.NewSlider(1, 9)
	slider.Step = 1
	slider.SetValue(defaultValue)
	slider.OnChanged = func(value float64) {
		valueLabel.SetText(fmt.Sprintf("%.0f", value))
	}
	slider.OnChangeEnded = onSet
	fixNumberEntry := widget.NewEntry()
	fixNumberEntry.OnSubmitted = func(s string) {
		fixNumberEntry.SetText("")

		number, err := strconv.Atoi(s)
		if err != nil || number == 0 {
			fmt.Println("Invalid input. Please enter a single number.")
			return
		}
		onFix(number)
	}

	fixButton := widget.NewButton("Fix", func() {
		fixNumberEntry.OnSubmitted(fixNumberEntry.Text)
	})

	container := container.NewVBox(
		container.NewGridWithColumns(3,
			widget.NewLabel(labelText),
			valueLabel,
			container.NewHBox(fixNumberEntry, fixButton),
		),
		slider,
	)

	return container
}

type timer struct {
	showMillis bool
	startTime  time.Time
	mtx        *sync.Mutex
	text       *canvas.Text
	stop       chan struct{}
}

func newTimer(showMillis bool) *timer {
	initText := "00:00"
	if showMillis {
		initText = "00:00.000"
	}
	return &timer{
		showMillis: showMillis,
		startTime:  time.Time{},
		mtx:        &sync.Mutex{},
		text:       canvas.NewText(initText, nil),
		stop:       make(chan struct{}),
	}
}

func (t *timer) Set(start time.Time) {
	t.mtx.Lock()
	t.startTime = start
	t.mtx.Unlock()
}

func (t *timer) Stop() {
	close(t.stop)
}

func (t *timer) Go(waitForStart chan struct{}) {
	d := time.Second
	if t.showMillis {
		d = 64 * time.Millisecond
	}

	go func() {
		<-waitForStart
		for range time.Tick(d) {
			select {
			case <-t.stop:
				return
			default:
			}
			fyne.Do(func() {
				t.mtx.Lock()
				elapsed := time.Since(t.startTime)
				minutes := int(elapsed.Minutes())
				seconds := int(elapsed.Seconds()) % 60
				if t.showMillis {
					millis := int(elapsed.Milliseconds()) % 1000
					t.text.Text = fmt.Sprintf("%02d:%02d.%03d", minutes, seconds, millis)
				} else {
					t.text.Text = fmt.Sprintf("%02d:%02d", minutes, seconds)
				}
				t.text.Refresh()
				t.mtx.Unlock()
			})
		}
	}()
}

func createLogAccordion() (*widget.Accordion, *widget.Entry) {
	logScroll := widget.NewMultiLineEntry()
	logScroll.Wrapping = fyne.TextWrapWord
	logScroll.SetMinRowsVisible(10)

	// disable editing by undoing changes. this allows it to not have changed colors from Disable
	logScroll.OnChanged = func(_ string) {
		logScroll.Undo()
	}

	return widget.NewAccordion(
		widget.NewAccordionItem("Logs", logScroll),
	), logScroll
}

type RoasterUI struct {
	logEntry *widget.Entry
}

func NewRoasterUI() *RoasterUI {
	return &RoasterUI{}
}

func (ui *RoasterUI) Run(ctx context.Context, w io.Writer) {
	application := app.New()

	window := application.NewWindow("Auto Roast")

	currentState := stateNone

	overallTimer := newTimer(false)
	lastEventTimer := newTimer(true)
	fcTimer := newTimer(true)

	waitForStart := make(chan struct{})
	overallTimer.Go(waitForStart)
	lastEventTimer.Go(waitForStart)

	waitForFC := make(chan struct{})
	fcTimer.Go(waitForFC)

	var stateButton *widget.Button
	stateButton = widget.NewButton(currentState.next().String(), func() {
		currentState++

		lastEventTimer.Set(time.Now())
		stateButton.SetText(currentState.next().String())

		switch currentState {
		case stateFirstCrack:
			fcTimer.text.Color = color.RGBA{R: 139, G: 0, B: 0, A: 255}
			fcTimer.Set(time.Now())
			close(waitForFC)
		case stateFirstCrack + 1:
			fcTimer.Stop()
		case 1:
			overallTimer.Set(time.Now())
			close(waitForStart)
		case stateDone:
			stateButton.Disable()
			overallTimer.Stop()
			lastEventTimer.Stop()
		}

		stateCommand := currentState.command()
		if stateCommand != "" {
			w.Write(fmt.Appendf([]byte{}, "%s\n", stateCommand))
		}
	})

	fanContainer := createSlider(
		"Fan",
		func(f float64) {
			fmt.Printf("Set Fan: %.0f\n", f)
			w.Write(fmt.Appendf([]byte{}, "F%.0f\n", f))
			lastEventTimer.Set(time.Now())
		},
		func(value int) {
			fmt.Printf("Fixing fan: %d\n", value)
			w.Write(fmt.Appendf([]byte{}, "f%d\n", value))
		},
	)

	powerContainer := createSlider(
		"Power",
		func(f float64) {
			fmt.Printf("Set Power: %.0f\n", f)
			w.Write(fmt.Appendf([]byte{}, "P%.0f\n", f))
			lastEventTimer.Set(time.Now())
		},
		func(value int) {
			fmt.Printf("Fixing power: %d\n", value)
			w.Write(fmt.Appendf([]byte{}, "p%d\n", value))
		},
	)

	logAccordion, logEntry := createLogAccordion()
	ui.logEntry = logEntry

	contentContainer := container.NewVBox(
		container.NewHBox(
			container.NewPadded(overallTimer.text),
			container.NewPadded(lastEventTimer.text),
			layout.NewSpacer(),
			container.NewPadded(fcTimer.text),
		),
		stateButton,
		fanContainer,
		powerContainer,
		logAccordion,
	)

	go func() {
		<-ctx.Done()
		fyne.Do(func() {
			application.Quit()
		})
	}()

	window.SetContent(contentContainer)
	window.Resize(fyne.NewSize(300, 200))
	window.ShowAndRun()
}

// Write implements io.Writer to enable writing logs to the log entry
func (ui *RoasterUI) Write(p []byte) (n int, err error) {
	if ui.logEntry == nil {
		return len(p), nil
	}

	text := string(p)

	fyne.Do(func() {
		ui.logEntry.Append(text)
		ui.logEntry.CursorRow = len(ui.logEntry.Text) // auto-scroll
	})

	return len(p), nil
}
