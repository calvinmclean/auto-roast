package ui

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/calvinmclean/autoroast"
	"github.com/calvinmclean/autoroast/controller"
)

// TODO: Add Note inputs

type RoasterUI struct {
	// logEntry is used as the target for writing to RoasterUI
	logEntry *widget.Entry
}

func NewRoasterUI() *RoasterUI {
	return &RoasterUI{}
}

func (ui *RoasterUI) Run(ctx context.Context, cfg controller.Config, debug bool) {
	application := app.NewWithID("auto.roast.calvinmclean.github.io")

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

	cw := &controllerWrapper{lastEventTimer: lastEventTimer}

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

		cw.RunStateCommand(currentState)
	})

	fanContainer := createSlider(
		"Fan",
		cw.SetFan,
		cw.FixFan,
		window.Canvas().Focus,
	)

	powerContainer := createSlider(
		"Power",
		cw.SetPower,
		cw.FixPower,
		window.Canvas().Focus,
	)

	logAccordion, logEntry := createLogAccordion()
	ui.logEntry = logEntry

	cButton := widget.NewButton("Click", func() {
		cw.Click()
	})
	dButton := widget.NewButton("Debug", func() {
		cw.Debug()
	})

	noteEntry := widget.NewEntry()
	noteEntry.OnSubmitted = func(s string) {
		if s == "" {
			return
		}
		cw.Note(s)
		noteEntry.SetText("")
	}
	noteButton := widget.NewButtonWithIcon("", theme.ConfirmIcon(), func() {
		noteEntry.OnSubmitted(noteEntry.Text)
	})

	buttonContainer := container.NewGridWithColumns(3,
		cButton,
		dButton,
	)

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
		container.NewBorder(nil, nil, nil, noteButton, noteEntry),
		buttonContainer,
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

	// Show config window on startup
	configWindow := NewConfigWindow(application)
	configWindow.OnSubmit = func() {
		c, err := controller.New(cfg)
		if err != nil {
			showError(application, window, fmt.Errorf("error creating controller: %w", err))
			return
		}

		r, w := io.Pipe()
		cw.writer = w

		var controllerWriter io.Writer = ui
		if debug {
			// read/write Stdin/Stdout also
			go func() {
				defer w.Close()
				io.Copy(w, os.Stdin)
			}()

			controllerWriter = io.MultiWriter(os.Stdout, controllerWriter)
		}

		controllerCtx, cancel := context.WithCancel(ctx)
		go func() {
			err := c.Run(controllerCtx, r, controllerWriter)
			if err != nil {
				showError(application, window, fmt.Errorf("error running controller: %w", err))
				return
			}
		}()

		window.SetOnClosed(func() {
			cancel()
			_ = c.Close()
		})

		window.Show()
	}
	configWindow.Show(&cfg)

	application.Run()
}

// Write implements io.Writer to enable writing logs to the log entry
func (ui *RoasterUI) Write(p []byte) (n int, err error) {
	if ui.logEntry == nil {
		return len(p), nil
	}

	// clean up extra newlines and termination character
	clean := p[:0]
	for _, v := range p {
		if v != '\n' && v != autoroast.TerminationChar {
			clean = append(clean, v)
		}
	}
	if len(clean) == 0 {
		return len(p), nil
	}
	clean = append(clean, '\n')

	text := string(clean)

	fyne.Do(func() {
		ui.logEntry.Append(text)
		ui.logEntry.CursorRow = len(ui.logEntry.Text) // auto-scroll
	})

	return len(p), nil
}

func createSlider(labelText string, onSet func(float64), onFix func(int), setFocus func(fyne.Focusable)) *fyne.Container {
	defaultValue := 1.0
	valueLabel := widget.NewLabel(fmt.Sprintf("%.0f", defaultValue))

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

	slider := widget.NewSlider(1, 9)
	slider.Step = 1
	slider.SetValue(defaultValue)
	slider.OnChanged = func(value float64) {
		valueLabel.SetText(fmt.Sprintf("%.0f", value))
	}
	slider.OnChangeEnded = func(f float64) {
		onSet(f)
		setFocus(fixNumberEntry)
	}

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
