package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/input"
	"github.com/morozov/river-raid-ebiten/pkg/logic"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/render"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// terrainBufferHeight is the total height of the terrain buffer in pixels.
// Sized as viewport + one fragment lookahead.
const terrainBufferHeight = domain.ViewportHeight + domain.NumLinesPerTerrainProfile

// Player movement constant.
const planeMovementStep = 2

// Scroll-in sub-states.
const (
	scrollInFrames    = 42
	scrollInScrolling = 0
	scrollInWaiting   = 1
)

// Game implements the ebiten.Game interface.
type Game struct {
	terrain       *render.TerrainBuffer
	viewport      state.Viewport
	scroll        logic.ScrollState
	missile       logic.PlayerMissile
	tankShell     logic.TankShell
	heliMissile   logic.HeliMissile
	planeX        int
	fuel          int
	scrollInCount int
	scrollInState int
	screen        domain.GameScreen
	mode          domain.GameplayMode
	speed         domain.Speed
	currentPlayer domain.Player
	planeBanked   bool
	paused        bool
	inited        bool
}

func (g *Game) init() {
	g.terrain = render.NewTerrainBuffer(terrainBufferHeight)
	g.scroll.InitScroll(terrainBufferHeight)

	g.viewport = state.NewViewport()
	g.inited = true
	g.startScrollIn()
}

// resetPerLife resets all per-life state (called at scroll-in).
func (g *Game) resetPerLife() {
	g.planeX = domain.PlaneStartX
	g.fuel = domain.FuelLevelFull
	g.speed = domain.SpeedNormal
	g.planeBanked = false
}

// startScrollIn begins the scroll-in sequence.
func (g *Game) startScrollIn() {
	g.screen = domain.ScreenGameplay
	g.mode = domain.GameplayScrollIn
	g.scrollInCount = 0
	g.scrollInState = scrollInScrolling
	g.resetPerLife()
}

// Update updates a game by one tick.
func (g *Game) Update() error {
	if !g.inited {
		g.init()
	}

	switch g.screen {
	case domain.ScreenControlSelection:
		g.updateControlSelection()
	case domain.ScreenInstructions:
		g.updateInstructions()
	case domain.ScreenOverview:
		g.updateOverview()
	case domain.ScreenGameplay:
		g.updateGameplay()
	case domain.ScreenGameOver:
		g.updateGameOver()
	}

	return nil
}

// Draw draws the game screen by one frame.
func (g *Game) Draw(screen *ebiten.Image) {
	if !g.inited {
		return
	}

	switch g.screen {
	case domain.ScreenControlSelection:
		g.drawControlSelection(screen)
	case domain.ScreenInstructions:
		g.drawInstructions(screen)
	case domain.ScreenOverview:
		g.drawOverview(screen)
	case domain.ScreenGameplay:
		g.drawGameplay(screen)
	case domain.ScreenGameOver:
		g.drawGameOver(screen)
	}
}

// Layout accepts a native outside size in device-independent pixels and returns the game's logical screen
// size in pixels.
func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return platform.ScreenWidth, platform.ScreenHeight
}

func (g *Game) updateControlSelection() {
}

func (g *Game) updateInstructions() {
}

func (g *Game) updateOverview() {
}

func (g *Game) updateGameplay() {
	switch g.mode {
	case domain.GameplayScrollIn:
		g.updateScrollIn()
	case domain.GameplayNormal, domain.GameplayRefuel:
		g.updateNormalGameplay()
	case domain.GameplayOverview:
	}
}

func (g *Game) updateScrollIn() {
	switch g.scrollInState {
	case scrollInScrolling:
		frags := g.scroll.AdvanceLines(int(domain.SpeedFast), terrainBufferHeight)
		for _, f := range frags {
			g.terrain.RenderFragment(f.Fragment, f.Y, true)
		}
		g.scrollInCount++

		if g.scrollInCount >= scrollInFrames {
			g.scrollInState = scrollInWaiting
		}
	case scrollInWaiting:
		// Wait for any gameplay input (not Enter) to begin.
		in := input.ScanGameplay()
		if in.Left || in.Right || in.Up || in.Down || in.Fire {
			g.mode = domain.GameplayNormal
		}
	}
}

func (g *Game) updateNormalGameplay() {
	if g.paused {
		if input.IsPausePressed() {
			g.paused = false
		}

		return
	}

	// Step 1: Check pause.
	if input.IsPausePressed() {
		g.paused = true

		return
	}

	// Step 5: Process viewport objects.
	logic.MoveEnemies(&g.viewport)

	// Step 6: Animate player missile.
	g.missile.Update()

	// Step 7: Process tank shell.
	g.tankShell.Update(g.viewport.Tick)

	// Step 8: Process helicopter missile.
	g.heliMissile.Update()

	// Step 9: Advance scroll at current speed, then reset speed.
	frags := g.scroll.AdvanceLines(int(g.speed), terrainBufferHeight)
	for _, f := range frags {
		g.terrain.RenderFragment(f.Fragment, f.Y, false)
	}
	g.speed = domain.SpeedNormal
	g.planeBanked = false

	// Step 11: Scan in for next frame.
	in := input.ScanGameplay()
	if in.Left {
		g.planeX -= planeMovementStep
		g.planeBanked = true
	}

	if in.Right {
		g.planeX += planeMovementStep
		g.planeBanked = true
	}

	if in.Up {
		g.speed = domain.SpeedFast
	}

	if in.Down {
		g.speed = domain.SpeedSlow
	}

	if in.Fire {
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
	render.DrawTerrainBuffer(screen, g.terrain, g.scroll.ScrollY)

	// Draw viewport objects.
	render.DrawViewportSlots(screen, &g.viewport)

	// Draw projectiles.
	render.DrawPlayerMissile(screen, &g.missile)
	render.DrawTankShell(screen, &g.tankShell)
	render.DrawHeliMissile(screen, &g.heliMissile)

	render.DrawPlayer(screen, g.currentPlayer, g.planeX, g.planeBanked)
}

func (g *Game) drawGameOver(_ *ebiten.Image) {
}
