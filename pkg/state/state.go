package state

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/input"
)

// ControlFlags holds the expanded state of the original control byte.
type ControlFlags struct {
	Speed     domain.Speed
	FireSound bool
	LowFuel   bool
	FuelFull  bool
	BonusLife bool
	Exploding bool
}

// PlayerState holds a per-player state that persists across lives.
type PlayerState struct {
	Score         int
	Lives         int
	BridgeIndex   int
	BridgeCounter int
}

// ExplosionFragment represents the position of a single active explosion fragment.
// All fragments share the same animation frame, tracked in Explosion.Frame.
type ExplosionFragment struct {
	X int
	Y int
}

// Explosion holds all active explosion fragments and the shared animation frame.
// Frame is the current animation frame; fragments are removed once Frame
// advances past the last frame.
type Explosion struct {
	Fragments []ExplosionFragment
	Frame     int
}

// GameState holds all mutable game state.
type GameState struct {
	Viewport        *Viewport
	Missile         *PlayerMissile
	TankShell       *TankShell
	HeliMissile     *HeliMissile
	InputInterface  input.Interface
	Explosion       Explosion
	Players         [2]PlayerState
	HighScores      [4]int
	Controls        ControlFlags
	Config          domain.GameConfig
	BridgeYPosition int
	BridgeFragBufY  int                    // buffer Y of the current bridge fragment (for re-render on destruction)
	BridgeFragment  assets.TerrainFragment // the current bridge fragment (for re-render on destruction)
	GameplayMode    domain.GameplayMode
	BridgeIndex     int
	FragmentNum     int
	LineInFrag      int
	NextRenderY     int
	ScrollY         int
	PlaneSpriteBank int
	ScrollInCount   int
	ScrollInState   int
	DyingFrame      int
	PlaneX          int
	Fuel            int
	Speed           domain.Speed
	Screen          domain.GameScreen
	CollisionMode   domain.CollisionMode
	CurrentPlayer   domain.Player
	GameNumber      int
	ScrollOffset    uint16
	Tick            uint8
	Paused          bool
	BridgeSection   bool
	BridgeDestroyed bool
	OverviewMode    bool
}

// NewGameState creates a new GameState with persistent state only.
// Per-life state (fuel, position, viewport, scroll, etc.) is initialised separately
// by logic.ResetPerLife, which is called before gameplay starts (in game.applyGameMode)
// and on every respawn — ensuring a single code path for all life starts.
func NewGameState() *GameState {
	return &GameState{
		Players: [2]PlayerState{
			{Lives: domain.LivesInitial},
			{Lives: domain.LivesInitial},
		},
	}
}

// ResetForNewGame resets per-game state (scores, lives, bridge position) using the
// current Config, keeping Config and InputInterface intact.
// Call logic.ResetPerLife after this to complete the per-life reset.
func (s *GameState) ResetForNewGame() {
	bridgeCounter := int(s.Config.StartingBridge)
	s.Players[domain.Player1] = PlayerState{Lives: domain.LivesInitial, BridgeCounter: bridgeCounter}
	s.Players[domain.Player2] = PlayerState{Lives: domain.LivesInitial, BridgeCounter: bridgeCounter}
	s.BridgeIndex = bridgeCounter - 1
	s.GameplayMode = domain.GameplayScrollIn
}
