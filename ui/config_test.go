package ui

import "testing"

func TestRoastFileDisplay(t *testing.T) {
	for path, want := range map[string]string{
		"":                                  "No replay file selected",
		"ROAST.roast":                       "ROAST.roast",
		"/Users/test/Downloads/ROAST.roast": ".../Downloads/ROAST.roast",
	} {
		t.Run(path, func(t *testing.T) {
			if got := roastFileDisplay(path); got != want {
				t.Errorf("roastFileDisplay(%q) = %q, want %q", path, got, want)
			}
		})
	}
}
