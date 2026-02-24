package render

import (
	"image/color"
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// viewportCanvas wraps a draw.Image and clips all Set calls to the viewport Y range [0, ViewportHeight).
// This ensures sprites and projectiles cannot bleed into the status bar area below the viewport.
type viewportCanvas struct {
	draw.Image
}

func newViewportCanvas(img draw.Image) *viewportCanvas {
	return &viewportCanvas{img}
}

// Set forwards the pixel write only if y is within the viewport bounds.
func (vc *viewportCanvas) Set(x, y int, c color.Color) {
	if y >= 0 && y < domain.ViewportHeight {
		vc.Image.Set(x, y, c)
	}
}
