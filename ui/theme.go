package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	roastBackground = color.NRGBA{R: 25, G: 23, B: 22, A: 255}
	roastSurface    = color.NRGBA{R: 37, G: 33, B: 30, A: 255}
	roastForeground = color.NRGBA{R: 244, G: 235, B: 222, A: 255}
	roastMuted      = color.NRGBA{R: 184, G: 170, B: 154, A: 255}
	roastCopper     = color.NRGBA{R: 225, G: 158, B: 108, A: 255}
)

// roastTheme keeps the instrument palette consistent with either system appearance.
type roastTheme struct{ fyne.Theme }

func newRoastTheme() fyne.Theme { return &roastTheme{Theme: theme.DefaultTheme()} }

func (t *roastTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground, theme.ColorNameInputBackground:
		return roastBackground
	case theme.ColorNameButton:
		return color.NRGBA{R: 83, G: 67, B: 55, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 49, G: 43, B: 38, A: 255}
	case theme.ColorNameHeaderBackground, theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return roastSurface
	case theme.ColorNameForeground:
		return roastForeground
	case theme.ColorNamePrimary, theme.ColorNameHyperlink:
		return roastCopper
	case theme.ColorNameForegroundOnPrimary:
		return roastBackground
	case theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		return roastMuted
	case theme.ColorNameInputBorder, theme.ColorNameSeparator:
		return color.NRGBA{R: 85, G: 73, B: 63, A: 255}
	case theme.ColorNameHover:
		return color.NRGBA{R: 225, G: 158, B: 108, A: 24}
	case theme.ColorNameFocus, theme.ColorNameSelection, theme.ColorNamePressed:
		return color.NRGBA{R: 225, G: 158, B: 108, A: 65}
	default:
		return t.Theme.Color(name, theme.VariantDark)
	}
}

func (t *roastTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInputRadius, theme.SizeNameSelectionRadius:
		return 6
	default:
		return t.Theme.Size(name)
	}
}

func roastPanel(content fyne.CanvasObject) *fyne.Container {
	background := canvas.NewRectangle(roastSurface)
	background.CornerRadius = 10
	return container.NewStack(background, container.NewPadded(content))
}

func sectionTitle(text string) *widget.Label {
	return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
}

func timerTile(label string, timer *timer, prominent bool) *fyne.Container {
	timer.text.Color = roastForeground
	timer.text.TextStyle = fyne.TextStyle{Monospace: true}
	timer.text.TextSize = 20
	if prominent {
		timer.text.Color = roastCopper
		timer.text.TextSize = 32
	}
	if label == "" {
		return roastPanel(container.NewCenter(timer.text))
	}
	caption := canvas.NewText(label, roastMuted)
	caption.TextSize = 14
	caption.TextStyle = fyne.TextStyle{Bold: true}
	return roastPanel(container.NewVBox(caption, timer.text))
}

// sizedTextTheme changes label size within a specific container.
type sizedTextTheme struct {
	fyne.Theme
	textSize float32
}

func (t *sizedTextTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return t.textSize
	}
	return t.Theme.Size(name)
}

func replayStatusCard(status, countdown *widget.Label) (*fyne.Container, *fyne.Container) {
	status.TextStyle = fyne.TextStyle{Bold: true}
	countdown.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	countdown.Importance = widget.HighImportance
	remaining := container.NewVBox(
		container.NewThemeOverride(countdown, &sizedTextTheme{Theme: newRoastTheme(), textSize: 20}),
	)
	remaining.Hide()
	background := canvas.NewRectangle(roastBackground)
	background.CornerRadius = 8
	card := container.NewStack(background, container.NewPadded(container.NewBorder(
		nil, nil, nil, remaining,
		container.NewThemeOverride(status, &sizedTextTheme{Theme: newRoastTheme(), textSize: 14}),
	)))
	return card, remaining
}
