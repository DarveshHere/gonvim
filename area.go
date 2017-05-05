package gonvim

import (
	"fmt"

	"github.com/dzhou121/ui"
)

// Border is the border of area
type Border struct {
	color *RGBA
	width int
}

// AreaHandler is
type AreaHandler struct {
	area         *ui.Area
	width        int
	height       int
	span         *ui.Area
	borderTop    *Border
	borderRight  *Border
	borderLeft   *Border
	borderBottom *Border
	x, y         int
	shown        bool
}

func drawRect(dp *ui.AreaDrawParams, x, y, width, height int, color *RGBA) {
	p := ui.NewPath(ui.Winding)
	p.AddRectangle(float64(x), float64(y), float64(width), float64(height))
	p.End()
	dp.Context.Fill(p, &ui.Brush{
		Type: ui.Solid,
		R:    color.R,
		G:    color.G,
		B:    color.B,
		A:    color.A,
	})
	p.Free()
}

// MouseEvent is
func (ah *AreaHandler) MouseEvent(a *ui.Area, me *ui.AreaMouseEvent) {
}

// MouseCrossed is
func (ah *AreaHandler) MouseCrossed(a *ui.Area, left bool) {
}

// DragBroken is
func (ah *AreaHandler) DragBroken(a *ui.Area) {
}

// KeyEvent is
func (ah *AreaHandler) KeyEvent(a *ui.Area, key *ui.AreaKeyEvent) (handled bool) {
	if key.Up {
		return false
	}
	if key.Key == 0 && key.ExtKey == 0 {
		return false
	}
	namedKey := ""
	mod := ""

	switch key.Modifiers {
	case ui.Ctrl:
		mod = "C-"
	case ui.Alt:
		mod = "A-"
	case ui.Super:
		mod = "M-"
	}

	switch key.ExtKey {
	case ui.Escape:
		namedKey = "Esc"
	case ui.Insert:
		namedKey = "Insert"
	case ui.Delete:
		namedKey = "Del"
	case ui.Left:
		namedKey = "Left"
	case ui.Right:
		namedKey = "Right"
	case ui.Down:
		namedKey = "Down"
	case ui.Up:
		namedKey = "Up"
	}

	char := ""
	char = string(key.Key)
	if char == "\n" || char == "\r" {
		namedKey = "Enter"
	} else if char == "\t" {
		namedKey = "Tab"
	} else if key.Key == 127 {
		namedKey = "BS"
	} else if char == "<" {
		namedKey = "LT"
	}

	input := ""
	if namedKey != "" {
		input = fmt.Sprintf("<%s>", namedKey)
	} else if mod != "" {
		input = fmt.Sprintf("<%s%s>", mod, char)
	} else {
		input = char
	}
	editor.nvim.Input(input)
	return true
}

func (ah *AreaHandler) setPosition(x, y int) {
	if x == ah.x && y == ah.y {
		return
	}
	ah.x = x
	ah.y = y
	ui.QueueMain(func() {
		ah.area.SetPosition(x, y)
	})
}

func (ah *AreaHandler) show() {
	if ah.shown {
		return
	}
	ah.shown = true
	ui.QueueMain(func() {
		ah.area.Show()
	})
}

func (ah *AreaHandler) hide() {
	if !ah.shown {
		return
	}
	ah.shown = false
	ui.QueueMain(func() {
		ah.area.Hide()
	})
}

func (ah *AreaHandler) setSize(width, height int) {
	if width == ah.width && height == ah.height {
		return
	}
	ah.width = width
	ah.height = height
	ui.QueueMain(func() {
		ah.area.SetSize(width, height)
	})
}
