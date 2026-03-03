package render

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/hajimehoshi/ebiten/v2"
)

// PixelBuffer is an interface for setting individual pixels.
// Implementations may handle coordinate wrapping or clipping as needed.
type PixelBuffer interface {
	Set(x, y int, c color.Color)
}

// CircularImage wraps an ebiten.Image and provides automatic Y-coordinate wrapping
// for circular buffer behavior. This allows rendering code to work with coordinates
// as if the image were infinitely tall, while internally wrapping to the buffer height.
type CircularImage struct {
	img    *ebiten.Image
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
// (e.g., drawTerrainBuffer, which handles wrapping separately for viewport rendering).
func (ci *CircularImage) Image() image.Image {
	return ci.img
}

// Clear resets the entire image to its initial zero state (transparent black),
// matching what ebiten.NewImage produces.
func (ci *CircularImage) Clear() {
	ci.img.Clear()
}

// StaticImageBuffer wraps a draw.Image for static (non-circular) rendering.
// It performs no coordinate wrapping, allowing direct pixel access.
type StaticImageBuffer struct {
	img draw.Image
}

// NewStaticImageBuffer creates a new static image buffer wrapper.
func NewStaticImageBuffer(img draw.Image) *StaticImageBuffer {
	return &StaticImageBuffer{img: img}
}

// Set sets the color of the pixel at (x, y) without any coordinate transformation.
func (sib *StaticImageBuffer) Set(x, y int, c color.Color) {
	sib.img.Set(x, y, c)
}
