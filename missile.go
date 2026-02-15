package main

import "image/draw"

// Missile constants.
const (
	missileSpeed     = 6
	missileSpawnOffX = 4
	missileSpawnY    = 126
	missileTopY      = 8
	missileSoundY    = 112
)

// Missile tracks the player's missile state.
type Missile struct {
	X      int
	Y      int
	Active bool
}

// Fire launches a missile from the player's current position.
// Does nothing if a missile is already active.
func (m *Missile) Fire(planeX int) {
	if m.Active {
		return
	}

	m.X = planeX + missileSpawnOffX
	m.Y = missileSpawnY
	m.Active = true
}

// Update advances the missile upward and deactivates it at the top of the screen.
func (m *Missile) Update() {
	if !m.Active {
		return
	}

	m.Y -= missileSpeed

	if m.Y < missileTopY {
		m.Active = false
	}
}

// Draw renders the missile and its trail.
func (m *Missile) Draw(screen draw.Image) {
	if !m.Active {
		return
	}

	ink := Palette[ColorMissile]
	s := SpriteCatalog[SpritePlayerMissile]
	drawSprite(screen, s, m.X, m.Y, ink, false)

	// Trail behind the missile.
	trail := SpriteCatalog[SpriteMissileTrail]
	drawSprite(screen, trail, m.X, m.Y+s.Height(), ink, false)
}
