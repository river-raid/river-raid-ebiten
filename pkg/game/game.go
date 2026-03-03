package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/logic"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/render"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Game implements the ebiten.Game interface.
type Game struct {
	terrain *render.TerrainBuffer
	state   *state.GameState
}

// NewGame creates a new Game instance.
func NewGame() *Game {
	const startingBridge = domain.StartingBridge01
	bridgeIndex := int(startingBridge) - 1

	gs := state.NewGameState(bridgeIndex)
	gs.Config.StartingBridge = startingBridge

	return &Game{
		terrain: render.NewTerrainBuffer(),
		state:   gs,
	}
}

// Update updates a game by one tick.
func (g *Game) Update() error {
	switch g.state.Screen {
	case domain.ScreenControlSelection:
		g.updateControlSelection()
	case domain.ScreenInstructions:
		g.updateInstructions()
	case domain.ScreenOverview:
		g.updateOverview()
	case domain.ScreenGameplay:
		logic.UpdateGameplay(g.state, g.terrain)
	case domain.ScreenGameOver:
		g.updateGameOver()
	}

	return nil
}

// Draw draws the game screen by one frame.
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state.Screen {
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

func (g *Game) updateGameOver() {
}

func (g *Game) drawControlSelection(_ *ebiten.Image) {
}

func (g *Game) drawInstructions(_ *ebiten.Image) {
}

func (g *Game) drawOverview(_ *ebiten.Image) {
}

func (g *Game) drawGameplay(screen *ebiten.Image) {
	render.DrawGameplay(screen, g.state, g.terrain)
}

func (g *Game) drawGameOver(_ *ebiten.Image) {
}
