package logic

// PlayerMissile constants.
const (
	missileSpeed     = 6
	missileSpawnOffX = 4
	missileSpawnY    = 126
	missileTopY      = 8
	missileSoundY    = 112
)

// PlayerMissile tracks the player's missile state.
type PlayerMissile struct {
	X      int
	Y      int
	Active bool
}

// Fire launches a missile from the player's current position.
// Does nothing if a missile is already active.
func (m *PlayerMissile) Fire(planeX int) {
	if m.Active {
		return
	}

	m.X = planeX + missileSpawnOffX
	m.Y = missileSpawnY
	m.Active = true
}

// Update advances the missile upward and deactivates it at the top of the screen.
func (m *PlayerMissile) Update() {
	if !m.Active {
		return
	}

	m.Y -= missileSpeed

	if m.Y < missileTopY {
		m.Active = false
	}
}
