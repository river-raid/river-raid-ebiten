package main

import (
	"image"
	"image/color"
	"testing"
)

func TestRenderText_Position(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 24))

	renderText(img, []TextSpan{
		{Text: "A", Row: 1, Col: 2, Ink: ColorWhite},
	})

	wantInk := Palette[ColorWhite]

	// "A" at row=1, col=2 starts at pixel (16, 8).
	// Row 1 of "A" is 0x3c — bit 2 set → pixel (18, 9) is ink.
	assertColor(t, img, 18, 9, wantInk, "ink")

	// Origin (0,0) should be untouched (zero value).
	assertColor(t, img, 0, 0, color.RGBA{}, "untouched")
}

func TestRenderText_MultiCharAdvance(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 24, 8))

	renderText(img, []TextSpan{
		{Text: "AB", Row: 0, Col: 0, Ink: ColorRed},
	})

	// "B" starts at pixel x=8. Row 1 of "B" is 0x42 (01000010) — bit 1 set at x=9.
	assertColor(t, img, 9, 1, Palette[ColorRed], "ink")
}

func TestRenderText_MultipleSpans(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 24, 16))

	renderText(img, []TextSpan{
		{Text: "A", Row: 0, Col: 0, Ink: ColorRed},
		{Text: "B", Row: 1, Col: 1, Ink: ColorGreen},
	})

	// Span 1: "A" at (0,0). Row 1 is 0x3c — bit 2 set at pixel (2, 1).
	assertColor(t, img, 2, 1, Palette[ColorRed], "span1 ink")

	// Span 2: "B" at row=1, col=1 → pixel origin (8, 8).
	// Row 1 of "B" is 0x42 (01000010) — bit 1 set at pixel (9, 9).
	assertColor(t, img, 9, 9, Palette[ColorGreen], "span2 ink")
}

func TestRenderText_UDGGlyph(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))

	renderText(img, []TextSpan{
		{Text: string(GlyphPlane), Row: 0, Col: 0, Ink: ColorYellow},
	})

	wantInk := Palette[ColorYellow]

	// Row 0 of Plane is 0x10 (00010000) — only bit 3 is set.
	for x := range 8 {
		if x == 3 {
			assertColor(t, img, x, 0, wantInk, "ink")
		}
	}
}

func assertColor(t *testing.T, img *image.RGBA, x, y int, want color.RGBA, label string) {
	t.Helper()

	got := img.RGBAAt(x, y)
	if got != want {
		t.Errorf("pixel (%d,%d) %s: got %v, want %v", x, y, label, got, want)
	}
}
