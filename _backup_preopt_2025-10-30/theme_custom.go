// theme_custom.go
package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// MediumTheme is a slightly darker Light with higher contrast controls.
type MediumTheme struct {
	base fyne.Theme
}

func NewMediumTheme() fyne.Theme {
	return &MediumTheme{base: theme.LightTheme()}
}

func (m *MediumTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		// warm grey background (easier on eyes than bright white)
		return color.NRGBA{242, 243, 245, 255} // #F2F3F5
	case theme.ColorNameInputBackground:
		return color.NRGBA{230, 232, 236, 255}
	case theme.ColorNameButton:
		return color.NRGBA{98, 91, 246, 255} // purple-ish accent
	case theme.ColorNameFocus:
		return color.NRGBA{0, 114, 206, 255} // same as your selection blue
	case theme.ColorNameHover:
		return color.NRGBA{230, 235, 245, 255}
	}
	return m.base.Color(n, v)
}

func (m *MediumTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return m.base.Icon(n) }
func (m *MediumTheme) Font(s fyne.TextStyle) fyne.Resource     { return m.base.Font(s) }
func (m *MediumTheme) Size(n fyne.ThemeSizeName) float32       { return m.base.Size(n) }
