package ui

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/calvinmclean/autoroast/controller"
)

type ConfigWindow struct {
	app      fyne.App
	OnSubmit func()
}

func NewConfigWindow(app fyne.App) *ConfigWindow {
	return &ConfigWindow{
		app: app,
	}
}

func (cw *ConfigWindow) loadConfigFromPreferences(cfg *controller.Config) {
	prefs := cw.app.Preferences()
	cfg.SerialPort = prefs.StringWithFallback("serialPort", "")
	cfg.BaudRate = prefs.StringWithFallback("baudRate", "115200")
	cfg.TWChartAddr = prefs.StringWithFallback("twchartAddr", "")
	cfg.SessionName = prefs.StringWithFallback("sessionName", "")
	cfg.ProbesInput = prefs.StringWithFallback("probesInput", "1=Ambient,2=Beans")
}

func (cw *ConfigWindow) saveConfigToPreferences(cfg *controller.Config) {
	prefs := cw.app.Preferences()
	prefs.SetString("serialPort", cfg.SerialPort)
	prefs.SetString("baudRate", cfg.BaudRate)
	prefs.SetString("twchartAddr", cfg.TWChartAddr)
	prefs.SetString("sessionName", cfg.SessionName)
	prefs.SetString("probesInput", cfg.ProbesInput)
}

func (cw *ConfigWindow) Show(cfg *controller.Config) {
	window := cw.app.NewWindow("Auto Roast - Configuration")
	window.Resize(fyne.NewSize(400, 250))
	window.SetCloseIntercept(func() {
		// Treat window close as cancel
		window.Close()
		cw.app.Quit()
	})
	window.Show()

	// Load config from preferences
	cw.loadConfigFromPreferences(cfg)

	serialPorts, err := controller.GetSerialPorts()
	if err != nil && !errors.Is(err, controller.ErrNoUSBSerial) {
		showError(cw.app, window, fmt.Errorf("error getting serial ports: %w", err))
		return
	}

	serialPorts = append(serialPorts, controller.SerialPortNone)

	serialEntry := widget.NewSelect(serialPorts, nil)
	if cfg.SerialPort == "" {
		cfg.SerialPort = serialPorts[0]
	}
	serialEntry.Bind(binding.BindString(&cfg.SerialPort))

	sessionEntry := widget.NewEntry()
	sessionEntry.Bind(binding.BindString(&cfg.SessionName))

	probesEntry := widget.NewEntry()
	probesEntry.Bind(binding.BindString(&cfg.ProbesInput))

	baudRateEntry := widget.NewEntry()
	baudRateEntry.Bind(binding.BindString(&cfg.BaudRate))

	twchartAddrEntry := widget.NewEntry()
	twchartAddrEntry.Bind(binding.BindString(&cfg.TWChartAddr))

	submitButton := widget.NewButton("Submit", func() {
		cw.saveConfigToPreferences(cfg)
		cw.OnSubmit()
		window.Close()
	})
	submitButton.Disable()

	validateForm := func() {
		allFieldsValid := cfg.SerialPort != "" &&
			cfg.SessionName != "" &&
			cfg.ProbesInput != "" &&
			cfg.BaudRate != "" &&
			cfg.TWChartAddr != ""

		if allFieldsValid {
			submitButton.Enable()
		}
	}

	// Add listeners to field changes
	serialEntry.OnChanged = func(_ string) { validateForm() }
	sessionEntry.OnChanged = func(_ string) { validateForm() }
	probesEntry.OnChanged = func(_ string) { validateForm() }
	baudRateEntry.OnChanged = func(_ string) { validateForm() }
	twchartAddrEntry.OnChanged = func(_ string) { validateForm() }

	// Initial validation
	validateForm()

	form := container.NewVBox(
		widget.NewCard("Configuration", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("Serial Port:"),
				serialEntry,
			),
			container.NewGridWithColumns(2,
				widget.NewLabel("Baud Rate:"),
				baudRateEntry,
			),
			container.NewGridWithColumns(2,
				widget.NewLabel("TWChart Address:"),
				twchartAddrEntry,
			),
			container.NewGridWithColumns(2,
				widget.NewLabel("Session Name:"),
				sessionEntry,
			),
			container.NewGridWithColumns(2,
				widget.NewLabel("Probes Input:"),
				probesEntry,
			),
		)),
		container.NewHBox(
			widget.NewButton("Cancel", func() {
				window.Close()
				cw.app.Quit()
			}),
			submitButton,
		),
	)

	window.SetContent(form)
}

func showError(app fyne.App, window fyne.Window, err error) {
	d := dialog.NewError(err, window)
	d.SetOnClosed(func() {
		app.Quit()
	})
	d.Show()
}
