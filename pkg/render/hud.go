package render

import (
	"fmt"
	"image"
	"image/draw"
	"strings"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// HUD layout constants (character-cell row/column coordinates).
const (
	// Row 22 — scores.
	hudRowScore = 22

	hudColP1Label    = 2
	hudColP1Score    = 5
	hudColHILabel    = 18
	hudColHIScore    = 21
	hudScoreDigits   = 7
	hudHIScoreDigits = 7

	// Row 18 — game info and bridge.
	hudRowGameInfo = 18

	hudColGameLabel    = 2
	hudColBridgeLabel  = 18
	hudColBridgeCount  = 26
	hudBridgeCountCols = 2

	// Row 19 — fuel gauge scale and lives.
	hudRowFuelScale = 19

	hudColGaugeScaleStart = 8
	hudColGaugeScaleEnd   = 16 // inclusive; gauge is cols 8–16 = 9 characters wide
	hudColLivesStart      = 18
	hudGaugeWidth         = hudColGaugeScaleEnd - hudColGaugeScaleStart + 1 // 9

	// Row 20 — fuel gauge fill.
	hudRowFuelFill = 20

	// Fuel level divisor for gauge fill calculation.
	hudFuelMax = 255
)

// hudGameInfoText is the static "GAME  E   ½   F" string on row 18.
// Columns 2–17 = 16 characters.
var hudGameInfoText = "GAME  E   " + string(assets.GlyphHalf) + "   F" //nolint:gochecknoglobals // constant string

// hudGaugeEdgeCount is the number of edge marks (one left, one right) on the gauge scale.
const hudGaugeEdgeCount = 2

// hudGaugeScaleText is the 9-character fuel gauge scale on row 19 (cols 8–16).
// The left edge uses GlyphGaugeScaleLeft (tall left stripe + bottom border),
// interior cells use GlyphGaugeScaleTick (short mark + bottom border), and
// the right edge uses GlyphGaugeScaleRight (tall left stripe, no bottom border).
var hudGaugeScaleText = string(assets.GlyphGaugeScaleLeft) + //nolint:gochecknoglobals // constant string
	strings.Repeat(string(assets.GlyphGaugeScaleTick), hudGaugeWidth-hudGaugeEdgeCount) +
	string(assets.GlyphGaugeScaleRight)

// DrawHUD renders the full status bar onto screen.
func DrawHUD(screen draw.Image, s *state.GameState) {
	playerColor := playerColors[s.CurrentPlayer]

	p1Score := fmt.Sprintf("%0*d", hudScoreDigits, s.Players[domain.Player1].Score)
	DrawText(screen, []assets.TextSpan{
		{Row: hudRowScore, Col: hudColP1Label, Ink: platform.ColorYellow, Text: "P1"},
		{Row: hudRowScore, Col: hudColP1Score, Ink: platform.ColorYellow, Text: p1Score},
	})

	if s.Config.IsTwoPlayer {
		p2Score := fmt.Sprintf("%0*d", hudHIScoreDigits, s.Players[domain.Player2].Score)
		DrawText(screen, []assets.TextSpan{
			{Row: hudRowScore, Col: hudColHILabel, Ink: platform.ColorCyan, Text: "P2"},
			{Row: hudRowScore, Col: hudColHIScore, Ink: platform.ColorCyan, Text: p2Score},
		})
	} else {
		hiScore := fmt.Sprintf("%0*d", hudHIScoreDigits, s.HighScores[domain.HighScoreSlot(s.Config.StartingBridge)])
		DrawText(screen, []assets.TextSpan{
			{Row: hudRowScore, Col: hudColHILabel, Ink: platform.ColorWhite, Text: "HI"},
			{Row: hudRowScore, Col: hudColHIScore, Ink: platform.ColorWhite, Text: hiScore},
		})
	}

	bridgeCount := fmt.Sprintf("%*d", hudBridgeCountCols, s.Players[s.CurrentPlayer].BridgeCounter)
	DrawText(screen, []assets.TextSpan{
		{Row: hudRowGameInfo, Col: hudColGameLabel, Ink: platform.ColorWhite, Text: hudGameInfoText},
		{Row: hudRowGameInfo, Col: hudColBridgeLabel, Ink: playerColor, Text: "BRIDGE"},
		{Row: hudRowGameInfo, Col: hudColBridgeCount, Ink: playerColor, Text: bridgeCount},
	})

	livesText := buildLivesText(s.Players[s.CurrentPlayer].Lives)
	DrawText(screen, []assets.TextSpan{
		{Row: hudRowFuelScale, Col: hudColGaugeScaleStart, Ink: platform.ColorWhite, Text: hudGaugeScaleText},
		{Row: hudRowFuelScale, Col: hudColLivesStart, Ink: playerColor, Text: livesText},
	})

	drawFuelBar(screen, s.Fuel)
}

// hudGaugePixelWidth is the pixel width of the fillable fuel gauge area.
// The scale spans hudGaugeWidth tiles but the right-edge glyph (GlyphGaugeScaleRight)
// marks the boundary with only its leftmost pixel. The fill therefore runs from
// x=colStart to x=colStart+hudGaugePixelWidth-1, i.e. (hudGaugeWidth-1) full
// tiles plus that single boundary pixel.
const hudGaugePixelWidth = (hudGaugeWidth-1)*assets.GlyphSize + 1

// drawFuelBar fills the fuel gauge row with a solid magenta rectangle whose
// width is proportional to the current fuel level at 1-pixel precision.
// The bar occupies the full 8-pixel character-cell height at row hudRowFuelFill,
// starting at column hudColGaugeScaleStart.
func drawFuelBar(screen draw.Image, fuel int) {
	fillPx := (fuel * hudGaugePixelWidth) / hudFuelMax
	if fillPx > hudGaugePixelWidth {
		fillPx = hudGaugePixelWidth
	}

	if fillPx == 0 {
		return
	}

	x0 := hudColGaugeScaleStart * assets.GlyphSize
	y0 := hudRowFuelFill * assets.GlyphSize
	r := image.Rect(x0, y0, x0+fillPx, y0+assets.GlyphSize)
	ink := palette[platform.ColorMagenta]
	draw.Draw(screen, r, &image.Uniform{C: ink}, image.Point{}, draw.Src)
}

// buildLivesText returns a string of GlyphPlane runes (one per remaining life),
// padded with spaces so that previously displayed lives are erased.
// The maximum lives display width is domain.LivesInitial characters.
func buildLivesText(lives int) string {
	if lives > domain.LivesInitial {
		lives = domain.LivesInitial
	}

	planes := strings.Repeat(string(assets.GlyphPlane), lives)
	spaces := strings.Repeat(" ", domain.LivesInitial-lives)

	return planes + spaces
}
