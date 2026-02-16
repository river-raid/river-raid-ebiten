package render

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/hajimehoshi/ebiten/v2"
)

// CircularImage wraps a draw.Image and provides automatic Y-coordinate wrapping
// for circular buffer behavior. This allows rendering code to work with coordinates
// as if the image were infinitely tall, while internally wrapping to the buffer height.
type CircularImage struct {
	img    draw.Image
	height int
}

// NewCircularImage creates a new circular image wrapper with the given dimensions.
func NewCircularImage(width, height int) *CircularImage {
	return &CircularImage{
		img:    ebiten.NewImage(width, height),
		height: height,
	}
}

// Set sets the color of the pixel at (x, y), automatically wrapping y to buffer bounds.
func (ci *CircularImage) Set(x, y int, c color.Color) {
	y = ((y % ci.height) + ci.height) % ci.height
	ci.img.Set(x, y, c)
}

// Image returns the underlying image for operations that need direct access
// (e.g., DrawTerrainBuffer, which handles wrapping separately for viewport rendering).
func (ci *CircularImage) Image() image.Image {
	return ci.img
}
