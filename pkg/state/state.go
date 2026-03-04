package state

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// ControlFlags holds the expanded state of the original control byte.
type ControlFlags struct {
	Speed     domain.Speed
	FireSound bool
	LowFuel   bool
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

// ExplodingFragment represents an active explosion fragment.
type ExplodingFragment struct {
	X     int
	Y     int
	Frame int
}

// GameState holds all mutable game state.
type GameState struct {
	Viewport           *Viewport
	Missile            *PlayerMissile
	TankShell          *TankShell
	HeliMissile        *HeliMissile
	ExplodingFragments []ExplodingFragment
	Players            [2]PlayerState
	HighScores         [4]int
	Controls           ControlFlags
	Config             domain.GameConfig
	BridgeYPosition    int
	BridgeFragBufY     int                    // buffer Y of the current bridge fragment (for re-render on destruction)
	BridgeFragment     assets.TerrainFragment // the current bridge fragment (for re-render on destruction)
	GameplayMode       domain.GameplayMode
	BridgeIndex        int
	FragmentNum        int
	LineInFrag         int
	NextRenderY        int
	ScrollY            int
	PlaneSpriteBank    int
	ScrollInCount      int
	ScrollInState      int
	DyingFrame         int
	PlaneX             int
	Fuel               int
	Speed              domain.Speed
	Screen             domain.GameScreen
	CollisionMode      domain.CollisionMode
	InputInterface     domain.InputInterface
	CurrentPlayer      domain.Player
	ScrollOffset       uint16
	Tick               uint8
	Paused             bool
	BridgeSection      bool
	BridgeDestroyed    bool
	OverviewMode       bool
}

// NewGameState creates a new GameState.
func NewGameState(bridgeIndex int) *GameState {
	const initialScrollOffset = domain.NumLinesPerTerrainProfile
	vp := NewViewport()
	// Align SpawnIndex to the initial ScrollOffset so the first scroll step does not
	// spuriously spawn an object (mirrors the logic in logic.resetPerLife).
	vp.SpawnIndex = initialScrollOffset / domain.NumLinesPerSpawnSlot

	return &GameState{
		Viewport:    vp,
		Missile:     &PlayerMissile{},
		TankShell:   &TankShell{},
		HeliMissile: &HeliMissile{},

		// Initialize for scroll-in
		Screen:       domain.ScreenGameplay,
		GameplayMode: domain.GameplayScrollIn,

		// per-life state
		Fuel:   domain.FuelLevelFull,
		PlaneX: domain.PlaneStartX,
		Speed:  domain.SpeedNormal,

		// per-player state
		Players: [2]PlayerState{
			{Lives: domain.LivesInitial},
			{Lives: domain.LivesInitial},
		},

		// ignore the first terrain fragment
		ScrollY:      initialScrollOffset,
		ScrollOffset: initialScrollOffset,
		FragmentNum:  1,

		BridgeIndex:     bridgeIndex,
		BridgeDestroyed: true,
	}
}
