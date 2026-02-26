package assets

import "github.com/morozov/river-raid-ebiten/pkg/platform"

// Sprite holds a 1-bit-per-pixel bitmap and its pixel dimensions.
// The bitmap is stored as ceil(Width/8) bytes per row, MSB first.
// Height is derived from len(Data) / BytesPerRow.
type Sprite struct {
	Data        []byte
	Width       int
	BytesPerRow int
}

// Height returns the sprite height in pixels.
func (s Sprite) Height() int {
	return len(s.Data) / s.BytesPerRow
}

// newSprite creates a Sprite from raw 1bpp bitmap Data.
// w is the visual width in pixels; height is derived from len(Data).
func newSprite(data []byte, w int) Sprite {
	bpr := (w + platform.BitsPerByte - 1) / platform.BitsPerByte

	return Sprite{
		Data:        data,
		Width:       w,
		BytesPerRow: bpr,
	}
}
