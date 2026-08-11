package ui

import (
	"bytes"
	"testing"

	"fyne.io/fyne/v2/test"
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

func TestConfigWindowPersistsRoastFile(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	configWindow := NewConfigWindow(app)
	configWindow.saveConfigToPreferences(&controller.Config{RoastFile: "/tmp/roast.txt"})

	var cfg controller.Config
	configWindow.loadConfigFromPreferences(&cfg)
	if got, want := cfg.RoastFile, "/tmp/roast.txt"; got != want {
		t.Errorf("RoastFile = %q, want %q", got, want)
	}
}
