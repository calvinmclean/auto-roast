package ui

import (
	"testing"

	"fyne.io/fyne/v2/theme"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// Equal digit advances keep timer values stable without a monospace face.
func TestTimerFontHasEqualDigitWidths(t *testing.T) {
	face, err := sfnt.Parse(theme.DefaultTextFont().Content())
	if err != nil {
		t.Fatal(err)
	}
	var buffer sfnt.Buffer
	var width fixed.Int26_6
	for digit := '0'; digit <= '9'; digit++ {
		glyph, err := face.GlyphIndex(&buffer, digit)
		if err != nil {
			t.Fatal(err)
		}
		advance, err := face.GlyphAdvance(&buffer, glyph, fixed.I(32), font.HintingNone)
		if err != nil {
			t.Fatal(err)
		}
		if digit == '0' {
			width = advance
		} else if advance != width {
			t.Errorf("digit %c width = %v, want %v", digit, advance, width)
		}
	}
}
