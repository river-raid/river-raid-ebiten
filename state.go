package main

// GameState holds all mutable game state.
type GameState struct {
	ExplodingFragments []ExplodingFragment

	// Players
	Players    [2]PlayerState
	HighScores [4]int // one per starting bridge

	// Viewport
	Slots [15]Slot

	// Plane (reset each life)
	PlaneX          int
	PlaneSpriteBank int
	Fuel            int // 0-255

	// Core
	GameplayMode GameplayMode
	Speed        Speed

	// Collision
	CollisionMode CollisionMode

	// Control flags (sound/effect triggers)
	Controls ControlFlags

	// Configuration
	Config         GameConfig
	InputInterface InputInterface
	CurrentPlayer  Player

	ScrollOffset uint16 // wrapping 16-bit, overflow is intentional
	Tick         uint8  // wrapping 0-255 counter, overflow is intentional

	Paused          bool
	BridgeSection   bool
	BridgeDestroyed bool
	OverviewMode    bool
}
