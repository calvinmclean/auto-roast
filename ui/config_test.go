package ui

import "testing"

func TestRoastFileDisplay(t *testing.T) {
	for path, want := range map[string]string{
		"":                                "No replay file selected",
		"ROAST.txt":                       "ROAST.txt",
		"/Users/test/Downloads/ROAST.txt": ".../Downloads/ROAST.txt",
	} {
		t.Run(path, func(t *testing.T) {
			if got := roastFileDisplay(path); got != want {
				t.Errorf("roastFileDisplay(%q) = %q, want %q", path, got, want)
			}
		})
	}
}
