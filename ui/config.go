package ui

import (
	"errors"
	"fmt"
	"strconv"

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
	cfg.InitialFanSetting = prefs.IntWithFallback("initialFanSetting", 5)
	cfg.InitialPowerSetting = prefs.IntWithFallback("initialPowerSetting", 5)
}

func (cw *ConfigWindow) saveConfigToPreferences(cfg *controller.Config) {
	prefs := cw.app.Preferences()
	prefs.SetString("serialPort", cfg.SerialPort)
	prefs.SetString("baudRate", cfg.BaudRate)
	prefs.SetString("twchartAddr", cfg.TWChartAddr)
	prefs.SetString("sessionName", cfg.SessionName)
	prefs.SetString("probesInput", cfg.ProbesInput)
	prefs.SetInt("initialFanSetting", cfg.InitialFanSetting)
	prefs.SetInt("initialPowerSetting", cfg.InitialPowerSetting)
}

func (cw *ConfigWindow) Show(cfg *controller.Config) {
	window := cw.app.NewWindow("Auto Roast - Configuration")
	window.Resize(fyne.NewSize(450, 300))
	window.SetCloseIntercept(func() {
		// Treat window close as cancel
		window.Close()
		cw.app.Quit()
	})
	window.Show()

	// Load config from preferences
	cw.loadConfigFromPreferences(cfg)

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
			cfg.TWChartAddr != "" &&
			cfg.InitialFanSetting >= 0 && cfg.InitialFanSetting <= 9 &&
			cfg.InitialPowerSetting >= 0 && cfg.InitialPowerSetting <= 9

		if allFieldsValid {
			submitButton.Enable()
		}
	}

	serialPorts, err := controller.GetSerialPorts()
	if err != nil && !errors.Is(err, controller.ErrNoUSBSerial) {
		showError(cw.app, window, fmt.Errorf("error getting serial ports: %w", err))
		return
	}

	serialPorts = append(serialPorts, controller.SerialPortNone)

	serialEntry := widget.NewSelect(serialPorts, func(s string) {
		validateForm()
		cfg.SerialPort = s
	})
	if cfg.SerialPort == "" {
		cfg.SerialPort = serialPorts[0]
	}
	serialEntry.SetSelected(serialPorts[0])

	sessionEntry := widget.NewEntry()
	sessionEntry.Bind(binding.BindString(&cfg.SessionName))

	probesEntry := widget.NewEntry()
	probesEntry.Bind(binding.BindString(&cfg.ProbesInput))

	baudRateEntry := widget.NewEntry()
	baudRateEntry.Bind(binding.BindString(&cfg.BaudRate))

	twchartAddrEntry := widget.NewEntry()
	twchartAddrEntry.Bind(binding.BindString(&cfg.TWChartAddr))

	fanEntry := widget.NewSelect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, func(s string) {
		if fan, err := strconv.Atoi(s); err == nil {
			cfg.InitialFanSetting = fan
			validateForm()
		}
	})
	fanEntry.SetSelected(strconv.Itoa(cfg.InitialFanSetting))
	fanEntry.Resize(fyne.NewSize(60, fanEntry.MinSize().Height))

	powerEntry := widget.NewSelect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, func(s string) {
		if power, err := strconv.Atoi(s); err == nil {
			cfg.InitialPowerSetting = power
			validateForm()
		}
	})
	powerEntry.SetSelected(strconv.Itoa(cfg.InitialPowerSetting))
	powerEntry.Resize(fyne.NewSize(60, powerEntry.MinSize().Height))

	initSettingsEntries := container.NewHBox(fanEntry, powerEntry)
	initSettingsEntries.Resize(fyne.NewSize(120, initSettingsEntries.MinSize().Height))

	// Add listeners to field changes
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
			container.NewGridWithColumns(2,
				widget.NewLabel("Initial Fan/Power:"),
				container.NewWithoutLayout(initSettingsEntries),
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
