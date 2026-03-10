package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

// Instructions screen layout constants.
const (
	instrColDirection = 8  // left margin for directional key-binding text
	instrColFireCont  = 19 // col for "row" continuation (aligns with "Bottom")
	instrColPause     = 7
	instrColEnter     = 6
	instrColCaps1     = 4
	instrColCaps2     = 3  // "to reset the game you have"
	instrColCaps3     = 10 // "just played"
	instrColSym1      = 0  // "Press SYM SHIFT & ENTER to reset" (full width)
	instrColSym2      = 6  // "the menu selections"

	instrRowLeft     = 1
	instrRowRight    = 3
	instrRowFaster   = 5
	instrRowSlower   = 7
	instrRowFire     = 9
	instrRowFireCont = 10
	instrRowPause    = 12
	instrRowEnter    = 14
	instrRowCaps1    = 16
	instrRowCaps2    = 17
	instrRowCaps3    = 18
	instrRowSym1     = 20
	instrRowSym2     = 21
)

// DrawInstructions draws the static instructions screen.
func DrawInstructions(screen draw.Image) {
	DrawText(screen, []assets.TextSpan{
		{Row: instrRowLeft, Col: instrColDirection, Ink: platform.ColorRed, Text: "LEFT........O"},
		{Row: instrRowRight, Col: instrColDirection, Ink: platform.ColorMagenta, Text: "RIGHT.......P"},
		{Row: instrRowFaster, Col: instrColDirection, Ink: platform.ColorYellow, Text: "FASTER......Q"},
		{Row: instrRowSlower, Col: instrColDirection, Ink: platform.ColorGreen, Text: "SLOWER......A"},
		{Row: instrRowFire, Col: instrColDirection, Ink: platform.ColorCyan, Text: "FIRE......Bottom"},
		{Row: instrRowFireCont, Col: instrColFireCont, Ink: platform.ColorCyan, Text: "row"},
		{Row: instrRowPause, Col: instrColPause, Ink: platform.ColorWhite, Text: "Press H to pause"},
		{Row: instrRowEnter, Col: instrColEnter, Ink: platform.ColorWhite, Text: "Press ENTER to play"},
		{Row: instrRowCaps1, Col: instrColCaps1, Ink: platform.ColorWhite, Text: "Press CAPS SHIFT & ENTER"},
		{Row: instrRowCaps2, Col: instrColCaps2, Ink: platform.ColorWhite, Text: "to reset the game you have"},
		{Row: instrRowCaps3, Col: instrColCaps3, Ink: platform.ColorWhite, Text: "just played"},
		{Row: instrRowSym1, Col: instrColSym1, Ink: platform.ColorWhite, Text: "Press SYM SHIFT & ENTER to reset"},
		{Row: instrRowSym2, Col: instrColSym2, Ink: platform.ColorWhite, Text: "the menu selections"},
	})
}
