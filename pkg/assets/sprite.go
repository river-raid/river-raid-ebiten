package assets

import (
	"fmt"

	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

// Sprite holds a 1-bit-per-pixel bitmap and its pixel dimensions.
// The bitmap is stored as ceil(Width/8) bytes per row, MSB first.
type Sprite struct {
	Data        []byte
	Width       int
	Height      int
	BytesPerRow int
}

// newSprite creates a Sprite from raw 1bpp bitmap Data.
// w is the visual width in pixels; h is the visual height in pixels.
// Panics if len(data) != h * bytesPerRow.
func newSprite(data []byte, w, h int) Sprite {
	bpr := (w + platform.BitsPerByte - 1) / platform.BitsPerByte

	if len(data) != h*bpr {
		panic(fmt.Sprintf("newSprite: data length %d != h(%d) * bytesPerRow(%d)", len(data), h, bpr))
	}

	return Sprite{
		Data:        data,
		Width:       w,
		Height:      h,
		BytesPerRow: bpr,
	}
}
