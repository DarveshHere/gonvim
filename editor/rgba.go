package editor

import (
	"fmt"

	"github.com/therecipe/qt/gui"
)

// RGBA is
type RGBA struct {
	R int
	G int
	B int
	A float64
}

func (rgba *RGBA) print() string {
	return fmt.Sprintf("rgba(%d, %d, %d, 1)", rgba.R, rgba.G, rgba.B)
}

func (rgba *RGBA) copy() *RGBA {
	return &RGBA{
		R: rgba.R,
		G: rgba.G,
		B: rgba.B,
		A: rgba.A,
	}
}

func (rgba *RGBA) equals(other *RGBA) bool {
	return rgba.R == other.R && rgba.G == other.G && rgba.B == other.B && rgba.A == other.A
}

func (rgba *RGBA) String() string {
	return fmt.Sprintf("rgba(%d, %d, %d, %f)", rgba.R, rgba.G, rgba.B, rgba.A)
}

// Hex is
func (rgba *RGBA) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", uint8(rgba.R), uint8(rgba.G), uint8(rgba.B))
}

// QColor is
func (rgba *RGBA) QColor() *gui.QColor {
	return gui.NewQColor3(rgba.R, rgba.G, rgba.B, int(rgba.A*255))
}

func calcColor(c int) *RGBA {
	b := c & 255
	g := (c >> 8) & 255
	r := (c >> 16) & 255
	return &RGBA{
		R: r,
		G: g,
		B: b,
		A: 1,
	}
}

func reverseColor(rgba *RGBA) *RGBA {
	if rgba == nil {
		return &RGBA{0, 0, 0, 1}
	}
	var r, g, b int

	r = 255 - rgba.R
	g = 255 - rgba.G
	b = 255 - rgba.B

	return &RGBA{
		R: r,
		G: g,
		B: b,
		A: rgba.A,
	}
}

func gradColor(rgba *RGBA) *RGBA {
	if rgba == nil {
		return &RGBA{0, 0, 0, 1}
	}
	var r, g, b int

	if rgba.R > 128 {
		r = rgba.R - (rgba.R-128)/2
	} else {
		r = rgba.R + (128-rgba.R)/2
	}

	if rgba.G > 128 {
		g = rgba.G - (rgba.G-128)/2
	} else {
		g = rgba.G + (128-rgba.G)/2
	}

	if rgba.B > 128 {
		b = rgba.B - (rgba.B-128)/2
	} else {
		b = rgba.B + (128-rgba.B)/2
	}

	return &RGBA{
		R: r,
		G: g,
		B: b,
		A: rgba.A,
	}
}

func warpColor(rgba *RGBA, v int) *RGBA {
	if rgba == nil {
		return &RGBA{0, 0, 0, 1}
	}
	var r, g, b int

	if rgba.R > 128 {
		r = rgba.R + v
	} else {
		r = rgba.R - (2 * v)
	}

	if rgba.G > 128 {
		g = rgba.G + v
	} else {
		g = rgba.G - (2 * v)
	}

	if rgba.B > 128 {
		b = rgba.B + v
	} else {
		b = rgba.B - (2 * v)
	}

	if r <= 0 {
		r = 0
	}
	if g <= 0 {
		g = 0
	}
	if b <= 0 {
		b = 0
	}

	if r >= 255 {
		r = 255
	}
	if g >= 255 {
		g = 255
	}
	if b >= 255 {
		b = 255
	}

	return &RGBA{
		R: r,
		G: g,
		B: b,
		A: rgba.A,
	}
}

func shiftColor(rgba *RGBA, v int) *RGBA {
	if rgba == nil {
		return &RGBA{0, 0, 0, 1}
	}
	var r, g, b int
	r = rgba.R - v
	g = rgba.G - v
	b = rgba.B - v

	if r <= 0 {
		r = 0
	}
	if g <= 0 {
		g = 0
	}
	if b <= 0 {
		b = 0
	}

	if r >= 255 {
		r = 255
	}
	if g >= 255 {
		g = 255
	}
	if b >= 255 {
		b = 255
	}

	return &RGBA{
		R: r,
		G: g,
		B: b,
		A: rgba.A,
	}
}

func newRGBA(r int, g int, b int, a float64) *RGBA {
	return &RGBA{
		R: r,
		G: g,
		B: b,
		A: a,
	}
}
