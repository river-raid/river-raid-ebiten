package main

import "github.com/hajimehoshi/ebiten/v2"

// terrainBufferHeight is the total height of the terrain buffer in pixels.
// Must be large enough for the viewport plus lookahead for scrolling.
const terrainBufferHeight = ViewportHeight + fragmentsPerLevel*fragmentLines

// Game implements the ebiten.Game interface.
type Game struct {
	terrain *TerrainBuffer
	scroll  ScrollState
	inited  bool
}

func (g *Game) init() {
	g.terrain = newTerrainBuffer(terrainBufferHeight)

	// Pre-fill the buffer with enough fragments to cover the viewport.
	initialFragments := (ViewportHeight + fragmentLines - 1) / fragmentLines
	for range initialFragments {
		frag := g.scroll.nextFragment()
		g.terrain.renderFragment(frag, g.scroll.GeneratedY)
		g.scroll.GeneratedY += fragmentLines
	}

	g.inited = true
}

func (g *Game) Update() error {
	if !g.inited {
		g.init()
	}

	g.scroll.advanceLines(g.terrain, int(SpeedNormal))

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.inited {
		return
	}

	drawTerrainBuffer(screen, g.terrain, g.scroll.ScrollY)
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}
