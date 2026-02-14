package main

import (
	"image/color"
)

const (
	// paletteOn is the ZX Spectrum non-bright channel value.
	paletteOn = 197
	// paletteAlpha is the fully opaque alpha value.
	paletteAlpha = 255
)

// Palette maps the Color enum to ZX Spectrum non-bright RGB values.
// Each channel is either 0 or paletteOn (215).
var Palette = [8]color.RGBA{
	{A: paletteAlpha},                                           // Black
	{B: paletteOn, A: paletteAlpha},                             // Blue
	{R: paletteOn, A: paletteAlpha},                             // Red
	{R: paletteOn, B: paletteOn, A: paletteAlpha},               // Magenta
	{G: paletteOn, A: paletteAlpha},                             // Green
	{G: paletteOn, B: paletteOn, A: paletteAlpha},               // Cyan
	{R: paletteOn, G: paletteOn, A: paletteAlpha},               // Yellow
	{R: paletteOn, G: paletteOn, B: paletteOn, A: paletteAlpha}, // White
}
