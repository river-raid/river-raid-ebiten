package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Helicopter missile constants.
const (
	heliMissileHorizSpeed = 8
	heliMissileSpawnOffY  = 4
	heliMissileAlignMask  = 0xF8 // align X to 8-pixel boundary
)

// FireHeliMissile launches a helicopter missile. Does nothing if one is already active.
func FireHeliMissile(hm *state.HeliMissile, heliX, heliY int, orient domain.Orientation) {
	if hm.Active {
		return
	}

	hm.X = heliX & heliMissileAlignMask
	hm.Y = heliY + heliMissileSpawnOffY
	hm.Orientation = orient
	hm.Active = true
}

// updateHeliMissile advances the missile horizontally and removes it at viewport boundaries.
// Vertical movement is not applied here — the world-scroll system advances Y for all
// viewport objects, so the downward drift is handled externally.
func updateHeliMissile(hm *state.HeliMissile) {
	if !hm.Active {
		return
	}

	if hm.Orientation == domain.OrientationLeft {
		hm.X -= heliMissileHorizSpeed
	} else {
		hm.X += heliMissileHorizSpeed
	}

	if hm.X < 0 || hm.X >= platform.ScreenWidth {
		hm.Active = false
	}
}
