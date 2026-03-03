package render

import (
	"fmt"
	"image/draw"
	"strings"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// HUD layout constants (character-cell row/column coordinates).
const (
	// Row 1 — scores.
	hudRowScore = 1

	hudColP1Label    = 2
	hudColP1Score    = 5
	hudColHILabel    = 18
	hudColHIScore    = 21
	hudScoreDigits   = 6
	hudHILabelWidth  = 2
	hudHIScoreDigits = 6

	// Row 19 — game info and bridge.
	hudRowGameInfo = 19

	hudColGameLabel    = 2
	hudColBridgeLabel  = 18
	hudColBridgeCount  = 26
	hudBridgeCountCols = 2

	// Row 20 — fuel gauge scale, FUEL label, and lives.
	hudRowFuelScale = 20

	hudColFuelLabel       = 4
	hudColGaugeScaleStart = 8
	hudColGaugeScaleEnd   = 16 // inclusive; gauge is cols 8–16 = 9 characters wide
	hudColLivesStart      = 18
	hudGaugeWidth         = hudColGaugeScaleEnd - hudColGaugeScaleStart + 1 // 9

	// Row 21 — fuel gauge fill.
	hudRowFuelFill = 21

	// Fuel level divisor for gauge fill calculation.
	hudFuelMax = 255
)

// hudGameInfoText is the static "GAME  E   ½   F" string on row 19.
// Columns 2–17 = 16 characters.
var hudGameInfoText = "GAME  E   " + string(assets.GlyphHalf) + "   F" //nolint:gochecknoglobals // constant string

// hudGaugeScaleText is the 9-character fuel gauge scale on row 20 (cols 8–16).
// Uses GlyphGaugeEmpty for the scale bar markers.
var hudGaugeScaleText = strings.Repeat(string(assets.GlyphGaugeEmpty), hudGaugeWidth) //nolint:gochecknoglobals // constant string

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
		hiScore := fmt.Sprintf("%0*d", hudHIScoreDigits, s.HighScores[startingBridgeSlot(s.Config.StartingBridge)])
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
		{Row: hudRowFuelScale, Col: hudColFuelLabel, Ink: platform.ColorWhite, Text: "FUEL"},
		{Row: hudRowFuelScale, Col: hudColGaugeScaleStart, Ink: platform.ColorWhite, Text: hudGaugeScaleText},
		{Row: hudRowFuelScale, Col: hudColLivesStart, Ink: playerColor, Text: livesText},
	})

	DrawText(screen, []assets.TextSpan{
		{Row: hudRowFuelFill, Col: hudColGaugeScaleStart, Ink: platform.ColorMagenta, Text: buildFuelGauge(s.Fuel)},
	})
}

// buildFuelGauge returns a string of GlyphGaugeFull runes whose length is
// proportional to the current fuel level (0–255) over hudGaugeWidth columns.
func buildFuelGauge(fuel int) string {
	fillCols := (fuel * hudGaugeWidth) / hudFuelMax
	if fillCols > hudGaugeWidth {
		fillCols = hudGaugeWidth
	}

	return strings.Repeat(string(assets.GlyphGaugeFull), fillCols)
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

// High score slot indices (0-based) corresponding to each StartingBridge option.
const (
	highScoreSlotBridge01 = 0
	highScoreSlotBridge05 = 1
	highScoreSlotBridge20 = 2
	highScoreSlotBridge30 = 3
)

// startingBridgeSlotTable maps the StartingBridge enum to a 0-based HighScores slot index.
var startingBridgeSlotTable = map[domain.StartingBridge]int{ //nolint:gochecknoglobals // constant lookup table
	domain.StartingBridge01: highScoreSlotBridge01,
	domain.StartingBridge05: highScoreSlotBridge05,
	domain.StartingBridge20: highScoreSlotBridge20,
	domain.StartingBridge30: highScoreSlotBridge30,
}

// startingBridgeSlot returns the 0-based HighScores slot index for a StartingBridge value.
func startingBridgeSlot(sb domain.StartingBridge) int {
	slot, ok := startingBridgeSlotTable[sb]
	if !ok {
		panic(fmt.Sprintf("startingBridgeSlot: unknown StartingBridge value %d", sb))
	}

	return slot
}
