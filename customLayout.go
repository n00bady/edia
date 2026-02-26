package main

import (

	"fyne.io/fyne/v2"
)

// CenteredButtonsLayout is a custom layout for centered buttons
type CenteredButtonsLayout struct {
	buttonWidth  float32
	buttonHeight float32
	spacing      float32
}

// NewCenteredButtonsLayout creates a new centered buttons layout
func NewCenteredButtonsLayout(btnWidth, btnHeight, spacing float32) *CenteredButtonsLayout {
	return &CenteredButtonsLayout{
		buttonWidth:  btnWidth,
		buttonHeight: btnHeight,
		spacing:      spacing,
	}
}

// Layout positions the objects according to the custom layout
func (c *CenteredButtonsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	// Calculate total height needed for buttons and spacing
	totalHeight := (c.buttonHeight * 3) + (c.spacing * 2)
	
	// Calculate starting Y position to center vertically
	startY := (size.Height - totalHeight) / 2
	
	// Calculate starting X position to center horizontally
	startX := (size.Width - c.buttonWidth) / 2

	// Position each button
	for i, obj := range objects {
		y := startY + (float32(i) * (c.buttonHeight + c.spacing))
		obj.Resize(fyne.NewSize(c.buttonWidth, c.buttonHeight))
		obj.Move(fyne.NewPos(startX, y))
	}
}

// MinSize returns the minimum size required by the layout
func (c *CenteredButtonsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(300, 300)
}
