package state

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// GameState holds all mutable game state.
type GameState struct {
	ExplodingFragments []domain.ExplodingFragment

	// Players
	Players    [2]domain.PlayerState
	HighScores [4]int // one per starting bridge

	// Plane (reset each life)
	PlaneX          int
	PlaneSpriteBank int
	Fuel            int // 0-255

	// Core
	GameplayMode domain.GameplayMode
	Speed        domain.Speed

	// Collision
	CollisionMode domain.CollisionMode

	// Control flags (sound/effect triggers)
	Controls domain.ControlFlags

	// Configuration
	Config         domain.GameConfig
	InputInterface domain.InputInterface
	CurrentPlayer  domain.Player

	ScrollOffset uint16 // wrapping 16-bit, overflow is intentional
	Tick         uint8  // wrapping 0-255 counter, overflow is intentional

	Paused          bool
	BridgeSection   bool
	BridgeDestroyed bool
	OverviewMode    bool
}
