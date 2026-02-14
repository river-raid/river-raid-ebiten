package main

import "github.com/hajimehoshi/ebiten/v2"

// demoSpans is temporary text used to verify text rendering (will be removed).
//
//nolint:mnd // temporary demo data with hardcoded positions
var demoSpans = []TextSpan{
	{Row: 0, Col: 0, Ink: ColorWhite, Text: "RIVER RAID"},
	{Row: 1, Col: 0, Ink: ColorCyan, Text: "by "},
	{Row: 1, Col: 3, Ink: ColorGreen, Text: string([]rune{
		GlyphLogo0, GlyphLogo1, GlyphLogo2,
		GlyphLogo3, GlyphLogo4, GlyphLogo5, GlyphLogo6,
	})},
	{Row: 1, Col: 10, Ink: ColorWhite, Text: string([]rune{GlyphTrademark})},
}

// Game implements the ebiten.Game interface.
type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	renderText(screen, demoSpans)
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}
