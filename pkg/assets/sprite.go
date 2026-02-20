package assets

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
	bpr := (w + 7) / 8 //nolint:mnd // ceiling division by 8 bits per byte

	return Sprite{
		Data:        data,
		Width:       w,
		BytesPerRow: bpr,
	}
}
