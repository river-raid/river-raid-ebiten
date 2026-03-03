package state

import "github.com/morozov/river-raid-ebiten/pkg/domain"

// PlayerMissile tracks the player's missile state.
type PlayerMissile struct {
	X      int
	Y      int
	Active bool
}

// TankShell tracks the tank shell projectile state.
type TankShell struct {
	X              int
	Y              int
	Speed          int // 1-4 horizontal pixels per frame
	TrajectoryStep int // 0-7
	ExplosionFrame int
	Orientation    domain.Orientation
	IsFlying       bool
	IsExploding    bool
}

// HeliMissile tracks the advanced helicopter's missile state.
type HeliMissile struct {
	X           int
	Y           int
	Orientation domain.Orientation
	Active      bool
}
