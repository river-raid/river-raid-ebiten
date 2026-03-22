package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// PlayerMissile constants.
const (
	missileSpeed     = 6
	missileSpawnOffX = 3
	missileSpawnOffY = 8
	missileTopY      = 8
	missileSoundY    = 112
)

// FireMissile launches a missile from the player's current position.
// Does nothing if a missile is already active.
func FireMissile(m *state.PlayerMissile, planeX int) {
	if m.Active {
		return
	}

	m.X = planeX + missileSpawnOffX
	m.Y = domain.PlaneY - missileSpawnOffY
	m.Active = true
}

// updateMissile advances the missile upward and deactivates it at the top of the screen.
// The missile tracks the player's current horizontal position each frame.
func updateMissile(m *state.PlayerMissile, planeX int) {
	if !m.Active {
		return
	}

	m.X = planeX + missileSpawnOffX
	m.Y -= missileSpeed

	if m.Y < missileTopY {
		m.Active = false
	}
}
