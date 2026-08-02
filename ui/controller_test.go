package ui

import (
	"bytes"
	"testing"
)

func TestControllerWrapperIncreaseTime(t *testing.T) {
	var output bytes.Buffer
	c := controllerWrapper{writer: &output}

	c.IncreaseTime()

	if got, want := output.String(), "T\n"; got != want {
		t.Errorf("IncreaseTime() wrote %q, want %q", got, want)
	}
}
