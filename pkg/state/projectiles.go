package state

import "github.com/morozov/river-raid-ebiten/pkg/domain"

// PlayerMissile tracks the player's missile state.
type PlayerMissile struct {
	X      domain.SP
	Y      domain.SP
	Active bool
}

// TankShell tracks the tank shell projectile state.
type TankShell struct {
	X              domain.SP
	Y              domain.SP
	Speed          domain.SP // sp/tick horizontal velocity
	TrajectoryStep domain.SP
	ExplosionFrame int
	Orientation    domain.Orientation
	IsFlying       bool
	IsExploding    bool
}

// HeliMissile tracks the advanced helicopter's missile state.
type HeliMissile struct {
	X           domain.SP
	Y           domain.SP
	Orientation domain.Orientation
	Active      bool
}
