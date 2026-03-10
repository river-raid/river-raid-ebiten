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

// Game implements the ebiten.Game interface.
type Game struct {
	terrain               *render.TerrainBuffer
	state                 *state.GameState
	controlSelectionPhase int // 0 = control type menu, 1 = game mode dialog
	controlSelectionTimer int // countdown for timeout (phase 0 only)
}

// NewGame creates a new Game instance.
func NewGame() *Game {
	return &Game{
		terrain:               render.NewTerrainBuffer(),
		state:                 state.NewGameState(),
		controlSelectionTimer: controlSelectionTimeout,
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
		if g.handleGameplayEnter() {
			return nil
		}

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

// handleGameplayEnter checks for Enter+modifier combos during gameplay.
// Returns true if a transition was triggered (caller should skip further update).
func (g *Game) handleGameplayEnter() bool {
	switch {
	case input.IsRestartPressed():
		// Caps+Enter: restart gameplay with the same config, skipping control selection
		// and instructions.
		g.state.ResetForNewGame()
		logic.ResetPerLife(g.state, g.terrain)

		return true

	case input.IsControlSelectPressed():
		// Symbol+Enter: return to control selection screen.
		g.state = state.NewGameState()
		g.controlSelectionPhase = 0
		g.controlSelectionTimer = controlSelectionTimeout

		return true
	}

	return false
}

func (g *Game) updateInstructions() {
	if input.IsEnterPressed() {
		g.state.Screen = domain.ScreenGameplay
	}
}

func (g *Game) updateOverview() {
}

func (g *Game) updateGameOver() {
}

func (g *Game) drawControlSelection(screen *ebiten.Image) {
	render.DrawControlSelection(screen, g.controlSelectionPhase)
}

func (g *Game) drawInstructions(screen *ebiten.Image) {
	render.DrawInstructions(screen)
}

func (g *Game) drawOverview(_ *ebiten.Image) {
}

func (g *Game) drawGameplay(screen *ebiten.Image) {
	render.DrawGameplay(screen, g.state, g.terrain)
}

func (g *Game) drawGameOver(_ *ebiten.Image) {
}
