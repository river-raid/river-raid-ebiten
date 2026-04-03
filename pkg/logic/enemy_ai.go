package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Physical enemy speeds (px/sec).
const (
	standardSpeedPxSec = 12 // ships, helicopters, road and bank tanks
	balloonSpeedPxSec  = 6
	fighterSpeedPxSec  = 48
)

// Derived enemy movement steps (sp/tick).
const (
	speedStandard domain.SP = standardSpeedPxSec * domain.SubpixelScale / domain.Tps
	speedBalloon  domain.SP = balloonSpeedPxSec * domain.SubpixelScale / domain.Tps
	speedFighter  domain.SP = fighterSpeedPxSec * domain.SubpixelScale / domain.Tps
)

// Fighter wrap/reset positions in subpixels.
const (
	fighterWrapLeftX   domain.SP = 0
	fighterWrapRightX  domain.SP = 232 * domain.SubpixelScale
	fighterResetLeftX  domain.SP = 232 * domain.SubpixelScale
	fighterResetRightX domain.SP = 4 * domain.SubpixelScale
)

// Boundary calculation constants.
const (
	// boundaryPaddingPx defines the horizontal padding in screen pixels.
	boundaryPaddingPx domain.Px = 8
)

// TerrainBuffer is an interface for querying terrain edges.
// This allows us to avoid importing the render package directly.
type TerrainBuffer interface {
	GetEdges(x, y domain.Px, spriteHeight int) (leftX, rightX domain.Px)
}

// initializeObjectBoundaries calculates movement boundaries for a newly spawned object.
// This is called once at spawn time to set MinX/MaxX based on terrain.
// scrollY is the current scroll position in the terrain buffer.
func initializeObjectBoundaries(obj *state.ViewportObject, terrain TerrainBuffer, scrollY domain.SP) {
	// Skip rocks and fuel depots (they don't move).
	if obj.IsRock || obj.Type == domain.ObjectFuel {
		return
	}

	// Get sprite dimensions for this enemy type.
	sprite := assets.SpriteObjects[obj.Type]
	spriteWidth := domain.Px(sprite.Width)
	spriteHeight := sprite.Height

	// Query terrain edges across all scanlines the sprite overlaps.
	// Pass enemy's spawn X position to determine which shoulder it's in.
	// obj.X and scrollY are in subpixels; shift to screen pixels for terrain query.
	leftEdge, rightEdge := terrain.GetEdges(obj.X.ToPx(), scrollY.ToPx(), spriteHeight)

	if obj.Type == domain.ObjectTank && obj.TankLocation == domain.TankLocationBank {
		if obj.X.ToPx() < leftEdge {
			obj.MinX = 0
			obj.MaxX = (leftEdge - spriteWidth - boundaryPaddingPx).ToSP()
		} else {
			obj.MinX = (rightEdge + boundaryPaddingPx).ToSP()
			obj.MaxX = (domain.Px(platform.ScreenWidth) - spriteWidth).ToSP()
		}
	} else {
		obj.MinX = (leftEdge + boundaryPaddingPx).ToSP()
		obj.MaxX = (rightEdge - spriteWidth - boundaryPaddingPx).ToSP()
	}
}

// convertToBankTank switches a road tank to bank-tank behavior after bridge destruction
// on a late level (spec §7.5.4). Bounds are derived from the bridge gap position so the
// tank cannot cross the gap — it will stop at the bank edge and fire.
func convertToBankTank(obj *state.ViewportObject) {
	obj.TankLocation = domain.TankLocationBank
	sprite := assets.SpriteObjects[domain.ObjectTank]
	if obj.X.ToPx()+tankGapProbe < tankGapLeftEdge {
		obj.MinX = 0
		obj.MaxX = (tankGapLeftEdge - domain.Px(sprite.Width) - boundaryPaddingPx).ToSP()
	} else {
		obj.MinX = (tankGapRightEdge + boundaryPaddingPx).ToSP()
		obj.MaxX = (domain.Px(platform.ScreenWidth) - domain.Px(sprite.Width)).ToSP()
	}
}

