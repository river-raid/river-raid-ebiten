package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Physical missile speed (px/sec).
const missileSpeedPxSec = 72 // px/sec

// Derived missile constants (subpixels).
const (
	missileSpeed     domain.SP = missileSpeedPxSec * domain.SubpixelScale / domain.Tps
	missileSpawnOffX domain.SP = 3 * domain.SubpixelScale
	missileSpawnOffY domain.SP = 8 * domain.SubpixelScale
	missileTopY      domain.SP = 8 * domain.SubpixelScale
)

// FireMissile launches a missile from the player's current position.
// Does nothing if a missile is already active.
func FireMissile(m *state.PlayerMissile, planeX domain.SP) {
	if m.Active {
		return
	}

	m.X = planeX + missileSpawnOffX
	m.Y = domain.PlaneYSP - missileSpawnOffY
	m.Active = true
}

// updateMissile advances the missile upward and deactivates it at the top of the screen.
// The missile tracks the player's current horizontal position each frame.
func updateMissile(m *state.PlayerMissile, planeX domain.SP) {
	if !m.Active {
		return
	}

	m.X = planeX + missileSpawnOffX
	m.Y -= missileSpeed

	if m.Y < missileTopY {
		m.Active = false
	}
}
