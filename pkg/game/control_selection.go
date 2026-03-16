package game

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/input"
	"github.com/morozov/river-raid-ebiten/pkg/logic"
)

const (
	controlSelectionTimeout = 10 * Tps // 10 seconds
	ctrlTypeCount           = 4
	gameConfigCount         = 8
)

// gameConfigs contains available game configurations.
var gameConfigs = [gameConfigCount]domain.GameConfig{
	{StartingBridge: domain.StartingBridge01, IsTwoPlayer: false},
	{StartingBridge: domain.StartingBridge01, IsTwoPlayer: true},
	{StartingBridge: domain.StartingBridge05, IsTwoPlayer: false},
	{StartingBridge: domain.StartingBridge05, IsTwoPlayer: true},
	{StartingBridge: domain.StartingBridge20, IsTwoPlayer: false},
	{StartingBridge: domain.StartingBridge20, IsTwoPlayer: true},
	{StartingBridge: domain.StartingBridge30, IsTwoPlayer: false},
	{StartingBridge: domain.StartingBridge30, IsTwoPlayer: true},
}

func (g *Game) updateControlSelection() {
	if g.controlSelectionPhase == 0 {
		g.controlSelectionTimer--
		if g.controlSelectionTimer <= 0 {
			g.initOverview()

			return
		}

		n := input.ScanMenuNumber(ctrlTypeCount)
		if n > 0 {
			g.state.InputInterface = input.InterfaceFor(n - 1)
			g.controlSelectionPhase = 1
		}

		return
	}

	// Phase 1: game mode dialog.
	n := input.ScanMenuNumber(gameConfigCount)
	if n > 0 {
		g.applyConfig(gameConfigs[n-1])
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
