package main

import (
	"image/draw"
)

// highBit is a bitmask for the most significant bit of a byte.
const highBit = 0x80

// TextSpan defines a run of text at a given character-cell position with
// ink (foreground) and paper (background) colors.
type TextSpan struct {
	Text     string
	Row, Col int
	Ink      Color
}

// renderText draws text spans onto the screen. Each character is rendered as
// an 8×8 glyph using the ZX Spectrum ROM font or UDG bitmaps.
func renderText(screen draw.Image, spans []TextSpan) {
	for _, span := range spans {
		col := span.Col
		ink := Palette[span.Ink]

		for _, r := range span.Text {
			glyph := glyphData(r)
			px := col * glyphSize
			py := span.Row * glyphSize

			for row := range glyphSize {
				b := glyph[row]
				for bit := range glyphSize {
					if b&(highBit>>bit) != 0 {
						screen.Set(px+bit, py+row, ink)
					}
				}
			}

			col++
		}
	}
}
