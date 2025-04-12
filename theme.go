package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MyTheme struct {
	base fyne.Theme
}

func (t *MyTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{43, 45, 66, 255}
	case theme.ColorNameForeground:
		return color.RGBA{217, 215, 201, 255}
	case theme.ColorNamePrimary:
		return color.RGBA{230, 190, 125, 255}
	case theme.ColorNameButton:
		return color.RGBA{74, 55, 35, 255}
	case theme.ColorNameSelection:
		return color.RGBA{217, 168, 108, 255}
	case theme.ColorNameHover:
		return color.RGBA{184, 151, 120, 255}
	case theme.ColorNameSeparator:
		return color.RGBA{230, 190, 125, 255}
	default:
		return t.base.Color(name, variant)
	}
}

func (t *MyTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *MyTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *MyTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameCaptionText:
		return 14
	case theme.SizeNamePadding:
		return 6
	default:
		return t.base.Size(name)
	}
}
