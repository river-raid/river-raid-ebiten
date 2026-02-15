package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

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
	Orientation domain.Orientation
	Active      bool
}

// Fire launches a helicopter missile. Does nothing if one is already active.
func (hm *HeliMissile) Fire(heliX, heliY int, orient domain.Orientation) {
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

	if hm.Orientation == domain.OrientationLeft {
		hm.X -= heliMissileHorizSpeed
	} else {
		hm.X += heliMissileHorizSpeed
	}

	hm.Y++ // moves downward

	if hm.X < 0 || hm.X >= platform.ScreenWidth || hm.Y >= domain.ViewportHeight {
		hm.Active = false
	}
}
