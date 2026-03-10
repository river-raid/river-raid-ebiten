package game

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/input"
	"github.com/morozov/river-raid-ebiten/pkg/logic"
)

const (
	controlSelectionTimeout = 500 // frames (~10 s at 50 Hz)
	ctrlTypeCount           = 4
	gameModeCount           = 8
)

// ModeConfig returns the GameConfig for the given 1-based game mode number (1–8).
func ModeConfig(n int) domain.GameConfig {
	bridges := [4]domain.StartingBridge{
		domain.StartingBridge01,
		domain.StartingBridge05,
		domain.StartingBridge20,
		domain.StartingBridge30,
	}

	return domain.GameConfig{
		IsTwoPlayer:    n%2 == 0,
		StartingBridge: bridges[(n-1)/2],
	}
}

func (g *Game) updateControlSelection() {
	if g.controlSelectionPhase == 0 {
		g.controlSelectionTimer--
		if g.controlSelectionTimer <= 0 {
			g.state.Screen = domain.ScreenOverview

			return
		}

		n := input.ScanMenuNumber(ctrlTypeCount)
		if n > 0 {
			g.state.InputInterface = domain.InputInterface(n - 1)
			g.controlSelectionPhase = 1
		}

		return
	}

	// Phase 1: game mode dialog.
	n := input.ScanMenuNumber(gameModeCount)
	if n > 0 {
		g.applyConfig(ModeConfig(n))
		g.state.Screen = domain.ScreenInstructions
	}
}

func (g *Game) applyConfig(cfg domain.GameConfig) {
	bridgeIndex := int(cfg.StartingBridge) - 1

	g.state.Config = cfg
	g.state.BridgeIndex = bridgeIndex
	g.state.Players[domain.Player1].BridgeCounter = int(cfg.StartingBridge)
	g.state.Players[domain.Player2].BridgeCounter = int(cfg.StartingBridge)
	g.state.GameplayMode = domain.GameplayScrollIn

	logic.ResetPerLife(g.state, g.terrain)
}
