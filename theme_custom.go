// theme_custom.go
package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// MediumTheme wraps DarkTheme so all interactive widgets (Select, Label, Button)
// naturally render with white text in every state (idle, hover, focused).
// The board cells are explicitly colored so they are unaffected by the app theme.
type MediumTheme struct {
	base fyne.Theme
}

func NewMediumTheme() fyne.Theme {
	return &MediumTheme{base: theme.DarkTheme()}
}

func (m *MediumTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	// Do NOT override the variant — let dark theme manage text/fg colors natively.
	switch n {
	case theme.ColorNameBackground:
		return color.NRGBA{40, 42, 48, 255} // dark grey app background
	case theme.ColorNameFocus:
		return color.NRGBA{0, 114, 206, 255}
	case theme.ColorNameHover:
		return color.NRGBA{255, 255, 255, 18} // very subtle light overlay on dark bg
	}
	return m.base.Color(n, v)
}

func (m *MediumTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return m.base.Icon(n) }
func (m *MediumTheme) Font(s fyne.TextStyle) fyne.Resource     { return m.base.Font(s) }
func (m *MediumTheme) Size(n fyne.ThemeSizeName) float32       { return m.base.Size(n) }