// moveEnemies updates all activated enemy positions based on their type-specific AI.
// tick is the global game tick (s.Tick), used for movement timing.
// gameplayMode is used to suppress helicopter missile firing during scroll-in.
// bridgeDestroyed freezes road tank movement when true.
func moveEnemies(vp *state.Viewport, tick int, ts *state.TankShell, hm *state.HeliMissile, gameplayMode domain.GameplayMode, bridgeDestroyed bool) {
	for i := range vp.Objects {
		obj := vp.Objects[i]
		if !obj.Activated {
			continue
		}

		switch obj.Type {
		case domain.ObjectHelicopterReg, domain.ObjectShip:
			moveShipOrHelicopter(obj)
		case domain.ObjectHelicopterAdv:
			moveShipOrHelicopter(obj)
			if gameplayMode != domain.GameplayScrollIn && gameplayMode != domain.GameplayOverview {
				FireHeliMissile(hm, obj.X, obj.Y, obj.Orientation)
			}
		case domain.ObjectFighter:
			moveFighter(obj)
		case domain.ObjectTank:
			moveTank(obj, tick, ts, bridgeDestroyed)
		case domain.ObjectBalloon:
			moveBalloon(obj)
		case domain.ObjectFuel:
			// Fuel depots are static.
		}
	}
}

// moveShipOrHelicopter moves at speedStandard sp/tick, reversing at terrain boundaries.
func moveShipOrHelicopter(obj *state.ViewportObject) {
	if obj.Orientation == domain.OrientationLeft {
		obj.X -= speedStandard
		if obj.X <= obj.MinX {
			obj.Orientation = domain.OrientationRight
		}
	} else {
		obj.X += speedStandard
		if obj.X >= obj.MaxX {
			obj.Orientation = domain.OrientationLeft
		}
	}
}

// moveFighter moves at speedFighter sp/tick, wrapping at screen edges.
func moveFighter(obj *state.ViewportObject) {
	if obj.Orientation == domain.OrientationLeft {
		obj.X -= speedFighter
		if obj.X <= fighterWrapLeftX {
			obj.X = fighterResetLeftX
		}
	} else {
		obj.X += speedFighter
		if obj.X >= fighterWrapRightX {
			obj.X = fighterResetRightX
		}
	}
}

// moveTank moves at speedStandard sp/tick.
// Road tanks are frozen when bridgeDestroyed is true (their X is checked separately by
// applyBridgeDestroyedTanks). Bank tanks move along the bank until the river edge is
// reached, then stop permanently and fire repeatedly.
func moveTank(obj *state.ViewportObject, tick int, ts *state.TankShell, bridgeDestroyed bool) {
	switch obj.TankLocation {
	case domain.TankLocationRoad:
		if bridgeDestroyed {
			// Bridge is gone: movement frozen; gap check runs in bridgeTarget.onHit.
			return
		}

		if obj.Orientation == domain.OrientationLeft {
			obj.X -= speedStandard
		} else {
			obj.X += speedStandard
		}

	case domain.TankLocationBank:
		// Terrain probe: solid terrain is still ahead when the tank has not yet
		// reached the river edge boundary.
		terrainAhead := (obj.Orientation == domain.OrientationLeft && obj.X > obj.MinX) ||
			(obj.Orientation == domain.OrientationRight && obj.X < obj.MaxX)

		if terrainAhead {
			// Still on the solid bank — advance toward the river.
			if obj.Orientation == domain.OrientationLeft {
				obj.X -= speedStandard
			} else {
				obj.X += speedStandard
			}
		} else {
			// River edge reached (or already at it): fire.
			// FireTankShell is a no-op while the shell is flying or exploding,
			// so this naturally implements the fire/wait cycle.
			FireTankShell(ts, shellSpawnX(obj.X, obj.Orientation), obj.Y, tick, obj.Orientation)
		}
	}
}

// moveBalloon moves at speedBalloon sp/tick, reversing at terrain boundaries.
func moveBalloon(obj *state.ViewportObject) {
	if obj.Orientation == domain.OrientationLeft {
		obj.X -= speedBalloon
		if obj.X <= obj.MinX {
			obj.Orientation = domain.OrientationRight
		}
	} else {
		obj.X += speedBalloon
		if obj.X >= obj.MaxX {
			obj.Orientation = domain.OrientationLeft
		}
	}
}
