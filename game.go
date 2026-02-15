package main

import "github.com/hajimehoshi/ebiten/v2"

// terrainBufferHeight is the total height of the terrain buffer in pixels.
// Must be large enough for the viewport plus lookahead for scrolling.
const terrainBufferHeight = ViewportHeight + fragmentsPerLevel*fragmentLines

// Game implements the ebiten.Game interface.
type Game struct {
	terrain *TerrainBuffer
	scroll  ScrollState
	screen  GameScreen
	mode    GameplayMode
	paused  bool
	inited  bool
}

func (g *Game) init() {
	g.screen = ScreenGameplay
	g.mode = GameplayNormal

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
		return
	}

	g.scroll.advanceLines(g.terrain, int(SpeedNormal))
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
}

func (g *Game) drawGameOver(_ *ebiten.Image) {
}
