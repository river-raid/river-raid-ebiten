package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Enemy movement constants.
const (
	EnemyMoveStep      = 2
	fighterMoveStep    = 4
	fighterWrapLeftX   = 0
	fighterWrapRightX  = 232
	fighterResetLeftX  = 232
	fighterResetRightX = 4
	balloonTickMask    = 3
	balloonTickMatch   = 1
	evenTickMask       = 1
)

// Boundary calculation constants.
const (
	// boundaryPadding defines the horizontal padding between the river banks
	// and the bounding box of moving objects in pixels
	boundaryPadding = 8
)

// TerrainBuffer is an interface for querying terrain edges.
// This allows us to avoid importing the render package directly.
type TerrainBuffer interface {
	GetEdges(x, y, spriteHeight int) (leftX, rightX int)
}

// initializeObjectBoundaries calculates movement boundaries for a newly spawned object.
// This is called once at spawn time to set MinX/MaxX based on terrain.
// scrollY is the current scroll position in the terrain buffer.
func initializeObjectBoundaries(obj *state.ViewportObject, terrain TerrainBuffer, scrollY int) {
	// Skip rocks and fuel depots (they don't move).
	if obj.IsRock || obj.Type == domain.ObjectFuel {
		return
	}

	// Get sprite dimensions for this enemy type.
	sprite := assets.SpriteObjects[obj.Type]
	spriteWidth := sprite.Width
	spriteHeight := sprite.Height

	// Query terrain edges across all scanlines the sprite overlaps.
	// Pass enemy's spawn X position to determine which shoulder it's in.
	leftEdge, rightEdge := terrain.GetEdges(obj.X, scrollY, spriteHeight)

	if obj.Type == domain.ObjectTank && obj.TankLocation == domain.TankLocationBank {
		if obj.X < leftEdge {
			obj.MinX = 0
			obj.MaxX = leftEdge - spriteWidth - boundaryPadding
		} else {
			obj.MinX = rightEdge + boundaryPadding
			obj.MaxX = platform.ScreenWidth - spriteWidth
		}
	} else {
		obj.MinX = leftEdge + boundaryPadding
		obj.MaxX = rightEdge - spriteWidth - boundaryPadding
	}
}

// moveEnemies updates all activated enemy positions based on their type-specific AI.
// gameplayMode is used to suppress helicopter missile firing during scroll-in.
// bridgeDestroyed freezes road tank movement when true.
func moveEnemies(vp *state.Viewport, ts *state.TankShell, hm *state.HeliMissile, gameplayMode domain.GameplayMode, bridgeDestroyed bool) {
	for i := range vp.Objects {
		obj := vp.Objects[i]
		if !obj.Activated {
			continue
		}

		switch obj.Type {
		case domain.ObjectHelicopterReg, domain.ObjectShip:
			moveShipOrHelicopter(obj, vp.Tick)
		case domain.ObjectHelicopterAdv:
			moveShipOrHelicopter(obj, vp.Tick)
			if gameplayMode != domain.GameplayScrollIn {
				FireHeliMissile(hm, obj.X, obj.Y, obj.Orientation)
			}
		case domain.ObjectFighter:
			moveFighter(obj)
		case domain.ObjectTank:
			moveTank(obj, vp.Tick, ts, bridgeDestroyed)
		case domain.ObjectBalloon:
			moveBalloon(obj, vp.Tick)
		case domain.ObjectFuel:
			// Fuel depots are static.
		}
	}
}

// moveShipOrHelicopter moves 2px on even ticks, reversing at terrain boundaries.
func moveShipOrHelicopter(obj *state.ViewportObject, tick int) {
	if tick&evenTickMask != 0 {
		return
	}

	if obj.Orientation == domain.OrientationLeft {
		obj.X -= EnemyMoveStep
		if obj.X <= obj.MinX {
			obj.Orientation = domain.OrientationRight
		}
	} else {
		obj.X += EnemyMoveStep
		if obj.X >= obj.MaxX {
			obj.Orientation = domain.OrientationLeft
		}
	}
}

// moveFighter moves 4px every frame, wrapping at screen edges.
func moveFighter(obj *state.ViewportObject) {
	if obj.Orientation == domain.OrientationLeft {
		obj.X -= fighterMoveStep
		if obj.X <= fighterWrapLeftX {
			obj.X = fighterResetLeftX
		}
	} else {
		obj.X += fighterMoveStep
		if obj.X >= fighterWrapRightX {
			obj.X = fighterResetRightX
		}
	}
}

