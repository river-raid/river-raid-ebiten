package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
)

// highBit is a bitmask for the most significant bit of a byte.
const highBit = 0x80

// DrawText draws text spans onto the screen. Each character is rendered as
// an 8×8 glyph using the ZX Spectrum ROM font or UDG bitmaps.
func DrawText(screen draw.Image, spans []assets.TextSpan) {
	for _, span := range spans {
		col := span.Col
		ink := palette[span.Ink]

		for _, r := range span.Text {
			glyph := assets.GlyphData(r)
			px := col * assets.GlyphSize
			py := span.Row * assets.GlyphSize

			for row := range assets.GlyphSize {
				b := glyph[row]
				for bit := range assets.GlyphSize {
					if b&(highBit>>bit) != 0 {
						screen.Set(px+bit, py+row, ink)
					}
				}
			}

			col++
		}
	}
}
