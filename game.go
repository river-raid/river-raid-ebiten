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

// demoEntry defines a sprite to render in the demo grid.
type demoEntry struct {
	label  string
	id     SpriteID
	ink    Color
	mirror bool
}

// demoSprites lists all sprites to display, grouped for visual assessment (will be removed).
//
//nolint:mnd // temporary demo layout with hardcoded positions
var demoSprites = []demoEntry{
	// Row 1: Player sprites
	{label: "PLV", id: SpritePlayerLevel, ink: ColorPlayer1},
	{label: "PBK", id: SpritePlayerBanked, ink: ColorPlayer1},
	{label: "P2L", id: SpritePlayerLevel, ink: ColorPlayer2},
	{label: "MIS", id: SpritePlayerMissile, ink: ColorMissile},
	{label: "TRL", id: SpriteMissileTrail, ink: ColorMissile},

	// Row 2: Enemies (left-facing)
	{label: "HEL", id: SpriteHelicopterReg, ink: ColorHelicopter},
	{label: "HAD", id: SpriteHelicopterAdv, ink: ColorHelicopter},
	{label: "SHP", id: SpriteShip, ink: ColorShip},
	{label: "FGT", id: SpriteFighter, ink: ColorBlue},
	{label: "TNK", id: SpriteTankBody, ink: ColorBlue},

	// Row 3: Enemies (mirrored)
	{label: "HEm", id: SpriteHelicopterReg, ink: ColorHelicopter, mirror: true},
	{label: "HAm", id: SpriteHelicopterAdv, ink: ColorHelicopter, mirror: true},
	{label: "SHm", id: SpriteShip, ink: ColorShip, mirror: true},
	{label: "FGm", id: SpriteFighter, ink: ColorBlue, mirror: true},
	{label: "TNm", id: SpriteTankBody, ink: ColorBlue, mirror: true},

	// Row 4: Rotors, caterpillars
	{label: "RTL", id: SpriteRotorLeft, ink: ColorHelicopter},
	{label: "RTR", id: SpriteRotorRight, ink: ColorHelicopter},
	{label: "CT0", id: SpriteTankCaterpillar0, ink: ColorBlue},
	{label: "CT1", id: SpriteTankCaterpillar1, ink: ColorBlue},
	{label: "CT2", id: SpriteTankCaterpillar2, ink: ColorBlue},

	// Row 5: Balloon, fuel depot
	{label: "BAL", id: SpriteBalloon, ink: ColorBalloon},
	{label: "FUE", id: SpriteFuelDepot, ink: ColorFuel},

	// Row 6: Explosions
	{label: "EXS", id: SpriteExplosionSmall, ink: ColorExplosion},
	{label: "EXM", id: SpriteExplosionMedium, ink: ColorExplosion},
	{label: "EXL", id: SpriteExplosionLarge, ink: ColorExplosion},

	// Row 7: Shell explosions
	{label: "SE0", id: SpriteShellExplosion0, ink: ColorExplosion},
	{label: "SE1", id: SpriteShellExplosion1, ink: ColorExplosion},
	{label: "SE2", id: SpriteShellExplosion2, ink: ColorExplosion},
	{label: "SE3", id: SpriteShellExplosion3, ink: ColorExplosion},
	{label: "SE4", id: SpriteShellExplosion4, ink: ColorExplosion},
	{label: "SE5", id: SpriteShellExplosion5, ink: ColorExplosion},

	// Row 8: Rocks
	{label: "RK0", id: SpriteRock0, ink: ColorRock},
	{label: "RK1", id: SpriteRock1, ink: ColorRock},
	{label: "RK2", id: SpriteRock2, ink: ColorRock},
	{label: "RK3", id: SpriteRock3, ink: ColorRock},
}

// Game implements the ebiten.Game interface.
type Game struct{}

func (g *Game) Update() error {
	return nil
}

//nolint:mnd // temporary demo rendering with hardcoded grid layout
func (g *Game) Draw(screen *ebiten.Image) {
	renderText(screen, demoSpans)

	const (
		colWidth  = 28
		rowHeight = 30
		startY    = 18
		startX    = 2
		cols      = 9
	)

	for i, e := range demoSprites {
		col := i % cols
		row := i / cols

		x := startX + col*colWidth
		y := startY + row*rowHeight

		s := SpriteCatalog[e.id]
		ink := Palette[e.ink]

		drawSprite(screen, s, x, y, ink, e.mirror)

		// Label below each sprite.
		labelRow := (y + s.Height + 1) / 8
		labelCol := x / 8

		renderText(screen, []TextSpan{
			{Row: labelRow, Col: labelCol, Ink: ColorWhite, Text: e.label},
		})
	}
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}
