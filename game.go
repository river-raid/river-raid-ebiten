package main

import "github.com/hajimehoshi/ebiten/v2"

// terrainBufferHeight is the total height of the terrain buffer in pixels.
// Must be large enough for the viewport plus lookahead for scrolling.
const terrainBufferHeight = ViewportHeight + fragmentsPerLevel*profileSize

// Player movement constant.
const planeMovementStep = 2

// Scroll-in sub-states.
const (
	scrollInFrames    = 40
	scrollInScrolling = 0
	scrollInWaiting   = 1
)

// Game implements the ebiten.Game interface.
type Game struct {
	terrain       *TerrainBuffer
	viewport      Viewport
	scroll        ScrollState
	missile       Missile
	tankShell     TankShell
	heliMissile   HeliMissile
	planeX        int
	fuel          int
	scrollInCount int
	scrollInState int
	screen        GameScreen
	mode          GameplayMode
	speed         Speed
	currentPlayer Player
	planeBanked   bool
	paused        bool
	inited        bool
}

func (g *Game) init() {
	g.terrain = newTerrainBuffer(terrainBufferHeight)
	g.scroll.InitScroll(terrainBufferHeight)

	// Pre-fill the buffer with enough fragments to cover the viewport.
	// Render from the viewport top upward so the initial screen is filled.
	initialFragments := (ViewportHeight + profileSize - 1) / profileSize
	renderY := g.scroll.ScrollY

	for range initialFragments {
		frag := g.scroll.nextFragment()
		g.terrain.renderFragment(frag, renderY, true)
		renderY += profileSize
	}

	g.viewport = NewViewport()
	g.inited = true
	g.startScrollIn()
}

// resetPerLife resets all per-life state (called at scroll-in).
func (g *Game) resetPerLife() {
	g.planeX = PlaneStartX
	g.fuel = FuelLevelFull
	g.speed = SpeedNormal
	g.planeBanked = false
}

// startScrollIn begins the scroll-in sequence.
func (g *Game) startScrollIn() {
	g.screen = ScreenGameplay
	g.mode = GameplayScrollIn
	g.scrollInCount = 0
	g.scrollInState = scrollInScrolling
	g.resetPerLife()
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
	switch g.mode {
	case GameplayScrollIn:
		g.updateScrollIn()
	case GameplayNormal, GameplayRefuel:
		g.updateNormalGameplay()
	case GameplayOverview:
	}
}

func (g *Game) updateScrollIn() {
	switch g.scrollInState {
	case scrollInScrolling:
		g.scroll.advanceLines(g.terrain, int(SpeedFast))
		g.scrollInCount++

		if g.scrollInCount >= scrollInFrames {
			g.scrollInState = scrollInWaiting
		}
	case scrollInWaiting:
		// Wait for any gameplay input (not Enter) to begin.
		input := ScanGameplay()
		if input.Left || input.Right || input.Up || input.Down || input.Fire {
			g.mode = GameplayNormal
		}
	}
}

func (g *Game) updateNormalGameplay() {
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

	// Step 5: Process viewport objects.
	MoveEnemies(&g.viewport)

	// Step 6: Animate player missile.
	g.missile.Update()

	// Step 7: Process tank shell.
	g.tankShell.Update(g.viewport.Tick)

	// Step 8: Process helicopter missile.
	g.heliMissile.Update()

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

	if input.Fire {
		g.missile.Fire(g.planeX)
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

	// Draw viewport objects.
	drawViewportSlots(screen, &g.viewport)

	// Draw projectiles.
	g.missile.Draw(screen)
	g.tankShell.Draw(screen)
	g.heliMissile.Draw(screen)

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
