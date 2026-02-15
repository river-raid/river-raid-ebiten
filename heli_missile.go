package main

import "image/draw"

// Helicopter missile constants.
const (
	heliMissileHorizSpeed = 8
	heliMissileSpawnOffY  = 4
	heliMissileAlignMask  = 0xF8 // align X to 8-pixel boundary
)

// HeliMissile tracks the advanced helicopter's missile state.
type HeliMissile struct {
	X           int
	Y           int
	Orientation Orientation
	Active      bool
}

// Fire launches a helicopter missile. Does nothing if one is already active.
func (hm *HeliMissile) Fire(heliX, heliY int, orient Orientation) {
	if hm.Active {
		return
	}

	hm.X = heliX & heliMissileAlignMask
	hm.Y = heliY + heliMissileSpawnOffY
	hm.Orientation = orient
	hm.Active = true
}

// Update advances the missile diagonally and removes it at viewport boundaries.
func (hm *HeliMissile) Update() {
	if !hm.Active {
		return
	}

	if hm.Orientation == OrientationLeft {
		hm.X -= heliMissileHorizSpeed
	} else {
		hm.X += heliMissileHorizSpeed
	}

	hm.Y++ // moves downward

	if hm.X < 0 || hm.X >= ScreenWidth || hm.Y >= ViewportHeight {
		hm.Active = false
	}
}

// Draw renders the helicopter missile as a small projectile.
func (hm *HeliMissile) Draw(screen draw.Image) {
	if !hm.Active {
		return
	}

	ink := Palette[ColorYellow]
	screen.Set(hm.X, hm.Y, ink)
	screen.Set(hm.X+1, hm.Y, ink)
}
