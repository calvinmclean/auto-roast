package ui

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

func alertHandler(window fyne.Window) func(message string) {
	return func(message string) {
		done := make(chan struct{})
		fyne.Do(func() {
			d := dialog.NewInformation("Roast Alert", message, window)
			d.SetOnClosed(func() { close(done) })
			d.Resize(fyne.NewSize(500, 250))
			d.Show()
		})
		<-done
	}
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

	cw := &controllerWrapper{}

	var stateButton *widget.Button
	refreshStateButton := func() {
		if currentState == stateDone {
			stateButton.SetText("Done")
			stateButton.Disable()
			return
		}
		stateButton.SetText(currentState.next().String())
		stateButton.Enable()
	}
	advanceState := func() {
		currentState = currentState.next()

		lastEventTimer.Set(time.Now())
		refreshStateButton()

		switch currentState {
		case stateRoasting:
			// reset the timer when roasting starts
			overallTimer.Set(time.Now())
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
			overallTimer.Stop()
			lastEventTimer.Stop()
		}
	}
	stateButton = widget.NewButton(currentState.next().String(), func() {
		stateButton.Disable()
		cw.RunStateCommand(currentState.next())
	})
	var setFanSlider, setPowerSlider func(float64)
	applyCommand := func(command string) {
		fyne.Do(func() {
			if target := stateForCommand(command); target != stateNone {
				if currentState.next() == target {
					advanceState()
				} else {
					refreshStateButton()
				}
			}
			if setting, value, ok := settingValue(command); ok {
				lastEventTimer.Set(time.Now())
				if setting == 'F' {
					setFanSlider(value)
				} else {
					setPowerSlider(value)
				}
			}
		})
	}

	fanContainer, setFanSlider := createSlider(
		"Fan",
		cw.SetFan,
		cw.FixFan,
		window.Canvas().Focus,
	)

	powerContainer, setPowerSlider := createSlider(
		"Power",
		cw.SetPower,
		cw.FixPower,
		window.Canvas().Focus,
	)

	logAccordion, logEntry := createLogAccordion()
	ui.logEntry = logEntry

	var replay *controller.Replay
	var replayQueueItems []controller.ReplayQueuedAction
	var replayQueue *widget.List
	var replayButton *widget.Button
	var skipReplayButton *widget.Button
	replayStatus := widget.NewLabel("Manual control")
	replayStatus.Wrapping = fyne.TextWrapWord
	waitCountdown := widget.NewLabel("")
	waitCountdown.Hide()
	var waitCountdownCancel context.CancelFunc
	var waitCountdownID uint
	updateWaitCountdown := func(waitUntil time.Time) {
		waitCountdownID++
		id := waitCountdownID
		if waitCountdownCancel != nil {
			waitCountdownCancel()
		}
		if waitUntil.IsZero() {
			waitCountdown.SetText("")
			waitCountdown.Hide()
			return
		}

		waitCountdown.Show()
		countdownCtx, cancelCountdown := context.WithCancel(ctx)
		waitCountdownCancel = cancelCountdown
		go func() {
			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()
			for {
				remaining := time.Until(waitUntil)
				fyne.Do(func() {
					if id == waitCountdownID {
						waitCountdown.SetText(formatWaitRemaining(remaining))
					}
				})
				if remaining <= 0 {
					return
				}

				select {
				case <-countdownCtx.Done():
					return
				case <-ticker.C:
				}
			}
		}()
	}
	replayQueue = widget.NewList(
		func() int { return len(replayQueueItems) },
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Truncation = fyne.TextTruncateEllipsis
			handle := newDragHandle(func(itemID, _, to int) {
				if replay != nil {
					replay.MoveQueuedTo(itemID, to)
				}
			})
			removeButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			return container.NewBorder(nil, nil, handle, removeButton, label)
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			item := replayQueueItems[id]
			row := object.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(item.Text)
			handle := row.Objects[1].(*dragHandle)
			handle.itemID = item.ID
			handle.index = id
			removeButton := row.Objects[2].(*widget.Button)
			removeButton.Enable()
			removeButton.OnTapped = func() {
				if replay != nil && replay.RemoveQueued(item.ID) {
					return
				}

				for i, queued := range replayQueueItems {
					if queued.ID == item.ID {
						replayQueueItems = append(replayQueueItems[:i], replayQueueItems[i+1:]...)
						replayQueue.Refresh()
						if len(replayQueueItems) == 0 {
							replayStatus.SetText("No planned actions remaining.")
							replayButton.Disable()
						}
						return
					}
				}
			}
			row.Refresh()
		},
	)
	addReplayEntry := widget.NewEntry()
	addReplayEntry.SetPlaceHolder("Command or WAIT 30s")
	addReplayAction := func() {
		if replay == nil || addReplayEntry.Text == "" {
			return
		}
		if err := replay.AddQueued(addReplayEntry.Text); err != nil {
			replayStatus.SetText("Invalid action: " + err.Error())
			return
		}
		addReplayEntry.SetText("")
	}
	addReplayEntry.OnSubmitted = func(string) { addReplayAction() }
	addReplayButton := widget.NewButton("+", addReplayAction)
	var cancelReplay context.CancelFunc
	var startReplay func()
	replayButton = widget.NewButton("Start Planned Roast", func() {
		if cancelReplay != nil {
			cancelReplay()
			return
		}
		if startReplay != nil {
			startReplay()
		}
	})
	replayButton.Disable()
	skipReplayButton = widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), func() {
		if replay != nil {
			replay.Skip()
		}
	})
	skipReplayButton.Hide()

	clickButton := widget.NewButton("Click", func() {
		cw.Click()
	})
	debugButton := widget.NewButton("Debug", func() {
		cw.Debug()
	})
	increaseTimeButton := widget.NewButton("Increase Time", func() {
		cw.IncreaseTime()
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
		clickButton,
		debugButton,
		increaseTimeButton,
	)

	manualControls := container.NewVBox(
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
	)
	replayControls := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Planned Roast"),
			container.NewBorder(nil, nil, nil, waitCountdown, replayStatus),
			container.NewHBox(replayButton, skipReplayButton),
			container.NewBorder(nil, nil, nil, addReplayButton, addReplayEntry),
		),
		nil,
		nil,
		nil,
		replayQueue,
	)
	roastSplit := container.NewHSplit(manualControls, replayControls)
	roastSplit.SetOffset(0.45)
	contentContainer := container.NewBorder(
		nil,
		logAccordion,
		nil,
		nil,
		roastSplit,
	)

	go func() {
		<-ctx.Done()
		fyne.Do(func() {
			application.Quit()
		})
	}()

	window.SetContent(contentContainer)
	window.Resize(fyne.NewSize(880, 480))

	// Show config window on startup
	configWindow := NewConfigWindow(application)
	configWindow.OnSubmit = func() {
		defer window.Show()

		replay = nil
		replayQueueItems = nil
		startReplay = nil
		replayButton.Disable()
		if cfg.RoastFile != "" {
			var err error
			actions, err := controller.LoadReplay(cfg.RoastFile)
			if err != nil {
				showError(application, window, fmt.Errorf("error loading replay file: %w", err))
				return
			}
			replay = controller.NewReplay(actions, func(state controller.ReplayState) {
				fyne.Do(func() {
					replayQueueItems = state.Queued
					replayQueue.Refresh()
					updateWaitCountdown(state.WaitUntil)
					switch {
					case state.Running && state.Current != "":
						replayStatus.SetText("Current: " + state.Current)
						replayButton.SetText("Cancel Planned Roast")
						replayButton.Enable()
						if strings.HasPrefix(state.Current, "WAIT") {
							skipReplayButton.Show()
						} else {
							skipReplayButton.Hide()
						}
					case state.Running:
						replayStatus.SetText("Planned roast starting")
						replayButton.SetText("Cancel Planned Roast")
						replayButton.Enable()
						skipReplayButton.Hide()
					case state.Cancelled:
						cancelReplay = nil
						replayStatus.SetText("Planned roast cancelled. Manual control enabled.")
						replayButton.SetText("Start Planned Roast")
						replayButton.Disable()
						skipReplayButton.Hide()
						refreshStateButton()
					case !state.Started && len(state.Queued) > 0:
						replayStatus.SetText("Planned roast ready. Manual control remains available.")
						replayButton.SetText("Start Planned Roast")
						replayButton.Enable()
						skipReplayButton.Hide()
					case !state.Started:
						replayStatus.SetText("No planned actions remaining.")
						replayButton.SetText("Start Planned Roast")
						replayButton.Disable()
						skipReplayButton.Hide()
					default:
						cancelReplay = nil
						replayStatus.SetText("Planned roast complete. Manual control enabled.")
						replayButton.SetText("Start Planned Roast")
						replayButton.Disable()
						skipReplayButton.Hide()
						refreshStateButton()
					}
				})
			}, alertHandler(window))
			replayQueueItems = replay.State().Queued
			replayQueue.Refresh()
			replayStatus.SetText("Planned roast ready. Manual control remains available.")
			replayButton.SetText("Start Planned Roast")
			replayButton.Enable()
		}

		setFanSlider(float64(cfg.InitialFanSetting))
		setPowerSlider(float64(cfg.InitialPowerSetting))

		c, err := controller.New(cfg)
		if err != nil {
			showError(application, window, fmt.Errorf("error creating controller: %w", err))
			return
		}

		commandReader, commandWriter := io.Pipe()
		controllerReader, controllerInputWriter := io.Pipe()
		cw.writer = commandWriter

		var controllerOutput io.Writer = ui
		if debug {
			// read/write Stdin/Stdout also
			go func() {
				defer commandWriter.Close()
				io.Copy(commandWriter, os.Stdin)
			}()

			controllerOutput = io.MultiWriter(os.Stdout, controllerOutput)
		}

		controllerCtx, cancel := context.WithCancel(ctx)
		go func() {
			err := runCommands(commandReader, controllerInputWriter, applyCommand)
			if err != nil {
				fyne.Do(func() {
					showError(application, window, fmt.Errorf("error processing UI command: %w", err))
				})
			}
		}()
		go func() {
			err := c.Run(controllerCtx, controllerReader, controllerOutput)
			if err != nil {
				fyne.Do(func() {
					showError(application, window, fmt.Errorf("error running controller: %w", err))
				})
				return
			}
		}()

		if replay != nil {
			startReplay = func() {
				replayCtx, replayCancel := context.WithCancel(controllerCtx)
				cancelReplay = replayCancel
				replayButton.SetText("Cancel Planned Roast")
				go func() {
					err := replay.Run(replayCtx, commandWriter)
					if err != nil {
						fyne.Do(func() {
							cancelReplay = nil
							replayStatus.SetText("Planned roast failed. Manual control enabled.")
							replayButton.Disable()
							showError(application, window, fmt.Errorf("error running replay: %w", err))
						})
					}
				}()
			}
		}

		window.SetOnClosed(func() {
			if cancelReplay != nil {
				cancelReplay()
			}
			if waitCountdownCancel != nil {
				waitCountdownCancel()
			}
			_ = commandWriter.Close()
			cancel()
			_ = c.Close()
		})
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

func createSlider(labelText string, onSet func(float64), onFix func(int), setFocus func(fyne.Focusable)) (*fyne.Container, func(float64)) {
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

	return container, func(f float64) {
		slider.Value = f
		slider.OnChanged(f)
		slider.Refresh()
	}
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

func formatWaitRemaining(remaining time.Duration) string {
	if remaining < 0 {
		remaining = 0
	}

	seconds := int(math.Ceil(remaining.Seconds()))
	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}
