package state

import (
	"fmt"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/input"
)

// ScrollInPhase tracks whether the player has taken control of the current life.
type ScrollInPhase int

// ScrollInPhase values.
const (
	// ScrollInScrolling is every phase but the wait: the terrain scrolling into
	// view, and play once the player has taken control.
	ScrollInScrolling ScrollInPhase = iota
	// ScrollInWaiting runs from the terrain being in place until the player
	// presses any control.
	ScrollInWaiting
)

// SoundFlags holds the flags indicating which sounds should be played.
type SoundFlags struct {
	Speed     domain.Speed
	FuelState FuelState
	Firing    bool
	Exploding bool
	BonusLife bool
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
	X domain.SP
	Y domain.SP
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
	HighScores      map[domain.StartingBridge]int
	InputInterface  input.Interface
	Explosion       Explosion
	Players         [2]PlayerState
	Sounds          SoundFlags
	Config          domain.GameConfig
	BridgeYPosition domain.SP
	BridgeFragBufY  int                    // buffer Y of the current bridge fragment (for re-render on destruction)
	BridgeFragment  assets.TerrainFragment // the current bridge fragment (for re-render on destruction)
	GameplayMode    domain.GameplayMode
	BridgeIndex     int
	FragmentNum     int
	LineInFrag      int
	NextRenderY     int
	ScrollY         domain.SP
	PlaneSpriteBank int
	ScrollInCount   int
	ScrollInState   ScrollInPhase
	DyingFrame      int
	PlaneX          domain.SP
	Fuel            int
	Speed           domain.Speed
	Screen          domain.GameScreen
	CollisionMode   domain.CollisionMode
	CurrentPlayer   domain.Player
	GameNumber      int
	ScrollOffset    int
	Tick            int
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
		HighScores:     make(map[domain.StartingBridge]int),
		InputInterface: input.InterfaceFor(0),
	}
}

// Halted reports whether the game is stopped waiting on the player: paused, or
// holding after a scroll-in until the player takes control of the new life.
//
// Nothing advances while it is true. Anything that must freeze with the game
// reads this rather than testing the underlying states, so the behaviors cannot
// drift apart.
func (s *GameState) Halted() bool {
	switch s.ScrollInState {
	case ScrollInWaiting:
		return true
	case ScrollInScrolling:
		return s.Paused
	default:
		panic(fmt.Sprintf("Halted: unexpected ScrollInPhase %d", s.ScrollInState))
	}
}

// Advance moves the game clock on by one tick unless the game is halted.
//
// Tick drives the real-time animations, so holding it still freezes the
// helicopter blades and the fuel station.
func (s *GameState) Advance() {
	if !s.Halted() {
		s.Tick++
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
	s.Paused = false
}
