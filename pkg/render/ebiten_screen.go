package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// EbitenScreen wraps an ebiten.Image to implement the Screen interface.
type EbitenScreen struct {
	img *ebiten.Image
}

// NewEbitenScreen creates a Screen wrapper around an ebiten.Image.
func NewEbitenScreen(img *ebiten.Image) Screen {
	return &EbitenScreen{img: img}
}

// ColorModel implements image.Image.
func (s *EbitenScreen) ColorModel() color.Model {
	return s.img.ColorModel()
}

// Bounds implements image.Image.
func (s *EbitenScreen) Bounds() image.Rectangle {
	return s.img.Bounds()
}

// At implements image.Image.
func (s *EbitenScreen) At(x, y int) color.Color {
	return s.img.At(x, y)
}

// Set implements draw.Image.
func (s *EbitenScreen) Set(x, y int, c color.Color) {
	s.img.Set(x, y, c)
}

// DrawImageRegion implements Screen.
func (s *EbitenScreen) DrawImageRegion(src image.Image, srcRect image.Rectangle, dstX, dstY int) {
	// Convert src to ebiten.Image if needed.
	var ebitenSrc *ebiten.Image
	if img, ok := src.(*ebiten.Image); ok {
		ebitenSrc = img
	} else {
		// If src is not an ebiten.Image, we need to convert it.
		// This should rarely happen in practice.
		ebitenSrc = ebiten.NewImageFromImage(src)
	}

	// Extract the sub-image for the source region.
	subImg := ebitenSrc.SubImage(srcRect).(*ebiten.Image) //nolint:errcheck // SubImage on *ebiten.Image always succeeds

	// Draw the sub-image at the destination position.
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(dstX), float64(dstY))
	s.img.DrawImage(subImg, op)
}
