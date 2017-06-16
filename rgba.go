package gonvim

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

func newRGBA(r int, g int, b int, a float64) *RGBA {
	return &RGBA{
		R: r,
		G: g,
		B: b,
		A: a,
	}
}
