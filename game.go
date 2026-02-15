package main

import "github.com/hajimehoshi/ebiten/v2"

// terrainBufferHeight is the total height of the terrain buffer in pixels.
// Must be large enough for the viewport plus lookahead for scrolling.
const terrainBufferHeight = ViewportHeight + fragmentsPerLevel*fragmentLines

// Player movement constant.
const planeMovementStep = 2

// Game implements the ebiten.Game interface.
type Game struct {
	terrain       *TerrainBuffer
	scroll        ScrollState
	planeX        int
	screen        GameScreen
	mode          GameplayMode
	speed         Speed
	currentPlayer Player
	planeBanked   bool
	paused        bool
	inited        bool
}

func (g *Game) init() {
	g.screen = ScreenGameplay
	g.mode = GameplayNormal
	g.planeX = PlaneStartX
	g.speed = SpeedNormal

	g.terrain = newTerrainBuffer(terrainBufferHeight)

	// Pre-fill the buffer with enough fragments to cover the viewport.
	initialFragments := (ViewportHeight + fragmentLines - 1) / fragmentLines
	for range initialFragments {
		frag := g.scroll.nextFragment()
		g.terrain.renderFragment(frag, g.scroll.GeneratedY, true)
		g.scroll.GeneratedY += fragmentLines
	}

	g.inited = true
}

func (g *Game) Update() error {
	if !g.inited {
		g.init()
	}

	switch g.screen {
	case ScreenControlSelection:
		g.updateControlSelection()
	case ScreenInstructions:
		g.updateInstructions()
	case ScreenOverview:
		g.updateOverview()
	case ScreenGameplay:
		g.updateGameplay()
	case ScreenGameOver:
		g.updateGameOver()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.inited {
		return
	}

	switch g.screen {
	case ScreenControlSelection:
		g.drawControlSelection(screen)
	case ScreenInstructions:
		g.drawInstructions(screen)
	case ScreenOverview:
		g.drawOverview(screen)
	case ScreenGameplay:
		g.drawGameplay(screen)
	case ScreenGameOver:
		g.drawGameOver(screen)
	}
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) updateControlSelection() {
}

func (g *Game) updateInstructions() {
}

func (g *Game) updateOverview() {
}

func (g *Game) updateGameplay() {
	if g.paused {
		if IsPausePressed() {
			g.paused = false
		}

		return
	}

	// Step 1: Check pause.
	if IsPausePressed() {
		g.paused = true

		return
	}

	// Step 9: Advance scroll at current speed, then reset speed.
	g.scroll.advanceLines(g.terrain, int(g.speed))
	g.speed = SpeedNormal
	g.planeBanked = false

	// Step 11: Scan input for next frame.
	input := ScanGameplay()
	if input.Left {
		g.planeX -= planeMovementStep
		g.planeBanked = true
	}

	if input.Right {
		g.planeX += planeMovementStep
		g.planeBanked = true
	}

	if input.Up {
		g.speed = SpeedFast
	}

	if input.Down {
		g.speed = SpeedSlow
	}
}

func (g *Game) updateGameOver() {
}

func (g *Game) drawControlSelection(_ *ebiten.Image) {
}

func (g *Game) drawInstructions(_ *ebiten.Image) {
}

func (g *Game) drawOverview(_ *ebiten.Image) {
}

func (g *Game) drawGameplay(screen *ebiten.Image) {
	drawTerrainBuffer(screen, g.terrain, g.scroll.ScrollY)

	// Draw player plane.
	spriteID := SpritePlayerLevel
	if g.planeBanked {
		spriteID = SpritePlayerBanked
	}

	ink := Palette[ColorPlayer1]
	if g.currentPlayer == Player2 {
		ink = Palette[ColorPlayer2]
	}

	s := SpriteCatalog[spriteID]
	drawSprite(screen, s, g.planeX, PlaneY, ink, false)
}

func (g *Game) drawGameOver(_ *ebiten.Image) {
}
