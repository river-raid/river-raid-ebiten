package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

// ColorFn returns the ink color for a sprite pixel at viewport screen coordinate (x, y).
type ColorFn func(x, y int) platform.Color

// staticColorFn returns a ColorFn that always returns the given color.
func staticColorFn(c platform.Color) ColorFn {
	return func(_, _ int) platform.Color { return c }
}

// drawSprite draws a sprite at pixel position (x, y) onto screen.
// Set bits are drawn using colorFn; unset bits are left unchanged (transparent).
// If mirror is true, the sprite is flipped horizontally.
func drawSprite(screen draw.Image, s assets.Sprite, x, y int, colorFn ColorFn, mirror bool) {
	for row := range s.Height {
		for col := range s.Width {
			byteIdx := row*s.BytesPerRow + col/platform.BitsPerByte
			bitIdx := (platform.BitsPerByte - 1) - col%platform.BitsPerByte

			if s.Data[byteIdx]&(1<<bitIdx) != 0 {
				px := col
				if mirror {
					px = s.Width - 1 - col
				}

				screenX := x + px
				screenY := y + row
				screen.Set(screenX, screenY, palette[colorFn(screenX, screenY)])
			}
		}
	}
}
