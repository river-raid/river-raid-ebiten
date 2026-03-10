package render

import (
	"fmt"
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

// Row positions for the control type menu and game mode dialog.
const (
	ctrlRowPrompt1  = 3
	ctrlRowPrompt2  = 4
	ctrlRowKeyboard = 8
	ctrlRowSinclair = 10
	ctrlRowKempston = 12
	ctrlRowCursor   = 14
)

// Column positions for the control type menu.
const (
	ctrlColPrompt1 = 3
	ctrlColPrompt2 = 10
	ctrlColOptions = 6
)

// Row positions for the game mode dialog header.
const (
	modeRowHeader1 = 6
	modeRowHeader2 = 7
)

// Column positions for the game mode dialog header.
const (
	modeColGameHeader     = 7
	modeColPlayersHeader  = 15
	modeColBridgeHeader   = 22
	modeColGameHeader2    = 6
	modeColPlayersHeader2 = 14
	modeColBridgeHeader2  = 23
)

// Row positions for game mode rows (modes 1–8).
const (
	modeRowMode1 = 9
	modeRowMode2 = 10
	modeRowMode3 = 12
	modeRowMode4 = 13
	modeRowMode5 = 15
	modeRowMode6 = 16
	modeRowMode7 = 18
	modeRowMode8 = 19
)

// Column positions for game mode data columns.
const (
	modeColNumber  = 9
	modeColPlayers = 17
	modeColBridge  = 24
)

// DrawControlSelection draws the control selection screen.
// phase 0 = control type menu, phase 1 = game mode dialog.
func DrawControlSelection(screen draw.Image, phase int) {
	if phase == 0 {
		drawControlTypeMenu(screen)
	} else {
		drawGameModeDialog(screen)
	}
}

func drawControlTypeMenu(screen draw.Image) {
	DrawText(screen, []assets.TextSpan{
		{Text: "Press corresponding number", Row: ctrlRowPrompt1, Col: ctrlColPrompt1, Ink: platform.ColorWhite},
		{Text: "on keyboard", Row: ctrlRowPrompt2, Col: ctrlColPrompt2, Ink: platform.ColorWhite},
		{Text: "1. KEYBOARD CONTROL", Row: ctrlRowKeyboard, Col: ctrlColOptions, Ink: platform.ColorWhite},
		{Text: "2. SINCLAIR INTERFACE", Row: ctrlRowSinclair, Col: ctrlColOptions, Ink: platform.ColorWhite},
		{Text: "3. KEMPSTON INTERFACE", Row: ctrlRowKempston, Col: ctrlColOptions, Ink: platform.ColorWhite},
		{Text: "4. CURSOR INTERFACE", Row: ctrlRowCursor, Col: ctrlColOptions, Ink: platform.ColorWhite},
	})
}

func drawGameModeDialog(screen draw.Image) {
	type modeRow struct {
		row     int
		num     int
		players int
		bridge  int
	}

	modeRows := []modeRow{
		{modeRowMode1, 1, 1, 1},
		{modeRowMode2, 2, 2, 1},
		{modeRowMode3, 3, 1, 5},
		{modeRowMode4, 4, 2, 5},
		{modeRowMode5, 5, 1, 20},
		{modeRowMode6, 6, 2, 20},
		{modeRowMode7, 7, 1, 30},
		{modeRowMode8, 8, 2, 30},
	}

	spans := []assets.TextSpan{
		{Text: "Press corresponding number", Row: ctrlRowPrompt1, Col: ctrlColPrompt1, Ink: platform.ColorWhite},
		{Text: "on keyboard", Row: ctrlRowPrompt2, Col: ctrlColPrompt2, Ink: platform.ColorWhite},
		{Text: "Game", Row: modeRowHeader1, Col: modeColGameHeader, Ink: platform.ColorWhite},
		{Text: "No of", Row: modeRowHeader1, Col: modeColPlayersHeader, Ink: platform.ColorWhite},
		{Text: "Starting", Row: modeRowHeader1, Col: modeColBridgeHeader, Ink: platform.ColorWhite},
		{Text: "Number", Row: modeRowHeader2, Col: modeColGameHeader2, Ink: platform.ColorWhite},
		{Text: "Players", Row: modeRowHeader2, Col: modeColPlayersHeader2, Ink: platform.ColorWhite},
		{Text: "Bridge", Row: modeRowHeader2, Col: modeColBridgeHeader2, Ink: platform.ColorWhite},
	}

	for _, mr := range modeRows {
		spans = append(spans,
			assets.TextSpan{Text: fmt.Sprintf("%d", mr.num), Row: mr.row, Col: modeColNumber, Ink: platform.ColorWhite},
			assets.TextSpan{Text: fmt.Sprintf("%d", mr.players), Row: mr.row, Col: modeColPlayers, Ink: platform.ColorWhite},
			assets.TextSpan{Text: fmt.Sprintf("%2d", mr.bridge), Row: mr.row, Col: modeColBridge, Ink: platform.ColorWhite},
		)
	}

	DrawText(screen, spans)
}
