package ui

import (
	"context"
	"fmt"
	"image/color"
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

type State int

const (
	StateNone State = iota
	StateStart
	StatePreheat
	StateRoasting
	StateFirstCrack
	StateCooling
	StateDone
)

func (s State) String() string {
	switch s {
	case StateStart:
		return "Start"
	case StatePreheat:
		return "Preheat"
	case StateRoasting:
		return "Roasting"
	case StateFirstCrack:
		return "First Crack"
	case StateCooling:
		return "Cooling"
	case StateDone:
		return "Done"
	default:
		return "Unknown"
	}
}

func (s State) Next() State {
	if s == StateDone {
		// Done has no next State
		return StateDone
	}
	return s + 1
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

func createLogAccordion() *widget.Accordion {
	logContent := widget.NewLabel("")
	logScroll := container.NewVScroll(logContent)
	logScroll.SetMinSize(fyne.NewSize(300, 100))

	go func() {
		for range time.Tick(time.Second) {
			fyne.Do(func() {
				logLine := fmt.Sprintf("Mock Log Entry")
				logContent.SetText(logContent.Text + "\n" + logLine)
			})
		}
	}()

	return widget.NewAccordion(
		widget.NewAccordionItem("Logs", logScroll),
	)
}

type RoasterUI struct {
}

func NewRoasterUI() *RoasterUI {
	return &RoasterUI{}
}

func (ui *RoasterUI) Run(ctx context.Context) {
	application := app.New()

	window := application.NewWindow("Roasting App")

	currentState := StateNone

	overallTimer := newTimer(false)
	lastEventTimer := newTimer(true)
	fcTimer := newTimer(true)

	waitForStart := make(chan struct{})
	overallTimer.Go(waitForStart)
	lastEventTimer.Go(waitForStart)

	waitForFC := make(chan struct{})
	fcTimer.Go(waitForFC)

	var stateButton *widget.Button
	stateButton = widget.NewButton(currentState.Next().String(), func() {
		currentState++

		lastEventTimer.Set(time.Now())

		if currentState == StateFirstCrack {
			fcTimer.text.Color = color.RGBA{R: 139, G: 0, B: 0, A: 255}
			fcTimer.Set(time.Now())
			close(waitForFC)
		}

		if currentState == StateFirstCrack+1 {
			fcTimer.Stop()
		}

		if currentState == StateStart {
			overallTimer.Set(time.Now())
			close(waitForStart)
		}

		stateButton.SetText(currentState.Next().String())
		if currentState == StateDone {
			stateButton.Disable()
		}
	})

	fanContainer := createSlider(
		"Fan",
		func(f float64) {
			fmt.Printf("Set Fan: %.0f\n", f)
		},
		func(value int) {
			fmt.Printf("Fixing fan: %d\n", value)
		},
	)

	powerContainer := createSlider(
		"Power",
		func(f float64) {
			fmt.Printf("Set Power: %.0f\n", f)
		},
		func(value int) {
			fmt.Printf("Fixing power: %d\n", value)
		},
	)

	logAccordion := createLogAccordion()

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