// moveTank moves 2px on even ticks.
// Road tanks are frozen when bridgeDestroyed is true (their X is checked separately by
// applyBridgeDestroyedTanks). Bank tanks move along the bank until the river edge is
// reached, then stop permanently and fire repeatedly.
func moveTank(obj *state.ViewportObject, tick int, ts *state.TankShell, bridgeDestroyed bool) {
	if tick&evenTickMask != 0 {
		return
	}

	switch obj.TankLocation {
	case domain.TankLocationRoad:
		if bridgeDestroyed {
			// Bridge is gone: movement frozen; gap check runs in applyBridgeDestroyedTanks.
			return
		}

		if obj.Orientation == domain.OrientationLeft {
			obj.X -= EnemyMoveStep
		} else {
			obj.X += EnemyMoveStep
		}

	case domain.TankLocationBank:
		// Terrain probe: solid terrain is still ahead when the tank has not yet
		// reached the river edge boundary.
		terrainAhead := (obj.Orientation == domain.OrientationLeft && obj.X > obj.MinX) ||
			(obj.Orientation == domain.OrientationRight && obj.X < obj.MaxX)

		if terrainAhead {
			// Still on the solid bank — advance toward the river.
			if obj.Orientation == domain.OrientationLeft {
				obj.X -= EnemyMoveStep
			} else {
				obj.X += EnemyMoveStep
			}
		} else {
			// River edge reached (or already at it): fire.
			// FireTankShell is a no-op while the shell is flying or exploding,
			// so this naturally implements the fire/wait cycle.
			FireTankShell(ts, obj.X, obj.Y, tick, obj.Orientation)
		}
	}
}

// moveBalloon moves 2px every 4th frame, reversing at terrain boundaries.
func moveBalloon(obj *state.ViewportObject, tick int) {
	if tick&balloonTickMask != balloonTickMatch {
		return
	}

	if obj.Orientation == domain.OrientationLeft {
		obj.X -= EnemyMoveStep
		if obj.X <= obj.MinX {
			obj.Orientation = domain.OrientationRight
		}
	} else {
		obj.X += EnemyMoveStep
		if obj.X >= obj.MaxX {
			obj.Orientation = domain.OrientationLeft
		}
	}
}

// Tank gap X bounds. A road tank is in the river gap when X+10 >= $70 and X <= $90.
const (
	tankGapLeftEdge  = 0x70 // X+10 must be >= this to be in the gap
	tankGapRightEdge = 0x90 // X must be <= this to be in the gap
	tankGapProbe     = 10   // added to X before comparing with left edge
	bridgeEarlyLevel = 7    // bridge index threshold: <= this → remove tank; > this → bank-tank
)

// bridgeTankResult is the outcome of the per-frame frozen-tank gap check.
type bridgeTankResult struct {
	removeIndices      []int
	explosionFragments []state.ExplosionFragment
	pointsScored       int
}

// applyBridgeDestroyedTanks runs the frozen road-tank gap check every frame while the
// bridge is destroyed. For each road tank:
//   - If in the river gap (X+10 >= $70 and X <= $90): destroy it, award 250 pts, spawn 1 fragment.
//   - Otherwise: convert to bank-tank (bridge > 7) or remove (bridge <= 7).
func applyBridgeDestroyedTanks(vp *state.Viewport, bridgeIndex int) bridgeTankResult {
	var result bridgeTankResult

	for i, obj := range vp.Objects {
		if obj.Type != domain.ObjectTank || obj.TankLocation != domain.TankLocationRoad {
			continue
		}

		if obj.X+tankGapProbe >= tankGapLeftEdge && obj.X <= tankGapRightEdge {
			// Tank is over the river gap: destroy it.
			result.removeIndices = append(result.removeIndices, i)
			result.explosionFragments = append(result.explosionFragments, state.ExplosionFragment{
				X: obj.X, Y: obj.Y,
			})
			result.pointsScored += PointsTank
		} else {
			// Tank is on the bank.
			if bridgeIndex > bridgeEarlyLevel {
				// Convert to a stationary bank-tank.
				obj.TankLocation = domain.TankLocationBank
			} else {
				// Early level: just remove it.
				result.removeIndices = append(result.removeIndices, i)
			}
		}
	}

	return result
}
