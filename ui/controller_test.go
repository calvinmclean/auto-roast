package ui

import (
	"bytes"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/calvinmclean/autoroast/controller"
)

func TestControllerWrapperIncreaseTime(t *testing.T) {
	var output bytes.Buffer
	c := controllerWrapper{writer: &output}

	c.IncreaseTime()

	if got, want := output.String(), "T\n"; got != want {
		t.Errorf("IncreaseTime() wrote %q, want %q", got, want)
	}
}

func TestControllerWrapperIgnoresCommandsBeforeSetup(t *testing.T) {
	c := controllerWrapper{}

	c.SetFan(5)
	c.SetPower(5)
}

func TestConfigWindowPersistsRoastFile(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	configWindow := NewConfigWindow(app)
	configWindow.saveConfigToPreferences(&controller.Config{RoastFile: "/tmp/roast.roast"})

	var cfg controller.Config
	configWindow.loadConfigFromPreferences(&cfg)
	if got, want := cfg.RoastFile, "/tmp/roast.roast"; got != want {
		t.Errorf("RoastFile = %q, want %q", got, want)
	}
}

func TestCreateSliderSetterUpdatesSlider(t *testing.T) {
	setCalls := 0
	container, setValue := createSlider("Fan", func(float64) { setCalls++ }, func(int) {}, func(fyne.Focusable) {})
	slider := container.Objects[1].(*widget.Slider)

	setValue(5)
	if got, want := slider.Value, 5.0; got != want {
		t.Errorf("slider.Value = %v, want %v", got, want)
	}
	if setCalls != 0 {
		t.Errorf("programmatic slider update submitted %d commands, want 0", setCalls)
	}
}
