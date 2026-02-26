package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

// drawSprite draws a sprite at pixel position (x, y) onto screen.
// Set bits are drawn in ink color; unset bits are left unchanged (transparent).
// If mirror is true, the sprite is flipped horizontally.
func drawSprite(screen draw.Image, s assets.Sprite, x, y int, color platform.Color, mirror bool) {
	ink := palette[color]
	for row := range s.Height() {
		for col := range s.Width {
			byteIdx := row*s.BytesPerRow + col/platform.BitsPerByte
			bitIdx := (platform.BitsPerByte - 1) - col%platform.BitsPerByte

			if s.Data[byteIdx]&(1<<bitIdx) != 0 {
				px := col
				if mirror {
					px = s.Width - 1 - col
				}

				screen.Set(x+px, y+row, ink)
			}
		}
	}
}
