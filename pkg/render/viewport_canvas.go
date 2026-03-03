package render

import (
	"image/color"
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// viewportCanvas wraps a draw.Image, clips Set calls to game Y ∈ [ViewportBlankZone, TotalViewportHeight),
// and maps game Y to screen Y by subtracting ViewportBlankZone.
// This hides the blank buffer zone (rows 0–7) and prevents bleed into the status bar.
type viewportCanvas struct {
	draw.Image
}

func newViewportCanvas(img draw.Image) *viewportCanvas {
	return &viewportCanvas{img}
}

// Set writes the pixel only when game Y is in the visible range, mapping to screen Y = game Y − ViewportBlankZone.
func (vc *viewportCanvas) Set(x, y int, c color.Color) {
	if y >= domain.ViewportBlankZone && y < domain.TotalViewportHeight {
		vc.Image.Set(x, y-domain.ViewportBlankZone, c)
	}
}
