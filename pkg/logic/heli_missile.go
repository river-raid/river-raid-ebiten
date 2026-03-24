package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Physical helicopter missile speed (px/sec).
const heliMissileSpeedPxSec = 96 // px/sec

// Derived helicopter missile constants (subpixels).
const (
	heliMissileHorizSP     domain.SP = heliMissileSpeedPxSec * domain.SubpixelScale / domain.Tps
	heliMissileSpawnOffYSP domain.SP = 4 * domain.SubpixelScale
	heliMissileAlignSP     domain.SP = 8 * domain.SubpixelScale // 8-pixel boundary
	heliMissileWidth                 = 8                        // sprite width in screen pixels
	heliMissileHeight                = 1                        // sprite height in screen pixels
)

// FireHeliMissile launches a helicopter missile. Does nothing if one is already active.
func FireHeliMissile(hm *state.HeliMissile, heliX, heliY domain.SP, orient domain.Orientation) {
	if hm.Active {
		return
	}

	hm.X = heliX &^ (heliMissileAlignSP - 1)
	hm.Y = heliY + heliMissileSpawnOffYSP
	hm.Orientation = orient
	hm.Active = true
}

// updateHeliMissile advances the missile horizontally and removes it at viewport boundaries
// or on terrain collision.
// Vertical movement is not applied here — the world-scroll system advances Y for all
// viewport objects, so the downward drift is handled externally.
func updateHeliMissile(hm *state.HeliMissile, terrain TerrainBuffer, scrollY domain.SP) {
	if !hm.Active {
		return
	}

	if hm.Orientation == domain.OrientationLeft {
		hm.X -= heliMissileHorizSP
	} else {
		hm.X += heliMissileHorizSP
	}

	if hm.X < 0 || hm.X >= platform.ScreenWidth*domain.SubpixelScale {
		hm.Active = false
		return
	}

	// Remove if any pixel of the missile overlaps a bank.
	startX := hm.X.ToPx()
	if hm.Orientation == domain.OrientationLeft {
		startX = hm.X.ToPx() - heliMissileWidth + 1
	}

	leftEdge, rightEdge := terrain.GetEdges(hm.X.ToPx(), (scrollY + hm.Y).ToPx(), heliMissileHeight)
	if startX < leftEdge || startX+heliMissileWidth > rightEdge {
		hm.Active = false
	}
}
