package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

// Bundled from https://github.com/googlefonts/atkinson-hyperlegible-next.
// See fonts/OFL.txt for the SIL Open Font License.
//
//go:embed fonts/AtkinsonHyperlegibleNext-Regular.ttf
var atkinsonRegularData []byte

//go:embed fonts/AtkinsonHyperlegibleNext-Bold.ttf
var atkinsonBoldData []byte

//go:embed fonts/AtkinsonHyperlegibleNext-Italic.ttf
var atkinsonItalicData []byte

//go:embed fonts/AtkinsonHyperlegibleNext-BoldItalic.ttf
var atkinsonBoldItalicData []byte

//go:embed fonts/AtkinsonHyperlegibleMono-Regular.ttf
var atkinsonMonoData []byte

//go:embed fonts/AtkinsonHyperlegibleMono-Bold.ttf
var atkinsonMonoBoldData []byte

var (
	atkinsonMono       = fyne.NewStaticResource("AtkinsonHyperlegibleMono-Regular.ttf", atkinsonMonoData)
	atkinsonMonoBold   = fyne.NewStaticResource("AtkinsonHyperlegibleMono-Bold.ttf", atkinsonMonoBoldData)
	atkinsonRegular    = fyne.NewStaticResource("AtkinsonHyperlegibleNext-Regular.ttf", atkinsonRegularData)
	atkinsonBold       = fyne.NewStaticResource("AtkinsonHyperlegibleNext-Bold.ttf", atkinsonBoldData)
	atkinsonItalic     = fyne.NewStaticResource("AtkinsonHyperlegibleNext-Italic.ttf", atkinsonItalicData)
	atkinsonBoldItalic = fyne.NewStaticResource("AtkinsonHyperlegibleNext-BoldItalic.ttf", atkinsonBoldItalicData)
)

func (t *roastTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Symbol {
		return t.Theme.Font(style)
	}
	if style.Monospace {
		if style.Bold {
			return atkinsonMonoBold
		}
		return atkinsonMono
	}
	if style.Bold && style.Italic {
		return atkinsonBoldItalic
	}
	if style.Bold {
		return atkinsonBold
	}
	if style.Italic {
		return atkinsonItalic
	}
	return atkinsonRegular
}
